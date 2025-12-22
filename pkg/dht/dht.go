package dht

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/valhalla/go-torrent/pkg/bencode"
)

// DHTConfig holds configuration for the DHT node.
type DHTConfig struct {
	Port           int
	BootstrapNodes []string
}

// DHT represents a Distributed Hash Table node.
type DHT struct {
	ID    NodeID
	Table *RoutingTable
	Conn  *net.UDPConn

	// Channels for handling responses (simplified for this implementation)
	// In a full implementation, we'd need a map of TransactionID -> Channel

	// Transaction management
	transactions map[string]chan Msg

	// Peers found (IP:Port)
	PeersRequest chan [20]byte
	PeersFound   chan string

	mu sync.Mutex
}

// NewDHT creates a new DHT node.
func NewDHT(config DHTConfig) (*DHT, error) {
	id := GenerateID() // Or load from disk

	addr, err := net.ResolveUDPAddr("udp", ":0") // Bind to random port or config.Port
	if config.Port != 0 {
		addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", config.Port))
	}
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP("127.0.0.1") // TODO: Resolve public IP

	node := Node{ID: id, IP: ip, Port: addr.Port}
	table := NewRoutingTable(node)

	dht := &DHT{
		ID:           id,
		Table:        table,
		Conn:         conn,
		transactions: make(map[string]chan Msg),
		PeersFound:   make(chan string, 100),
	}

	return dht, nil
}

// Start begins listening for UDP packets and bootstraps the node.
func (d *DHT) Start() {
	go d.listen()
	go d.Bootstrap()
}

// Bootstrap connects to the initial nodes to populate the routing table.
func (d *DHT) Bootstrap() {
	// 1. Ping bootstrap nodes
	defaults := []string{
		"router.bittorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"router.utorrent.com:6881",
	}

	for _, addrStr := range defaults {
		addr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			continue
		}
		d.sendPing(addr)
	}

	time.Sleep(1 * time.Second)

	// Find Self to populate near buckets
	d.FindNode(d.ID)
}

func (d *DHT) listen() {
	buf := make([]byte, 2048)
	for {
		n, rAddr, err := d.Conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("DHT Read error: %v", err)
			continue
		}

		// Handle packet asynchronously
		go d.handlePacket(buf[:n], rAddr)
	}
}

func (d *DHT) handlePacket(data []byte, addr *net.UDPAddr) {
	val, err := bencode.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	msgMap, ok := val.(map[string]interface{})
	if !ok {
		return
	}

	y, _ := msgMap["y"].(string)
	t, _ := msgMap["t"].(string)

	switch y {
	case "q":
		q, _ := msgMap["q"].(string)
		a, _ := msgMap["a"].(map[string]interface{})
		d.handleQuery(t, q, a, addr)
	case "r":
		r, _ := msgMap["r"].(map[string]interface{})
		d.handleResponse(t, r, addr)

		// passive check for 'values' (peers) in any response, typical in response to get_peers
		if values, ok := r["values"].([]interface{}); ok {
			for _, v := range values {
				if peerStr, ok := v.(string); ok {
					// PeerStr is typically byte string of IP+Port (6 bytes)
					if len(peerStr) == 6 {
						// Parse compact
						ip := net.IP([]byte(peerStr[0:4]))
						port := binary.BigEndian.Uint16([]byte(peerStr[4:6]))
						d.PeersFound <- fmt.Sprintf("%s:%d", ip.String(), port)
					}
				}
			}
		}
	}
}

func (d *DHT) handleQuery(tid string, q string, a map[string]interface{}, addr *net.UDPAddr) {
	// Basic query handling
	switch q {
	case "ping":
		d.sendResponse(tid, map[string]interface{}{"id": string(d.ID[:])}, addr)
	case "find_node":
		targetVal, ok := a["target"].(string)
		if ok && len(targetVal) == 20 {
			var target NodeID
			copy(target[:], targetVal)
			nodes := d.Table.NearestNodes(target)
			compact := d.compactNodes(nodes)
			d.sendResponse(tid, map[string]interface{}{"id": string(d.ID[:]), "nodes": compact}, addr)
		}
	case "get_peers":
		infoHashVal, ok := a["info_hash"].(string)
		if ok && len(infoHashVal) == 20 {
			var target NodeID
			copy(target[:], infoHashVal)
			nodes := d.Table.NearestNodes(target)
			compact := d.compactNodes(nodes)
			token := "token" // Simplified token
			d.sendResponse(tid, map[string]interface{}{"id": string(d.ID[:]), "nodes": compact, "token": token}, addr)
		}
	case "announce_peer":
		d.sendResponse(tid, map[string]interface{}{"id": string(d.ID[:])}, addr)
	}
}

func (d *DHT) handleResponse(tid string, r map[string]interface{}, addr *net.UDPAddr) {
	// Notify transaction waiter
	d.mu.Lock()
	ch, ok := d.transactions[tid]
	d.mu.Unlock()

	// Construct Msg for channel
	msg := Msg{
		Y: "r",
		R: r,
	}

	if ok {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (d *DHT) compactNodes(nodes []Node) string {
	var buf bytes.Buffer
	for _, n := range nodes {
		buf.Write(n.ID[:])
		// IP/Port encoding
		ip := n.IP.To4()
		buf.Write(ip)
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(n.Port))
		buf.Write(portBytes)
	}
	return buf.String()
}

// FindNode performs an iterative lookup for a node ID.
func (d *DHT) FindNode(target NodeID) {
	nodes := d.Table.NearestNodes(target)
	for _, n := range nodes {
		addr := &net.UDPAddr{IP: n.IP, Port: n.Port}
		d.sendFindNode(target, addr)
	}
}

// GetPeers looks up the DHT for peers sharing the InfoHash.
func (d *DHT) GetPeers(infoHash [20]byte) []string {
	target := NodeID(infoHash)
	nodes := d.Table.NearestNodes(target)
	for _, n := range nodes {
		addr := &net.UDPAddr{IP: n.IP, Port: n.Port}
		d.sendGetPeers(infoHash, addr)
	}
	// Note: In a real implementation this would collect responses channel
	// and return correct peers. For now we assume the bootstrap helps population.
	return nil
}

// Helpers for sending queries

func (d *DHT) sendPing(addr *net.UDPAddr) {
	tid := d.newTransaction()
	msg := NewQuery(tid, "ping", map[string]interface{}{"id": string(d.ID[:])})
	d.sendMsg(msg, addr)
}

func (d *DHT) sendFindNode(target NodeID, addr *net.UDPAddr) {
	tid := d.newTransaction()
	msg := NewQuery(tid, "find_node", map[string]interface{}{
		"id":     string(d.ID[:]),
		"target": string(target[:]),
	})
	d.sendMsg(msg, addr)
}

func (d *DHT) sendGetPeers(infoHash [20]byte, addr *net.UDPAddr) {
	tid := d.newTransaction()
	msg := NewQuery(tid, "get_peers", map[string]interface{}{
		"id":        string(d.ID[:]),
		"info_hash": string(infoHash[:]),
	})
	d.sendMsg(msg, addr)
}

func (d *DHT) sendResponse(tid string, r map[string]interface{}, addr *net.UDPAddr) {
	msg := NewResponse(tid, r)
	d.sendMsg(msg, addr)
}

func (d *DHT) sendMsg(msg Msg, addr *net.UDPAddr) {
	// Encode with bencode
	var buf bytes.Buffer
	// We need to map Msg struct to map for bencode currently
	data := map[string]interface{}{
		"t": string(msg.T),
		"y": msg.Y,
	}
	if msg.Q != "" {
		data["q"] = msg.Q
	}
	if msg.A != nil {
		data["a"] = msg.A
	}
	if msg.R != nil {
		data["r"] = msg.R
	}
	if msg.E != nil {
		data["e"] = msg.E
	}

	if err := bencode.Encode(&buf, data); err == nil {
		d.Conn.WriteToUDP(buf.Bytes(), addr)
	}
}

func (d *DHT) newTransaction() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Simple 2 byte TID
	b := make([]byte, 2)
	rand.Read(b) // crypto/rand needed? Or math/rand. Using math/rand imported via imports?
	// Wait, imports has "math/rand" and "crypto/rand" conflict?
	// routing.go used crypto/rand. dht.go doesn't import it yet.
	// Let's use simple counter or hex for now.
	return hex.EncodeToString(b)
}
