package input

import (
	"context"
	"github.com/it1804/kafka-bridge/common/handlers"
	"log"
	"net"
	"sync"
)

type (
	UdpServerConf struct {
		Listen        string
		Workers       uint
		MaxPacketSize uint32
	}

	udpConnHandler struct {
		phandler handlers.UdpPacketHandler
		stop     bool
		server   *UdpServer
	}

	UdpServer struct {
		wg          *sync.WaitGroup
		serviceName string
		conf        *UdpServerConf
		bufferPool  sync.Pool
	}
)

func NewUdpServer(serviceName string, conf *UdpServerConf) *UdpServer {
	return &UdpServer{
		conf:        conf,
		wg:          &sync.WaitGroup{},
		serviceName: serviceName,
		bufferPool:  sync.Pool{New: func() interface{} { return make([]byte, conf.MaxPacketSize) }},
	}
}

func (r *UdpServer) Shutdown() {
	r.wg.Wait()
	log.Printf("[%s] UDP server stopped", r.serviceName)
	return
}

func (r *UdpServer) Run(ctx context.Context, phandler handlers.UdpPacketHandler) (err error) {

	handler := &udpConnHandler{
		phandler: phandler,
		stop:     false,
		server:   r,
	}

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", r.conf.Listen)
	if err != nil {
		return
	}
	conn := lp.(*net.UDPConn)
	for i := uint(0); i < r.conf.Workers; i++ {
		r.wg.Add(1)
		go handler.receivePacket(conn, r.wg, r.conf.MaxPacketSize, i)
	}
	select {
	case <-ctx.Done():
		handler.stop = true
		lp.Close()
		return nil
	}
	return nil
}

func (h *udpConnHandler) receivePacket(c net.PacketConn, wg *sync.WaitGroup, packetSize uint32, worker_id uint) {
	defer wg.Done()
	defer c.Close()
	for h.stop == false {
		msg := h.server.bufferPool.Get().([]byte)
		nbytes, addr, err := c.ReadFrom(msg)

		if h.stop {
			break
		}
		if err != nil {
			continue
		}
		src := addr.((*net.UDPAddr)).IP
		h.phandler.Handle(msg[:nbytes], nbytes, src)
		h.server.bufferPool.Put(msg)
	}
}
