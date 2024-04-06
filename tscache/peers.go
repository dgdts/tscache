package tscache

import pb "tscache/tscachepb"

// PeerPicker is an interface for picking a peer based on a given key.
type PeerPicker interface {
	// PickPeer selects a peer based on the provided key.
	// It returns the selected PeerGetter and a boolean indicating whether a peer was found.
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is an interface for getting data from a peer.
type PeerGetter interface {
	// Get fetches the value associated with the provided key from a peer.
	// It takes a Request message as input and populates the Response message with the fetched data.
	Get(in *pb.Request, out *pb.Response) error
}
