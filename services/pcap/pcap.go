package pcap

import (
	"context"
	"encoding/base64"
	"github.com/google/gopacket"
	"github.com/it1804/kafka-bridge/config"
	"github.com/it1804/kafka-bridge/input"
	"github.com/it1804/kafka-bridge/input/handlers"
	"github.com/it1804/kafka-bridge/output"
	"github.com/it1804/kafka-bridge/stat"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"net"
	_ "strconv"
	"sync"
)

type (
	pcapService struct {
		conf   *config.ServiceConf
		ctx    context.Context
		wg     *sync.WaitGroup
		output *output.KafkaWriter
		input  *input.PcapSource
	}
)

func NewPcapService(ctx context.Context, wg *sync.WaitGroup, conf *config.ServiceConf) *pcapService {

	config.ValidatePcapServerConfig(&conf.PcapService, conf.Name)

	config.ValidateKafkaProducerConfig(&conf.KafkaProducer, conf.Name)

	s := &pcapService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: input.NewPcapSource(conf.Name, &input.PcapSourceConf{
			Device:     conf.PcapService.Device,
			SnapLen:    conf.PcapService.SnapLen,
			Expression: conf.PcapService.Expression,
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

func (s *pcapService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()
	go func() {
		handler, _ := handlers.NewPcapPacketHandler(s.handle)
		err := s.input.Run(s.ctx, handler)
		if err != nil {
			log.Fatalf("[%s] Pcap error: %s", s.conf.Name, err)
		}
	}()
	select {
	case <-s.ctx.Done():
		s.input.Shutdown()
		s.output.Shutdown()
	}
	return nil
}

func (s *pcapService) GetStat() *stat.ServiceStat {
	return s.output.GetStat()
}

func (s *pcapService) handle(packet gopacket.Packet) error {

	src := (net.IP)(packet.NetworkLayer().NetworkFlow().Src().Raw())
	payload := packet.TransportLayer().LayerPayload()

	if len(payload) > 0 {
		var headers []kafka.Header
		if len(s.conf.PcapService.SrcIpHeaderName) > 0 {
			hdr := kafka.Header{s.conf.PcapService.SrcIpHeaderName, []byte(src.String())}
			headers = append(headers, hdr)
		}

		var key []byte
		if s.conf.PcapService.SrcIpAsPartitionKey {
			key = []byte(src.String())
		}

		if s.conf.PcapService.Base64Body {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString([]byte(payload))), Headers: headers, Key: key}
		} else {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: payload, Headers: headers, Key: key}
		}

	} else {
		log.Printf("[%s] Pcap received empty packet\n", s.conf.Name)
	}
	return nil
}
