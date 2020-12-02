package kafka

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
	"sync"
)

type (
	kafkaService struct {
		conf   *config.ServiceConf
		input  *input.KafkaReader
		output *output.KafkaWriter
		wg     *sync.WaitGroup
		ctx    context.Context
	}
)

func NewKafkaService(ctx context.Context, wg *sync.WaitGroup, conf *config.ServiceConf) *kafkaService {

	config.ValidateKafkaConsumerServiceConfig(&conf.KafkaConsumerService, conf.Name)
	config.ValidateKafkaProducerConfig(&conf.KafkaProducer, conf.Name)

	s := &kafkaService{
		ctx:  ctx,
		conf: conf,
		wg:   wg,
		input: input.NewKafkaReader(conf.Name, &input.KafkaReaderConf{
			Brokers: conf.KafkaConsumerService.Brokers,
			Topic:   conf.KafkaConsumerService.Topic,
			Group:   conf.KafkaConsumerService.Group,
			Options: conf.KafkaConsumerService.Options,
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

func (s *kafkaService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()
	go func() {
		handler, _ := handlers.NewKafkaMessageHandler(s.handle)
		err := s.input.Run(s.ctx, handler)
		if err != nil {
			log.Fatalf("[%s] Kafka consumer error: %s", s.conf.Name, err)
		}
	}()

	select {
	case <-s.ctx.Done():
		s.input.Shutdown()
		s.output.Shutdown()
	}
	return nil
}

func (s *kafkaService) GetStat() *stat.ServiceStat {
	return s.output.GetStat()
}

func (s *kafkaService) handle(payload []byte, headers []kafka.Header) error {
	if s.conf.KafkaConsumerService.Base64Body {
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString([]byte(payload))), Headers: headers}
	} else {
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: payload, Headers: headers}
	}
	return nil
}
