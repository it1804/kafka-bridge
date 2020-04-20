package utils

type Configuration struct {
    Listen struct {
	Address string
	MaxPacketSize int
	Workers int
    }
    Kafka struct {
	Brokers  string
	Topic   string
	Workers int
    }
}
