package utils

import (
    "log"
    "net"
    "os"
    "context"
    "syscall"
    "os/signal"
    "sync/atomic"
    "time"
)

var ops uint64 = 0
var total uint64 = 0
var flushTicker *time.Ticker

type (
    ListenServer struct {
	proto string
	address string
	workers int
	packetSize int
    }
)

const ( 
    flushInterval = time.Duration(30) * time.Second
)


func NewListenServer(proto string, address string, packetSize int, workers int) *ListenServer {
    return &ListenServer {
	    proto: proto, 
	    address: address,
	    workers: workers,
	    packetSize: packetSize,
    }
}

func (r *ListenServer) Listen(output chan []byte) (err error) {

    ctx := context.Background()
    lc := net.ListenConfig {
        Control: func(network, address string, c syscall.RawConn) error {
            var opErr error
            err := c.Control(func(fd uintptr) {
//                opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, 65536)
                opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
            })
            if err != nil {
                return err
            }
            return opErr
        }, 
    }

    lp, err := lc.ListenPacket(ctx, "udp", r.address)
    if err != nil {
        log.Printf("UDP listen error: %s\n", err)
	os.Exit(0)
    } else {
        log.Printf("UDP new listener on %s\n",r.address)
    }
    conn := lp.(*net.UDPConn)

    for i := 0; i < r.workers; i++ {
	go func() {
    	    receive(conn,output,r)
	}()
    }


    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
	for range c {
	    atomic.AddUint64(&total, ops)
	    log.Printf("UDP total packets received: %d\n", total)
	    os.Exit(0)
	}
    }()
    return
}

func (r *ListenServer) PrintStat() () {
    flushTicker = time.NewTicker(flushInterval)
    for range flushTicker.C {
	atomic.AddUint64(&total, ops)
	log.Printf("UDP received packets: %d (pps: %f)\n",total, float64(ops)/flushInterval.Seconds())
	atomic.StoreUint64(&ops, 0)
    }
}

func receive(c net.PacketConn, output chan []byte,r *ListenServer) {
    defer c.Close()
    msg := make([]byte, r.packetSize)
    for {
        nbytes, _, err := c.ReadFrom(msg[0:])
        if err != nil {
            log.Printf("UDP receive error %s\n", err)
            continue
        }
	output <- msg[:nbytes]
	atomic.AddUint64(&ops, 1)
    }
}
