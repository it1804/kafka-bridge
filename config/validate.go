package config

import (
	"log"
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
		conf.QueueBufferLen = 10000
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
	if len(conf.Listen) == 0 {
		log.Fatalf("[%s] You must specify listen address", serviceName)
	}
	if conf.Workers < 1 {
		log.Printf("[%s] udp workers value not defined, set to default (1)", serviceName)
		conf.Workers = 1
	}
	if conf.MaxPacketSize == 0 {
		log.Printf("[%s] udp max_packet_size value not defined, set to default (65535)", serviceName)
		conf.MaxPacketSize = 65535
	}
}

func ValidatePcapServerConfig(conf *PcapServiceConf, serviceName string) {
	if len(conf.Device) == 0 {
		log.Fatalf("[%s] You must specify device name", serviceName)
	}
	if conf.SnapLen == 0 {
		log.Printf("[%s] Pcap snap len value not defined, set to default (65535)", serviceName)
		conf.SnapLen = 65535
	}
	if len(conf.Expression) == 0 {
		log.Fatalf("[%s] You must specify pcap filter", serviceName)
	}

}

func ValidateHttpServerConfig(conf *HttpServiceConf, serviceName string) {
	if len(conf.Listen) == 0 {
		log.Fatalf("[%s] You must specify listen address", serviceName)
	}
	if len(conf.Path) == 0 {
		log.Printf("[%s] http path value not defined, set to default (\"/\")", serviceName)
		conf.Path = "/"
	}
}

func ValidateStatServerConfig(conf *StatServiceConf, serviceName string) {
	if len(conf.Listen) == 0 {
		log.Fatalf("[%s] You must specify listen address", serviceName)
	}
	if len(conf.MetricsPath) == 0 {
		log.Printf("[%s] stat metrics_path value not defined, set to default (\"/metrics\")", serviceName)
		conf.MetricsPath = "/metrics"
	}
	if len(conf.JsonPath) == 0 {
		log.Printf("[%s] stat json_path value not defined, set to default (\"/stat\")", serviceName)
		conf.JsonPath = "/stat"
	}
}
