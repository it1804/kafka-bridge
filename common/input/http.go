package input

import (
	"context"
	"github.com/it1804/kafka-bridge/common/handlers"
	"github.com/valyala/fasthttp"
	"log"
	"sync"
	"time"
)

type (
	HttpServerConf struct {
		Listen string
	}

	HttpServer struct {
		conf        *HttpServerConf
		wg          *sync.WaitGroup
		serviceName string
	}

	httpRequestHandler struct {
		phandler handlers.HttpPacketHandler
	}
)

func NewHttpServer(serviceName string, conf *HttpServerConf) *HttpServer {
	return &HttpServer{
		conf:        conf,
		wg:          &sync.WaitGroup{},
		serviceName: serviceName,
	}
}

func (r *HttpServer) Run(ctx context.Context, phandler handlers.HttpPacketHandler) (err error) {
	r.wg.Add(1)
	defer r.wg.Done()

	requestHandler := &httpRequestHandler{
		phandler: phandler,
	}

	server := &fasthttp.Server{
		Handler:          requestHandler.handleRequest,
		ReadTimeout:      300 * time.Second,
		WriteTimeout:     300 * time.Second,
		DisableKeepalive: false,
		IdleTimeout:      5 * time.Second,
	}
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe(r.conf.Listen)
		if err != nil {
			errs <- err
			return
		}
	}()

	select {
	case err := <-errs:
		close(errs)
		return err
	case <-ctx.Done():
		if err := server.Shutdown(); err != nil {
			log.Printf("[%s] HTTP error with graceful close: %s", r.serviceName, err)
			return err
		}
		return nil
	}
	return nil
}

func (r *HttpServer) Shutdown() {
	r.wg.Wait()
	log.Printf("[%s] HTTP server stopped", r.serviceName)
	return
}

func (h *httpRequestHandler) handleRequest(ctx *fasthttp.RequestCtx) {
	h.phandler.Handle(ctx)
	return
}
