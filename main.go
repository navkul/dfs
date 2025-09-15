package main

import (
	"log"
	"time"

	"github.com/navkul/dfs/p2p"
)

func main() {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		// TODO: onPeer func
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StoreRoot:     "3000_network",
		PathTransform: CASPathTransformFunc,
		Transport:     tcpTransport,
	}

	s := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(1 * time.Second)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

}
