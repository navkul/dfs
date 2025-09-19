package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/navkul/dfs/p2p"
)

type FileServerOpts struct {
	ListenAddress  string
	StoreRoot      string
	PathTransform  PathTransformFunc
	Transport      p2p.Transport
	BootStrapNodes []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StoreRoot,
		PathTransformFunc: opts.PathTransform,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

// ** insane go lang engineering
func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buffer := new(bytes.Buffer)
	if err := gob.NewEncoder(buffer).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buffer.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}

	fmt.Printf("Don't have file (%s) locally, fetching from network", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Millisecond)

	for _, peer := range s.peers {
		fileBuffer := new(bytes.Buffer)
		n, err := io.Copy(fileBuffer, peer)
		if err != nil {
			return nil, err
		}
		fmt.Println(n, " bytes received over the network, and written to disk")
	}

	return nil, nil

}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store this file to disc
	// 2. Broadcast this file to all KNOWN peers in the network
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	// TODO: use a multiwriter here
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Println(n, "bytes received and written to disk")
	}

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	log.Printf("Connected with remote %s", p.RemoteAddr().String())
	return nil
}

func (s *FileServer) loop() {

	defer func() {
		log.Println("File server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("Decoding error:", err)
			}
			if err := s.handleMessage(&msg, rpc.From); err != nil {
				log.Println("Handle error:", err)
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(msg *Message, from string) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:

		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("Need to serve file (%s) but it does not exist on disk", msg.Key)
	}

	fmt.Printf("Serving file (%s) over the network\n", msg.Key)
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not known in network", from)
	}

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("Wtiten %d bytes over the network to %s ", n, from)
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in the peer map", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("[%s] wrote %d bytes to disk\n", s.Transport.Addr(), n)

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (s *FileServer) bootStrapNetwork() error {

	for _, addr := range s.BootStrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Println("Attempting to connect with remote: ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootStrapNetwork()

	s.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
