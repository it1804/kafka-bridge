package stat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/it1804/kafka-bridge/common/config"
	"github.com/it1804/kafka-bridge/common/handlers"
	"github.com/it1804/kafka-bridge/common/stat"
	"github.com/it1804/kafka-bridge/services"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"sync"
)

type statService struct {
	conf  *config.StatServiceConf
	ctx   context.Context
	wg    *sync.WaitGroup
	input *stat.HttpServer
	watch []services.Service
}

func NewStatService(ctx context.Context, wg *sync.WaitGroup, conf *config.StatServiceConf) *statService {

	config.ValidateStatServerConfig(conf, "stat")

	s := &statService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: stat.NewHttpServer("stat", &stat.HttpServerConf{
			Listen: conf.Listen,
		}),
	}
	go s.run()
	return s
}

func (s *statService) Watch(service services.Service) {
	s.watch = append(s.watch, service)
	return
}

func (s *statService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()

	collector := NewCollector(&s.watch)
	prometheus.MustRegister(collector)

	go func() {
		handler, _ := handlers.NewHttpPacketHandler(s.handle)
		err := s.input.Run(s.ctx, handler)
		if err != nil {
			log.Fatalf("[%s] HTTP error: %s", "stat", err)
		}
	}()

	select {
	case <-s.ctx.Done():
		s.input.Shutdown()
	}
	return nil
}

func (s *statService) handle(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "405 method not allowed\n")
		return nil
	}
	switch r.URL.Path {
	case s.conf.JsonPath:
		var stat stat.ServiceStatList
		for service := range s.watch {
			stat.Services = append(stat.Services, *s.watch[service].GetStat())
		}
		b, _ := json.MarshalIndent(stat, "", "  ")
		fmt.Fprintf(w, string(b))
	case s.conf.MetricsPath:
		promhttp.Handler().ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
	return nil

}
