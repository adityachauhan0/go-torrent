package dht

import (
	"encoding/binary"
	"net"
)

// NodeID is a 20-byte unique identifier for a DHT node.
type NodeID [20]byte

// Node represents a participant in the DHT network.
type Node struct {
	ID   NodeID
	IP   net.IP
	Port int
}

// CompactNodeInfo returns the compact node info (26 bytes) used in KRPC messages.
func (n Node) Compact() []byte {
	buf := make([]byte, 26)
	copy(buf[0:20], n.ID[:])
	copy(buf[20:24], n.IP.To4())
	binary.BigEndian.PutUint16(buf[24:26], uint16(n.Port))
	return buf
}

// Distance calculates the XOR distance between two NodeIDs.
// Returns true if n is closer to target than other.
// Note: This is a simplified comparison (lexicographical check of XOR result).
// A full implementation would compare the XOR byte by byte.
func (a NodeID) Distance(b NodeID) []byte {
	dist := make([]byte, 20)
	for i := 0; i < 20; i++ {
		dist[i] = a[i] ^ b[i]
	}
	return dist
}
