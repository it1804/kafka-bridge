package handlers

import "net/http"

type (
	HttpPacketHandler interface {
		Handle(w http.ResponseWriter, r *http.Request) error
	}

	pHttpPacketHandler func(w http.ResponseWriter, r *http.Request) error

	httpPacketHandler struct {
		handler pHttpPacketHandler
	}
)

func (ph *httpPacketHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	return ph.handler(w, r)
}

func NewHttpPacketHandler(handler pHttpPacketHandler) (*httpPacketHandler, error) {
	return &httpPacketHandler{
		handler: handler,
	}, nil
}
