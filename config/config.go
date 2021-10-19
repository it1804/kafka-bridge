package config

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

	PcapServiceConf struct {
		Device          string `json:"device"`
		SnapLen         uint32 `json:"snap_len"`
		Expression      string `json:"filter"`
		Base64Body      bool   `json:"base64_encode_body,omitempty"`
		SrcIpHeaderName string `json:"source_ip_header,omitempty"`
	}

	HttpServiceConf struct {
		Listen         string   `json:"listen"`
		Path           string   `json:"path"`
		HttpMode       string   `json:"mode"`
		AllowedHeaders []string `json:"allowed_headers,omitempty"`
	}

	UdpServiceConf struct {
		Listen          string            `json:"listen"`
		Base64Body      bool              `json:"base64_encode_body,omitempty"`
		SignatureBytes  map[string]string `json:"signature_bytes,omitempty"`
		Workers         uint              `json:"workers,omitempty"`
		MaxPacketSize   uint32            `json:"max_packet_size,omitempty"`
		SrcIpHeaderName string            `json:"source_ip_header,omitempty"`
	}

	ServiceConf struct {
		Type string `json:"type"`
		Name string `json:"name"`

		HttpService          HttpServiceConf          `json:"http_options"`
		UdpService           UdpServiceConf           `json:"udp_options"`
		PcapService          PcapServiceConf          `json:"pcap_options"`
		KafkaConsumerService KafkaConsumerServiceConf `json:"kafka_consumer_options"`

		KafkaProducer KafkaProducerConf `json:"kafka_producer"`
	}

	StatServiceConf struct {
		Listen      string `json:"listen,omitempty"`
		JsonPath    string `json:"json_path,omitempty"`
		MetricsPath string `json:"metrics_path,omitempty"`
	}

	ServiceList struct {
		Services []ServiceConf   `json:"services"`
		Stat     StatServiceConf `json:"stat"`
	}
)
