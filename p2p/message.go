package p2p

// RPC represents any arbitrary data that is sent over each
// transport between two nodes in the network.
type RPC struct {
	Payload []byte
	From    string
}
