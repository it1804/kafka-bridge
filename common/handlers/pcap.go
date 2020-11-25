package handlers

import (
	"github.com/google/gopacket"
)

type (
	PcapPacketHandler interface {
		Handle(packet gopacket.Packet) error
	}

	pPcapPacketHandler func(packet gopacket.Packet) error

	pcapPacketHandler struct {
		handler pPcapPacketHandler
	}
)

func (ph *pcapPacketHandler) Handle(packet gopacket.Packet) error {
	return ph.handler(packet)
}

func NewPcapPacketHandler(handler pPcapPacketHandler) (*pcapPacketHandler, error) {
	return &pcapPacketHandler{
		handler: handler,
	}, nil
}
