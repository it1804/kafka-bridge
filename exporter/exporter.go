package exporter

import (
	"crypto/tls"
	"encoding/json"
	"github.com/it1804/kafka-bridge/common/stat"
	"github.com/it1804/kafka-bridge/config"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	_ "log"
	"net/http"
	"time"
)

type collector struct {
	tpmMetric        *prometheus.Desc
	queueMetric      *prometheus.Desc
	queueBytesMetric *prometheus.Desc
	errMetric        *prometheus.Desc
	txmsgsMetric     *prometheus.Desc
	conf             *config.ExporterServiceList
	client           *http.Client
}

func NewCollector(config *config.ExporterServiceList) *collector {
	return &collector{
		conf: config,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
		tpmMetric: prometheus.NewDesc("kafka_bridge_total_processed_messages",
			"Shows total processed messages",
			[]string{"instance_id", "service_name"}, nil,
		),
		queueMetric: prometheus.NewDesc("kafka_bridge_messages_count_in_queue",
			"Shows messages count in queue",
			[]string{"instance_id", "service_name"}, nil,
		),
		queueBytesMetric: prometheus.NewDesc("routers_metrics_converter_messages_bytes_in_queue",
			"Shows messages bytes in queue",
			[]string{"instance_id", "service_name"}, nil,
		),
		errMetric: prometheus.NewDesc("kafka_bridge_delivered_errors_count",
			"Shows delivered errors count",
			[]string{"instance_id", "service_name"}, nil,
		),
		txmsgsMetric: prometheus.NewDesc("kafka_bridge_delivered_messages_count",
			"Shows delivered messages count",
			[]string{"instance_id", "service_name"}, nil,
		),
	}
}

func (collector *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.tpmMetric
	ch <- collector.queueMetric
	ch <- collector.queueBytesMetric
	ch <- collector.errMetric
	ch <- collector.txmsgsMetric
}

func (collector *collector) Collect(ch chan<- prometheus.Metric) {

	for i := range collector.conf.Services {
		instance := collector.conf.Services[i].Name
		stat := stat.ServiceStatList{}
		if err := collector.getServiceStatList(collector.conf.Services[i].Url, &stat); err != nil {
			return
		}
		for i := range stat.Services {
			name := stat.Services[i].Name
			total_cnt := stat.Services[i].Output.Total
			msg_cnt := stat.Services[i].Output.Msg_cnt
			msg_size := stat.Services[i].Output.Msg_size
			errors_cnt := stat.Services[i].Output.Errors
			txmsgs_cnt := stat.Services[i].Output.Txmsgs
			ch <- prometheus.MustNewConstMetric(collector.tpmMetric, prometheus.CounterValue, float64(total_cnt), instance, name)
			ch <- prometheus.MustNewConstMetric(collector.queueMetric, prometheus.GaugeValue, float64(msg_cnt), instance, name)
			ch <- prometheus.MustNewConstMetric(collector.queueBytesMetric, prometheus.GaugeValue, float64(msg_size), instance, name)
			ch <- prometheus.MustNewConstMetric(collector.errMetric, prometheus.CounterValue, float64(errors_cnt), instance, name)
			ch <- prometheus.MustNewConstMetric(collector.txmsgsMetric, prometheus.CounterValue, float64(txmsgs_cnt), instance, name)
		}
	}
}

func (collector *collector) getServiceStatList(url string, stat interface{}) error {
	r, err := collector.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return err
	}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &stat,
	})

	decoder.Decode(data)
	return nil
}
