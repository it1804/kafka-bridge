package input

import (
	"context"
	"github.com/it1804/kafka-bridge/input/handlers"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"sync"
)

type (
	KafkaReaderConf struct {
		Brokers string
		Topic   string
		Options []string
		Group   string
	}

	KafkaReader struct {
		conf        *KafkaReaderConf
		wg          *sync.WaitGroup
		serviceName string
	}

	kafkaMessageHandler struct {
		phandler handlers.KafkaMessageHandler
	}
)

func NewKafkaReader(serviceName string, conf *KafkaReaderConf) *KafkaReader {
	return &KafkaReader{
		conf:        conf,
		wg:          &sync.WaitGroup{},
		serviceName: serviceName,
	}
}

func (r *KafkaReader) Run(ctx context.Context, phandler handlers.KafkaMessageHandler) (err error) {
	handler := &kafkaMessageHandler{
		phandler: phandler,
	}
	r.wg.Add(1)

	defer r.wg.Done()

	errs := make(chan error, 1)

	go func() {
		kafkaConf := &kafka.ConfigMap{}
		for _, opt := range r.conf.Options {
			kafkaConf.Set(opt)
		}
		kafkaConf.Set("bootstrap.servers=" + r.conf.Brokers)
		kafkaConf.Set("group.id=" + r.conf.Group)
		kafkaConf.SetKey("go.events.channel.enable", true)

		c, err := kafka.NewConsumer(kafkaConf)
		if err != nil {
			errs <- err
			return
		}

		defer c.Close()

		err = c.Subscribe(r.conf.Topic, nil)

		if err != nil {
			errs <- err
			return
		}
		for {
			select {
			case <-ctx.Done():
				log.Printf("[%s] Consumer shutdown\n", r.serviceName)
				return
			case ev := <-c.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					c.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					c.Unassign()
				case *kafka.Message:
					handler.handleMessage(e.Value, e.Headers)
				case kafka.Error:
					log.Printf("[%s] Consumer error: %v\n", r.serviceName, e)
				}
			}
		}
	}()

	select {
	case err := <-errs:
		close(errs)
		return err
	case <-ctx.Done():
		return nil
	}
	return nil
}

func (r *KafkaReader) Shutdown() {
	r.wg.Wait()
	log.Printf("[%s] Kafka consumer stopped", r.serviceName)
	return
}

func (h *kafkaMessageHandler) handleMessage(message []byte, headers []kafka.Header) {
	h.phandler.Handle(message, headers)
}
