package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/it1804/kafka-bridge/common/handlers"
	"github.com/it1804/kafka-bridge/common/input"
	"github.com/it1804/kafka-bridge/common/stat"
	"github.com/it1804/kafka-bridge/config"
	"github.com/valyala/fasthttp"
	"log"
	"sync"
)

type statService struct {
	conf  *config.StatService
	ctx   context.Context
	wg    *sync.WaitGroup
	input *input.HttpServer
	watch []Service
}

func NewStatService(ctx context.Context, wg *sync.WaitGroup, conf *config.StatService) *statService {
	s := &statService{
		conf: conf,
		ctx:  ctx,
		wg:   wg,
		input: input.NewHttpServer("stat", &input.HttpServerConf{
			Listen: conf.Listen,
		}),
	}
	go s.run()
	return s
}

func (s *statService) Watch(service Service) {
	s.watch = append(s.watch, service)
	return
}

func (s *statService) run() (err error) {
	s.wg.Add(1)
	defer s.wg.Done()

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

func (s *statService) handle(ctx *fasthttp.RequestCtx) (err error) {
	var stat stat.ServiceStatList
	ctx.SetContentType("text/plain; charset=utf8")
	for w := range s.watch {
		stat.Services = append(stat.Services, *s.watch[w].GetStat())
	}
	b, _ := json.MarshalIndent(stat, "", "  ")
	fmt.Fprintf(ctx, string(b))
	return nil
}