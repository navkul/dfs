package main

import (
	"fmt"
	"log"

	"github.com/navkul/dfs/p2p"
)

func OnPeer(p2p.Peer) error {
	fmt.Println("executing logic outside TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
