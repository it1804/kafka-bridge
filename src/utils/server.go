package utils

import (
    "fmt"
    "net"
    "context"
)

type (
    ListenServer struct {
	proto string
	address string
    }
)


const maxBufferSize = 65535

func NewListenServer(proto string,address string) *ListenServer {
    return &ListenServer{proto: proto, address: address}
}

func (r *ListenServer) Run(ctx context.Context,output chan []byte) (err error) {

    pc, err := net.ListenPacket(r.proto, r.address)
    if err != nil {
	return
    }
    defer pc.Close()
    doneChan := make(chan error, 1)
    buffer := make([]byte, maxBufferSize)
    go func() {
	for {
	    n,_, err := pc.ReadFrom(buffer)
	    if err != nil {
		doneChan <- err
		return
	    }
	    output <- buffer[:n]
	}
    }()

    select {
    case <-ctx.Done():
	fmt.Println("cancelled")
	err = ctx.Err()
    case err = <-doneChan:
    }

    return
}
