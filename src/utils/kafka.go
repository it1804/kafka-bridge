package utils

import (
    "time"
    "fmt"
    "os"
    "log"
    "context"
    "github.com/segmentio/kafka-go"
    "github.com/segmentio/kafka-go/snappy"
)

type (
    KafkaWriter struct {
	Brokers []string
	Topic string
	QueueCapacity int
	BatchSize int
	MaxAttempts int
	RequiredAcks int
	Async bool
    }
)


func NewKafkaWriter(Brokers []string, Topic string, QueueCapacity int, BatchSize int, MaxAttempts int, RequiredAcks int, Async bool) *KafkaWriter {
    return &KafkaWriter{Brokers: Brokers, Topic: Topic, QueueCapacity: QueueCapacity,BatchSize: BatchSize, MaxAttempts: MaxAttempts, RequiredAcks: RequiredAcks, Async: Async}
}

func (r *KafkaWriter) MessageHandler(input chan []byte) (err error) {
    w := kafka.NewWriter(kafka.WriterConfig{
			Brokers: r.Brokers,
		        Topic:   r.Topic,
			QueueCapacity: r.QueueCapacity,
		    	BatchSize: r.BatchSize,
			MaxAttempts: r.MaxAttempts,
			RequiredAcks: r.RequiredAcks,
			Async: r.Async,
		        Balancer: &kafka.LeastBytes{},
			Logger: log.New(os.Stdout, "Kafka: ", 0),
			CompressionCodec: snappy.NewCompressionCodec(),
		    })

    defer w.Close()

    go func() {
	t := time.NewTicker(time.Duration(30)*time.Second)
	l := log.New(os.Stdout, "Kafka: " + " stats ", 0)

	for{
	    <- t.C
	    l.Println(fmt.Sprintf("%+v\n",w.Stats()))
	}
    }() 

    go func() {
	for {
	    msg := <- input
	    fmt.Println(msg)
	    w.WriteMessages(context.Background(), kafka.Message{Value:msg})
	}
    }()

    return;
}
