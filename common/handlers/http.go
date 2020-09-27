package handlers

import "github.com/valyala/fasthttp"

type (
	HttpPacketHandler interface {
		Handle(ctx *fasthttp.RequestCtx) error
	}

	pHttpPacketHandler func(ctx *fasthttp.RequestCtx) error

	httpPacketHandler struct {
		handler pHttpPacketHandler
	}
)

func (ph *httpPacketHandler) Handle(ctx *fasthttp.RequestCtx) error {
	return ph.handler(ctx)
}

func NewHttpPacketHandler(handler pHttpPacketHandler) (*httpPacketHandler, error) {
	return &httpPacketHandler{
		handler: handler,
	}, nil
}
