package utils

import (
    "fmt"
    "os"
    "encoding/json"
    "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type (
    KafkaWriter struct {
	brokers string
	topic string
    }
)

func NewKafkaWriter(brokers string, topic string) *KafkaWriter {
    return &KafkaWriter{brokers: brokers, topic: topic}
}

func (r *KafkaWriter) MessageHandler(input chan []byte) (err error) {
    go func() {
        p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": r.brokers,"statistics.interval.ms": 5000})
	if err != nil {
    	    panic(err)
	}
	defer p.Close()

	go func() {
    	    for e := range p.Events() {
		switch ev := e.(type) {
		    case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
		    case *kafka.Stats:
			var stats map[string]interface{}
			json.Unmarshal([]byte(e.String()), &stats)
			fmt.Printf("Stats: %v messages (%v bytes) messages consumed\n",
				stats["rxmsgs"], stats["rxmsg_bytes"])
		    case *kafka.Message:
			if ev.TopicPartition.Error != nil {
			    fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
			} 
		    default:
			fmt.Printf("Ignored %v\n", e)
		    }
		}
	}()

        for {
	    msg := <- input
//          fmt.Println(msg)
	    p.Produce(&kafka.Message { TopicPartition: kafka.TopicPartition{Topic: &r.topic, Partition: kafka.PartitionAny}, Value: msg, }, nil)
        }
	p.Flush(15 * 1000)
    }()
    return;
}
