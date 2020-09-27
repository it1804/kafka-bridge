package stat

type (
	KafkaStat struct {
		Name         string `json:"producer_name"`
		Total        uint64 `json:"total_processed_messages"`
		Ops          uint64 `json:"current_ops"`
		Errors       uint64 `json:"delivered_errors_count"`
		Txmsgs       uint64 `json:"delivered_messages_count"`
		Txmsg_bytes  uint64 `json:"delivered_messages_bytes"`
		Msg_cnt      uint64 `json:"messages_count_in_queue"`
		Msg_size     uint64 `json:"messages_bytes_in_queue"`
		Msg_max      uint64 `json:"max_queue_messages_count"`
		Msg_size_max uint64 `json:"max_queue_messages_bytes"`
	}

	StatisticsCounter struct {
		serviceName string
		output      KafkaStat
	}

	ServiceStatList struct {
		Services []ServiceStat `json:"stats"`
	}

	ServiceStat struct {
		Name   string    `json:"name"`
		Output KafkaStat `json:"output"`
	}
)

func NewStatisticsCounter(serviceName string) *StatisticsCounter {
	return &StatisticsCounter{
		serviceName: serviceName,
	}
}

func (s *StatisticsCounter) Update(val KafkaStat) {
	s.output = val
	return
}

func (s *StatisticsCounter) GetStat() *ServiceStat {
	return &ServiceStat{Name: s.serviceName, Output: s.output}
}
