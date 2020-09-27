package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/it1804/kafka-bridge/common/handlers"
	"github.com/it1804/kafka-bridge/common/input"
	"github.com/it1804/kafka-bridge/common/output"
	"github.com/it1804/kafka-bridge/common/stat"
	"github.com/it1804/kafka-bridge/config"
	"github.com/valyala/fasthttp"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
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

func (s *httpService) handle(ctx *fasthttp.RequestCtx) error {
	if string(ctx.Method()) != "POST" {
		ctx.Error("Not Found", 404)
		return nil
	}
	ctx.SetContentType("text/plain; charset=utf8")

	if s.conf.HttpService.UseJSON {
		msg := make(map[string]interface{})
		headers := make(map[string]string)
		for _, header := range s.conf.HttpService.AllowedHeaders {
			value := string(ctx.Request.Header.Peek(header))
			if len(value) > 0 {
				headers[header] = value
			}
		}
		msg["headers"] = headers
		if s.conf.HttpService.Base64Body {
			msg["body"] = ctx.PostBody()
		} else {
			msg["body"] = string(ctx.PostBody())
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
			value := string(ctx.Request.Header.Peek(header))
			if len(value) > 0 {
				hdr := kafka.Header{header, []byte(value)}
				headers = append(headers, hdr)
			}
		}
		if s.conf.HttpService.Base64Body {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: []byte(base64.StdEncoding.EncodeToString(ctx.PostBody())), Headers: headers}
		} else {
			s.output.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: s.output.GetTopic(), Partition: kafka.PartitionAny}, Value: ctx.PostBody(), Headers: headers}
		}
	}
	return nil
}
