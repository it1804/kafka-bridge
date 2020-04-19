package utils

import (
    "fmt"
    "net"
    "os"
    "context"
    "runtime"
)

type (
    ListenServer struct {
	proto string
	address string
    }
)

const ( 
    UDPPacketSize = 1500
    packetBufSize = 1024 * 1024 // 1 MB
)


func NewListenServer(proto string, address string) *ListenServer {
    return &ListenServer{proto: proto, address: address}
}

func (r *ListenServer) Run(ctx context.Context,output chan []byte) (err error) {

    c, err := net.ListenPacket(r.proto, r.address)
    if err != nil {
	return
    }
    doneChan := make(chan error, 1)
    for i := 0; i < runtime.NumCPU(); i++ {
	go func() {
    	    receive(c,output)
	}()
    }

    select {
	case <-ctx.Done():
	    fmt.Println("cancelled")
	    err = ctx.Err()
	case err = <-doneChan:
    }
    return
}

func receive(c net.PacketConn, output chan []byte) {
    defer c.Close()
    var buf []byte
//    var gid = getGID()
    for {
	if len(buf) < UDPPacketSize {
	    buf = make([]byte, packetBufSize, packetBufSize)
	}
	nbytes, _, err := c.ReadFrom(buf)
//	 fmt.Fprintf(os.Stderr, "Read: %d bytes thread %d\n", gid, nbytes)
	if err != nil {
	    fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	    continue
	}
	msg := buf[:nbytes]
	buf = buf[nbytes:]
	output <- msg
    }
}
