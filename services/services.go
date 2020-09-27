package services

import (
	"github.com/it1804/kafka-bridge/common/stat"
)

type (
	Service interface {
		GetStat() *stat.ServiceStat
	}

	StatService interface {
		Watch(service Service)
	}
)
