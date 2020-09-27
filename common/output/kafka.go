package output

import (
	"context"
	"encoding/json"
	"github.com/it1804/kafka-bridge/common/stat"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"sync"
	"sync/atomic"
)

type (
	KafkaWriterConf struct {
		Brokers        string
		Topic          string
		Options        []string
		SetHeaders     map[string]string
		QueueBufferLen uint
	}

	KafkaWriter struct {
		conf           *KafkaWriterConf
		wg             *sync.WaitGroup
		stat           *stat.StatisticsCounter
		serviceName    string
		produceChannel chan *kafka.Message
		ctx            context.Context
	}
)

func NewKafkaWriter(ctx context.Context, serviceName string, conf *KafkaWriterConf) *KafkaWriter {
	writer := &KafkaWriter{
		conf:           conf,
		wg:             &sync.WaitGroup{},
		stat:           stat.NewStatisticsCounter(serviceName),
		serviceName:    serviceName,
		produceChannel: make(chan *kafka.Message, conf.QueueBufferLen),
		ctx:            ctx,
	}
	go func() {
		err := writer.run()
		if err != nil {
			log.Fatalf("[%s] Kafka producer error: %v", serviceName, err)
		}
	}()

	return writer
}

func (w *KafkaWriter) ProduceChannel() chan *kafka.Message {
	return w.produceChannel
}

func (w *KafkaWriter) GetStat() *stat.ServiceStat {
	return w.stat.GetStat()
}

func (w *KafkaWriter) GetTopic() *string {
	return &w.conf.Topic
}

func (w *KafkaWriter) run() (err error) {
	w.wg.Add(1)
	defer w.wg.Done()

	kafkaConf := &kafka.ConfigMap{}
	for _, opt := range w.conf.Options {
		kafkaConf.Set(opt)
	}
	kafkaConf.Set("statistics.interval.ms=1000")
	kafkaConf.Set("bootstrap.servers=" + w.conf.Brokers)

	p, err := kafka.NewProducer(kafkaConf)

	if err != nil {
		return
	}
	defer p.Close()

	go func() {
		var ops uint64 = 0
		var total uint64 = 0
		var errs uint64 = 0
		for e := range p.Events() {
			switch ev := e.(type) {
			case kafka.Error:
				log.Printf("[%s] Kafka %v: Error: %v\n", w.serviceName, p, e)
				if e.(kafka.Error).IsFatal() {
					panic(e)
				}
			case *kafka.Stats:
				atomic.AddUint64(&total, ops)
				w.updateStats(e.String(), ops, total, errs)
				atomic.StoreUint64(&ops, 0)
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("[%s] Kafka %v: Delivery failed: %v\n", w.serviceName, p, ev.TopicPartition)
					atomic.AddUint64(&errs, 1)
				}
				atomic.AddUint64(&ops, 1)
			default:
				log.Printf("[%s] Kafka %v: Ignored event %v\n", w.serviceName, p, e)
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			p.Flush(5 * 1000)
			return nil
		case msg := <-w.produceChannel:
			for key, val := range w.conf.SetHeaders {
				hdr := kafka.Header{key, []byte(val)}
				msg.Headers = append(msg.Headers, hdr)
			}
			p.ProduceChannel() <- msg
		}
	}
	return nil
}

func (w *KafkaWriter) Shutdown() {
	w.wg.Wait()
	log.Printf("[%s] Kafka producer stopped", w.serviceName)
	return
}

func (w *KafkaWriter) updateStats(msg string, ops uint64, total uint64, errs uint64) {
	var stats map[string]interface{}
	json.Unmarshal([]byte(msg), &stats)
	name := stats["name"].(string)
	txmsgs := stats["txmsgs"].(float64)
	txmsg_bytes := stats["txmsg_bytes"].(float64)
	msg_cnt := stats["msg_cnt"].(float64)
	msg_size := stats["msg_size"].(float64)
	msg_max := stats["msg_max"].(float64)
	msg_size_max := stats["msg_size_max"].(float64)
	stat := &stat.KafkaStat{
		Name:         name,
		Total:        total,
		Ops:          ops,
		Errors:       errs,
		Txmsgs:       uint64(txmsgs),
		Txmsg_bytes:  uint64(txmsg_bytes),
		Msg_cnt:      uint64(msg_cnt),
		Msg_size:     uint64(msg_size),
		Msg_max:      uint64(msg_max),
		Msg_size_max: uint64(msg_size_max),
	}
	w.stat.Update(*stat)
	return
}
