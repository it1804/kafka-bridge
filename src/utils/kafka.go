package utils

import (
    "sync"
    "log"
    "encoding/json"
    "gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type (
    KafkaWriter struct {
	brokers string
	topic string
	workers int
    }
)


func NewKafkaWriter(brokers string, topic string,workers int) *KafkaWriter {
    return &KafkaWriter{brokers: brokers, topic: topic, workers: workers}
}


func (r *KafkaWriter) MessageHandler(input chan []byte) (err error) {
    go func() {  
	wg := new(sync.WaitGroup)
	 for i := 0; i < r.workers; i++ {
    	    wg.Add(1)
    	    worker(input, wg, r)
	}
	wg.Wait()
    }()
    return
}

func worker(input chan []byte, wg *sync.WaitGroup, r *KafkaWriter) {
    go func() {  
        p, err := kafka.NewProducer(&kafka.ConfigMap {
					    "bootstrap.servers": r.brokers,
					    "statistics.interval.ms": 30000,
					    "compression.codec": "snappy",
					    "socket.keepalive.enable": true,
					    "socket.timeout.ms": 1000,
					    "enable.idempotence": true,
//					    "debug": "all",
					    "retries": "100000",
					    "batch.num.messages": 100000,
					    "queue.buffering.max.messages":1000000,
					})
	if err != nil {
    	    panic(err)
	}
	defer p.Close()

	go func() {
    	    for e := range p.Events() {
		switch ev := e.(type) {
		    case kafka.Error:
			log.Printf("Kafka gid=%d: Error: %v\n",getGID(), e)
		    case *kafka.Stats:
			var stats map[string]interface{}
			json.Unmarshal([]byte(e.String()), &stats)
			log.Printf("Kafka stats gid=%d: %v messages (%v bytes) written\n",
				 getGID(),stats["txmsgs"],stats["txmsg_bytes"])
		    case *kafka.Message:
			if ev.TopicPartition.Error != nil {
			    log.Printf("Kafka gid=%d: Delivery failed: %v\n",getGID(), ev.TopicPartition)
			} 
		    default:
			log.Printf("Kafka gid=%d: Ignored event %v\n",getGID(), e)
		    }
		}
	}()

        for {
    	    msg := <- input
	    err = p.Produce(&kafka.Message { TopicPartition: kafka.TopicPartition{Topic: &r.topic, Partition: kafka.PartitionAny}, Value: msg, }, nil)
    	    if err != nil {
        	log.Printf("Kafka gid=%d: Failed to produce message: %v\n",getGID(), err)
    	    }
    	}
	p.Flush(15 * 1000)

    }()
    return;
}
