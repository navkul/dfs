package main

import (
	"bytes"
	"log"
	"time"

	"github.com/navkul/dfs/p2p"
)

func makeServer(listenAddress string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StoreRoot:      listenAddress + "_network",
		PathTransform:  CASPathTransformFunc,
		Transport:      tcpTransport,
		BootStrapNodes: nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go s2.Start()

	time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte("hello world"))
		s2.Store("myprivatedata", data)
		time.Sleep(5 * time.Millisecond)
	}

	select {}

}
