package p2p

// Message represents any arbitrary data that is sent over each
// transport between two nodes in the network.
type Message struct {
	Payload []byte
}
