package config

import (
	"log"
)

type (
	KafkaProducerConf struct {
		Brokers        string            `json:"brokers"`
		Topic          string            `json:"topic"`
		Options        []string          `json:"options,omitempty"`
		SetHeaders     map[string]string `json:"set_headers,omitempty"`
		QueueBufferLen uint              `json:"queue_buffer_len"`
	}

	KafkaConsumerServiceConf struct {
		Brokers    string   `json:"brokers"`
		Topic      string   `json:"topic"`
		Options    []string `json:"options,omitempty"`
		Group      string   `json:"group"`
		Base64Body bool     `json:"base64_encode_body"`
	}

	HttpServiceConf struct {
		Listen         string   `json:"listen"`
		Path           string   `json:"path"`
		Base64Body     bool     `json:"base64_encode_body"`
		UseJSON        bool     `json:"use_json,omitempty"`
		AllowedHeaders []string `json:"allowed_headers,omitempty"`
	}

	UdpServiceConf struct {
		Listen         string            `json:"listen"`
		Base64Body     bool              `json:"base64_encode_body"`
		SignatureBytes map[string]string `json:"signature_bytes,omitempty"`
		Workers        uint              `json:"workers,omitempty"`
		MaxPacketSize  uint              `json:"max_packet_size,omitempty"`
	}

	Service struct {
		Type string `json:"type"`
		Name string `json:"name"`

		HttpService          HttpServiceConf          `json:"http_options"`
		UdpService           UdpServiceConf           `json:"udp_options"`
		KafkaConsumerService KafkaConsumerServiceConf `json:"kafka_consumer_options"`

		KafkaProducer KafkaProducerConf `json:"kafka_producer"`
	}

	StatService struct {
		Listen   string `json:"listen,omitempty"`
		JsonPath string `json:"json_path,omitempty"`
	}

	ServiceList struct {
		Services []Service   `json:"services"`
		Stat     StatService `json:"stat"`
	}
)

func ValidateKafkaProducerConfig(conf *KafkaProducerConf, serviceName string) {
	if len(conf.Brokers) == 0 {
		log.Fatalf("[%s] You must specify one or more producer broker address", serviceName)
	}
	if len(conf.Topic) == 0 {
		log.Fatalf("[%s] You must specify kafka producer topic name", serviceName)
	}
	if conf.QueueBufferLen == 0 {
		log.Printf("[%s] queue_buffer_len value not defined, set to default (1000)", serviceName)
		conf.QueueBufferLen = 1000
	}
}

func ValidateKafkaConsumerServiceConfig(conf *KafkaConsumerServiceConf, serviceName string) {

	if len(conf.Brokers) == 0 {
		log.Fatalf("[%s] You must specify one or more kafka consumer broker address", serviceName)
	}
	if len(conf.Topic) == 0 {
		log.Fatalf("[%s] You must specify kafka consumer topic name", serviceName)
	}
	if len(conf.Group) == 0 {
		log.Fatalf("[%s] You must specify kafka consumer group id", serviceName)
	}
}

func ValidateUdpServerConfig(conf *UdpServiceConf, serviceName string) {

	if conf.Workers < 1 {
		log.Printf("[%s] udp workers value not defined, set to default (1)", serviceName)
		conf.Workers = 1
	}
	if conf.MaxPacketSize == 0 {
		log.Printf("[%s] udp max_packet_size value not defined, set to default (65535)", serviceName)
		conf.MaxPacketSize = 65535
	}
}

func ValidateHttpServerConfig(conf *HttpServiceConf, serviceName string) {
	if len(conf.Path) == 0 {
		log.Printf("[%s] http path value not defined, set to default (\"/\")", serviceName)
        conf.Path="/"
	}
}

