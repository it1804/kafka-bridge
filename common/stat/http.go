package stat

import (
	"context"
	"github.com/it1804/kafka-bridge/common/handlers"
	"log"
	"net/http"
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

	handler := &httpRequestHandler{
		phandler: phandler,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.handleRequest(w, r)
		},
	))

	server := &http.Server{
		Addr:         r.conf.Listen,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	errs := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errs <- err
			return
		}
	}()

	select {
	case err := <-errs:
		close(errs)
		return err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return nil
	}
	return nil
}

func (r *HttpServer) Shutdown() {
	r.wg.Wait()
	log.Printf("[%s] HTTP server stopped", r.serviceName)
	return
}

func (h *httpRequestHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	h.phandler.Handle(w, r)

	return
}
