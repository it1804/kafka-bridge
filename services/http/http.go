package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/it1804/kafka-bridge/config"
	"github.com/it1804/kafka-bridge/input"
	"github.com/it1804/kafka-bridge/input/handlers"
	"github.com/it1804/kafka-bridge/output"
	"github.com/it1804/kafka-bridge/stat"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type httpService struct {
	conf   *config.ServiceConf
	output *output.KafkaWriter
	input  *input.HttpServer
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewHttpService(ctx context.Context, wg *sync.WaitGroup, conf *config.ServiceConf) *httpService {

	config.ValidateHttpServerConfig(&conf.HttpService, conf.Name)
	config.ValidateKafkaProducerConfig(&conf.KafkaProducer, conf.Name)

	s := &httpService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: input.NewHttpServer(conf.Name, &input.HttpServerConf{
			Listen: conf.HttpService.Listen,
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

func (s *httpService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()
	go func() {
		handler, _ := handlers.NewHttpPacketHandler(s.handle)
		err := s.input.Run(s.ctx, handler)
		if err != nil {
			log.Fatalf("[%s] HTTP error: %s", s.conf.Name, err)
		}
	}()

	select {
	case <-s.ctx.Done():
		s.input.Shutdown()
		s.output.Shutdown()
	}
	return nil
}

func (s *httpService) GetStat() *stat.ServiceStat {
	return s.output.GetStat()
}

func (s *httpService) handle(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 method not allowed\n")
		return nil
	}

	if r.URL.Path != s.conf.HttpService.Path {
		http.NotFound(w, r)
		return nil
	}

	var headers []kafka.Header
	for _, header := range s.conf.HttpService.AllowedHeaders {
		value := r.Header.Get(header)
		if len(value) > 0 {
			hdr := kafka.Header{header, []byte(value)}
			headers = append(headers, hdr)
		}
	}

	switch s.conf.HttpService.HttpMode {
	case "json":
		bodyJson, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		msg := make(map[string]interface{})
		headersJson := make(map[string]string)
		for _, header := range s.conf.HttpService.AllowedHeaders {
			value := r.Header.Get(header)
			if len(value) > 0 {
				headersJson[header] = value
			}
		}
		msg["headers"] = headersJson
		msg["body"] = string(bodyJson)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: jsonMsg, Headers: headers}

	case "base64_body":
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString(body)), Headers: headers}

	case "body":
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: body, Headers: headers}

	case "raw":
		w := &bytes.Buffer{}
		err := r.WriteProxy(w)
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: w.Bytes(), Headers: headers}
	}

	return nil
}
