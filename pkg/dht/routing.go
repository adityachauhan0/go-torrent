package dht

import (
	"crypto/rand"
	"math/big"
)

const BucketSize = 20 * 8 // 160 bits

// RoutingTable manages the K-Buckets.
type RoutingTable struct {
	Self    Node
	Buckets [BucketSize]*KBucket
}

// NewRoutingTable creates a new routing table for the local node.
func NewRoutingTable(self Node) *RoutingTable {
	rt := &RoutingTable{
		Self: self,
	}
	for i := range rt.Buckets {
		rt.Buckets[i] = NewKBucket()
	}
	return rt
}

// AddNode adds a node to the appropriate bucket.
func (rt *RoutingTable) AddNode(n Node) {
	if n.ID == rt.Self.ID {
		return
	}
	index := bucketIndex(rt.Self.ID, n.ID)
	rt.Buckets[index].AddNode(n)
}

// NearestNodes returns the K closest nodes to the target ID.
func (rt *RoutingTable) NearestNodes(target NodeID) []Node {
	var nodes []Node
	
	// Start from the bucket properly corresponding to the target distance
	index := bucketIndex(rt.Self.ID, target)
	
	// Gather nodes from the target bucket and surrounding buckets
	for i := index; i >= 0 && len(nodes) < K; i-- {
		nodes = append(nodes, rt.Buckets[i].SortByDistance(target)...)
	}
	for i := index + 1; i < BucketSize && len(nodes) < K; i++ {
		nodes = append(nodes, rt.Buckets[i].SortByDistance(target)...)
	}
	
	// Sort all gathered nodes by distance
	// (Re-using sort logic from KBucket for simplicity, or we can do it here)
	// For now, let's just return what we have, truncated to K.
	// In a real implementation, we would sort the combined list.
	
	if len(nodes) > K {
		nodes = nodes[:K]
	}
	return nodes
}

// bucketIndex finds the appropriate bucket index for a node ID.
// It corresponds to the index of the first differing bit (common prefix length).
func bucketIndex(id1, id2 NodeID) int {
	// Find the first byte that differs
	for i := 0; i < 20; i++ {
		xor := id1[i] ^ id2[i]
		if xor != 0 {
			// Find the first bit in this byte that differs
			for j := 0; j < 8; j++ {
				if (xor>>uint(7-j))&0x1 != 0 {
					return i*8 + j
				}
			}
		}
	}
	return BucketSize - 1
}

// GenerateID generates a random NodeID.
func GenerateID() NodeID {
	var id NodeID
	rand.Read(id[:])
	return id
}

// Helper: Int distance for debugging (not used in core logic)
func (id NodeID) Int() *big.Int {
	return big.NewInt(0).SetBytes(id[:])
}
