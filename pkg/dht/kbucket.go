package dht

import (
	"sort"
	"sync"
	"time"
)

const K = 8 // K-Bucket size

// KBucket holds a list of up to K nodes.
type KBucket struct {
	mu          sync.RWMutex
	nodes       []Node
	LastChanged time.Time
}

// NewKBucket creates a new KBucket.
func NewKBucket() *KBucket {
	return &KBucket{
		nodes: make([]Node, 0, K),
	}
}

// AddNode adds a node to the bucket if there is space or replaces a bad node.
// Returns true if the node was added or updated.
func (kb *KBucket) AddNode(n Node) bool {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	// Check if node already exists
	for i, existing := range kb.nodes {
		if existing.ID == n.ID {
			// Move to tail (most recently seen)
			// Efficient way: delete and append
			kb.nodes = append(kb.nodes[:i], kb.nodes[i+1:]...)
			kb.nodes = append(kb.nodes, n)
			kb.LastChanged = time.Now()
			return true
		}
	}

	// If bucket is not full, append
	if len(kb.nodes) < K {
		kb.nodes = append(kb.nodes, n)
		kb.LastChanged = time.Now()
		return true
	}

	// If full, simplistic approach: drop (in real Kademlia we'd ping the oldest)
	// For this concise implementation, we'll just return false (bucket full)
	return false
}

// Nodes returns a copy of the nodes in the bucket.
func (kb *KBucket) Nodes() []Node {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	nodes := make([]Node, len(kb.nodes))
	copy(nodes, kb.nodes)
	return nodes
}

// SortByDistance sorts the nodes in the bucket by distance to the target ID.
func (kb *KBucket) SortByDistance(target NodeID) []Node {
	nodes := kb.Nodes()
	sort.Slice(nodes, func(i, j int) bool {
		distI := nodes[i].ID.Distance(target)
		distJ := nodes[j].ID.Distance(target)
		// Lexicographical comparison of distance bytes
		for k := 0; k < 20; k++ {
			if distI[k] < distJ[k] {
				return true
			}
			if distI[k] > distJ[k] {
				return false
			}
		}
		return false
	})
	return nodes
}
