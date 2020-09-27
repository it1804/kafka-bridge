package handlers

import "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

type (
	KafkaMessageHandler interface {
		Handle(payload []byte, headers []kafka.Header) error
	}

	pKafkaMessageHandler func(payload []byte, headers []kafka.Header) error

	kafkaMessageHandler struct {
		handler pKafkaMessageHandler
	}
)

func (ph *kafkaMessageHandler) Handle(payload []byte, headers []kafka.Header) error {
	return ph.handler(payload, headers)
}

func NewKafkaMessageHandler(handler pKafkaMessageHandler) (*kafkaMessageHandler, error) {
	return &kafkaMessageHandler{
		handler: handler,
	}, nil
}
