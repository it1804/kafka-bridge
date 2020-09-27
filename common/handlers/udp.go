package handlers

type (
	UdpPacketHandler interface {
		Handle(payload []byte, len int) error
	}

	pUdpPacketHandler func(payload []byte, len int) error

	udpPacketHandler struct {
		handler pUdpPacketHandler
	}
)

func (ph *udpPacketHandler) Handle(payload []byte, length int) error {
	return ph.handler(payload, length)
}

func NewUdpPacketHandler(handler pUdpPacketHandler) (*udpPacketHandler, error) {
	return &udpPacketHandler{
		handler: handler,
	}, nil
}
