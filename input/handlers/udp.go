package handlers

import (
	"net"
)

type (
	UdpPacketHandler interface {
		Handle(payload []byte, len int, src net.IP) error
	}

	pUdpPacketHandler func(payload []byte, len int, src net.IP) error

	udpPacketHandler struct {
		handler pUdpPacketHandler
	}
)

func (ph *udpPacketHandler) Handle(payload []byte, length int, src net.IP) error {
	return ph.handler(payload, length, src)
}

func NewUdpPacketHandler(handler pUdpPacketHandler) (*udpPacketHandler, error) {
	return &udpPacketHandler{
		handler: handler,
	}, nil
}
