package udp

import (
	"context"
	"encoding/base64"
	"github.com/it1804/kafka-bridge/config"
	"github.com/it1804/kafka-bridge/input"
	"github.com/it1804/kafka-bridge/input/handlers"
	"github.com/it1804/kafka-bridge/output"
	"github.com/it1804/kafka-bridge/stat"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"net"
	"strconv"
	"sync"
)

type (
	udpService struct {
		conf   *config.ServiceConf
		ctx    context.Context
		wg     *sync.WaitGroup
		output *output.KafkaWriter
		input  *input.UdpServer
	}
)

func NewUdpService(ctx context.Context, wg *sync.WaitGroup, conf *config.ServiceConf) *udpService {

	config.ValidateUdpServerConfig(&conf.UdpService, conf.Name)

	config.ValidateKafkaProducerConfig(&conf.KafkaProducer, conf.Name)

	s := &udpService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: input.NewUdpServer(conf.Name, &input.UdpServerConf{
			Listen:        conf.UdpService.Listen,
			MaxPacketSize: conf.UdpService.MaxPacketSize,
			Workers:       conf.UdpService.Workers,
		}),
		output: output.NewKafkaWriter(ctx, conf.Name, &output.KafkaWriterConf{
			Brokers:        conf.KafkaProducer.Brokers,
			Topic:          conf.KafkaProducer.Topic,
			Options:        conf.KafkaProducer.Options,
			QueueBufferLen: conf.KafkaProducer.QueueBufferLen,
			SetHeaders:     conf.KafkaProducer.SetHeaders,
		}),
	}
	go s.run()
	return s
}

func (s *udpService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()
	go func() {
		handler, _ := handlers.NewUdpPacketHandler(s.handle)
		err := s.input.Run(s.ctx, handler)
		if err != nil {
			log.Fatalf("[%s] UDP error: %s", s.conf.Name, err)
		}
	}()
	select {
	case <-s.ctx.Done():
		s.input.Shutdown()
		s.output.Shutdown()
	}
	return nil
}

func (s *udpService) GetStat() *stat.ServiceStat {
	return s.output.GetStat()
}

func (s *udpService) handle(payload []byte, length int, src net.IP) error {
	if length > 0 {
		for key, value := range s.conf.UdpService.SignatureBytes {
			offset, err := strconv.ParseUint(key, 10, 32)
			if err != nil {
				log.Printf("[%s] UDP invalid signature offset: %s\n", s.conf.Name, key)
				return nil
			}
			code, err := strconv.ParseUint(value, 0, 8)
			if err != nil {
				log.Printf("[%s] UDP invalid signature value: %v\n", s.conf.Name, value)
				return nil
			}
			symbol := byte(code)
			if offset >= uint64(length) {
				log.Printf("[%s] UDP invalid offset: %d with message len %d\n", s.conf.Name, offset, length)
				return nil
			} else if payload[offset] != symbol {
				log.Printf("[%s] UDP expected byte [0x%02x], got [0x%02x] at offset %d", s.conf.Name, symbol, payload[offset], offset)
				return nil
			}
		}
		var headers []kafka.Header
		if len(s.conf.UdpService.SrcIpHeaderName) > 0 {
			hdr := kafka.Header{s.conf.UdpService.SrcIpHeaderName, []byte(src.String())}
			headers = append(headers, hdr)
		}
		if s.conf.UdpService.Base64Body {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString([]byte(payload))), Headers: headers}
		} else {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: payload, Headers: headers}
		}
	} else {
		log.Printf("[%s] UDP received empty packet\n", s.conf.Name)
	}
	return nil
}
