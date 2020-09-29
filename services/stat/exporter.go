package stat

import (
	"github.com/it1804/kafka-bridge/services"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	tpmMetric        *prometheus.Desc
	queueMetric      *prometheus.Desc
	queueBytesMetric *prometheus.Desc
	errMetric        *prometheus.Desc
	txmsgsMetric     *prometheus.Desc
	txmsgBytesMetric *prometheus.Desc
	watch            *[]services.Service
}

func NewCollector(watch *[]services.Service) *collector {
	return &collector{
		watch: watch,
		tpmMetric: prometheus.NewDesc("kafka_bridge_total_processed_messages",
			"Shows total processed messages",
			[]string{"service_name"}, nil,
		),
		queueMetric: prometheus.NewDesc("kafka_bridge_messages_count_in_queue",
			"Shows messages count in queue",
			[]string{"service_name"}, nil,
		),
		queueBytesMetric: prometheus.NewDesc("kafka_bridge_messages_bytes_in_queue",
			"Shows messages bytes in queue",
			[]string{"service_name"}, nil,
		),
		errMetric: prometheus.NewDesc("kafka_bridge_delivered_errors_count",
			"Shows delivered errors count",
			[]string{"service_name"}, nil,
		),
		txmsgsMetric: prometheus.NewDesc("kafka_bridge_delivered_messages_count",
			"Shows delivered messages count",
			[]string{"service_name"}, nil,
		),
		txmsgBytesMetric: prometheus.NewDesc("kafka_bridge_delivered_messages_bytes",
			"Shows delivered bytes with messages",
			[]string{"service_name"}, nil,
		),
	}
}

func (collector *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.tpmMetric
	ch <- collector.queueBytesMetric
	ch <- collector.queueMetric
	ch <- collector.errMetric
	ch <- collector.txmsgsMetric
	ch <- collector.txmsgBytesMetric
}

func (collector *collector) Collect(ch chan<- prometheus.Metric) {
	for service := range *collector.watch {
		stat := (*collector.watch)[service].GetStat()
		name := stat.Name
		total_cnt := stat.Output.Total
		msg_cnt := stat.Output.Msg_cnt
		msg_size := stat.Output.Msg_size
		errors_cnt := stat.Output.Errors
		txmsgs_cnt := stat.Output.Txmsgs
		txmsg_bytes := stat.Output.Txmsg_bytes
		ch <- prometheus.MustNewConstMetric(collector.tpmMetric, prometheus.CounterValue, float64(total_cnt), name)
		ch <- prometheus.MustNewConstMetric(collector.queueMetric, prometheus.GaugeValue, float64(msg_cnt), name)
		ch <- prometheus.MustNewConstMetric(collector.queueBytesMetric, prometheus.GaugeValue, float64(msg_size), name)
		ch <- prometheus.MustNewConstMetric(collector.errMetric, prometheus.CounterValue, float64(errors_cnt), name)
		ch <- prometheus.MustNewConstMetric(collector.txmsgsMetric, prometheus.CounterValue, float64(txmsgs_cnt), name)
		ch <- prometheus.MustNewConstMetric(collector.txmsgBytesMetric, prometheus.CounterValue, float64(txmsg_bytes), name)
	}
}
