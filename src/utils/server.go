package utils

import (
    "fmt"
    "net"
    "os"
    "context"
    "runtime"
    "bytes"
    "strconv"
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

func getGID() uint64 {
    b := make([]byte, 64)
    b = b[:runtime.Stack(b, false)]
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}


func NewListenServer(proto string,address string) *ListenServer {
    return &ListenServer{proto: proto, address: address}
}

func (r *ListenServer) Run(ctx context.Context,output chan []byte) (err error) {

    c, err := net.ListenPacket(r.proto, r.address)
    if err != nil {
	return
    }
    defer c.Close()
    doneChan := make(chan error, 1)
      for i := 0; i < 3; i++ {
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

func receive(c net.PacketConn,output chan []byte) {
    defer c.Close() // TODO: closed multiple times

    var buf []byte
    for {
//	fmt.Fprintf(os.Stderr, "Current GID: %d\n", getGID())
	if len(buf) < UDPPacketSize {
	    buf = make([]byte, packetBufSize, packetBufSize)
	}
	nbytes, _, err := c.ReadFrom(buf)
	if err != nil {
	    fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	    continue
	}
	msg := buf[:nbytes]
	buf = buf[nbytes:]
	output <- msg
    }
}
