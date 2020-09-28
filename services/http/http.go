package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/it1804/kafka-bridge/common/handlers"
	"github.com/it1804/kafka-bridge/common/input"
	"github.com/it1804/kafka-bridge/common/output"
	"github.com/it1804/kafka-bridge/common/stat"
	"github.com/it1804/kafka-bridge/config"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type httpService struct {
	conf   *config.Service
	output *output.KafkaWriter
	input  *input.HttpServer
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewHttpService(ctx context.Context, wg *sync.WaitGroup, conf *config.Service) *httpService {

    config.ValidateHttpServerConfig(&conf.HttpService, conf.Name)
	config.ValidateKafkaProducerConfig(&conf.KafkaProducer, conf.Name)

	s := &httpService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: input.NewHttpServer(conf.Name, &input.HttpServerConf{
			Listen: conf.HttpService.Listen,
			Path:   "/",
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 method not allowed\n")
		return nil
	}

	if r.URL.Path != s.conf.HttpService.Path {
		http.NotFound(w, r)
		return nil
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
		return nil
	}

	if s.conf.HttpService.UseJSON {
		msg := make(map[string]interface{})
		headers := make(map[string]string)
		for _, header := range s.conf.HttpService.AllowedHeaders {
			value := r.Header.Get(header)
			if len(value) > 0 {
				headers[header] = value
			}
		}
		msg["headers"] = headers
		if s.conf.HttpService.Base64Body {
			msg["body"] = body
		} else {
			msg["body"] = string(body)
		}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[%s] HTTP error: %v", s.conf.Name, err.Error())
			return nil
		}
		s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: jsonMsg}
	} else {
		var headers []kafka.Header
		for _, header := range s.conf.HttpService.AllowedHeaders {
			value := r.Header.Get(header)
			if len(value) > 0 {
				hdr := kafka.Header{header, []byte(value)}
				headers = append(headers, hdr)
			}
		}
		if s.conf.HttpService.Base64Body {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString(body)), Headers: headers}
		} else {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: body, Headers: headers}
		}
	}
	return nil
}
