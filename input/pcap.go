package input

import (
	"context"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/it1804/kafka-bridge/input/handlers"
	"log"
	"sync"
)

type (
	PcapSourceConf struct {
		Device     string
		SnapLen    uint32
		Expression string
	}

	PcapSource struct {
		conf        *PcapSourceConf
		wg          *sync.WaitGroup
		serviceName string
	}

	pcapPacketHandler struct {
		phandler handlers.PcapPacketHandler
	}
)

func NewPcapSource(serviceName string, conf *PcapSourceConf) *PcapSource {
	return &PcapSource{
		conf:        conf,
		wg:          &sync.WaitGroup{},
		serviceName: serviceName,
	}
}

func (p *PcapSource) Run(ctx context.Context, phandler handlers.PcapPacketHandler) (err error) {
	p.wg.Add(1)
	defer p.wg.Done()

	handler := &pcapPacketHandler{
		phandler: phandler,
	}

	handle, err := pcap.OpenLive(p.conf.Device, int32(p.conf.SnapLen), true, pcap.BlockForever)
	if err != nil {
		return
	}
	err = handle.SetBPFFilter(p.conf.Expression)
	if err != nil {
		return
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		select {
		case <-ctx.Done():
			return nil
		default:
			handler.handlePacket(packet)
		}
	}

	return nil
}

func (p *PcapSource) Shutdown() {
	p.wg.Wait()
	log.Printf("[%s] Pcap input stopped", p.serviceName)
	return
}

func (h *pcapPacketHandler) handlePacket(packet gopacket.Packet) {
	h.phandler.Handle(packet)

	return
}
