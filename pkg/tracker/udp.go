package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"time"

	"github.com/valhalla/go-torrent/pkg/torrent"
)

const (
	protocolID int64 = 0x41727101980
	actionConnect int32 = 0
	actionAnnounce int32 = 1
)

type UDPClient struct {
	Timeout time.Duration
}

func NewUDPClient() *UDPClient {
	return &UDPClient{
		Timeout: 5 * time.Second,
	}
}

// AnnounceUDP performs the connect and announce flow for a UDP tracker.
func (c *UDPClient) AnnounceUDP(announceURL string, t *torrent.TorrentFile, peerID string, port uint16) ([]Peer, error) {
	// Parse URL
	parsed, err := url.Parse(announceURL)
	if err != nil {
		return nil, err
	}
	
	host := parsed.Host
	if parsed.Port() == "" {
		host = net.JoinHostPort(parsed.Host, "80") // Default if missing, though typically 6969
	}

	conn, err := net.DialTimeout("udp", host, c.Timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(c.Timeout * 2)) // Overall deadline

	// 1. Connect
	connID, err := c.sendConnect(conn)
	if err != nil {
		return nil, fmt.Errorf("udp connect failed: %w", err)
	}

	// 2. Announce
	return c.sendAnnounce(conn, connID, t, peerID, port)
}

func (c *UDPClient) sendConnect(conn net.Conn) (int64, error) {
	transactionID := rand.Int31()
	
	req := bytes.NewBuffer(make([]byte, 0, 16))
	binary.Write(req, binary.BigEndian, protocolID)
	binary.Write(req, binary.BigEndian, actionConnect)
	binary.Write(req, binary.BigEndian, transactionID)

	if _, err := conn.Write(req.Bytes()); err != nil {
		return 0, err
	}

	resBuf := make([]byte, 16)
	n, err := conn.Read(resBuf)
	if err != nil {
		return 0, err
	}
	if n < 16 {
		return 0, fmt.Errorf("read too short")
	}

	var action int32
	var resTxID int32
	var connID int64
	
	r := bytes.NewReader(resBuf)
	binary.Read(r, binary.BigEndian, &action)
	binary.Read(r, binary.BigEndian, &resTxID)
	binary.Read(r, binary.BigEndian, &connID)

	if action != actionConnect {
		return 0, fmt.Errorf("invalid action: %d", action)
	}
	if resTxID != transactionID {
		return 0, fmt.Errorf("transaction id mismatch")
	}

	return connID, nil
}

func (c *UDPClient) sendAnnounce(conn net.Conn, connID int64, t *torrent.TorrentFile, peerID string, port uint16) ([]Peer, error) {
	transactionID := rand.Int31()
	
	req := bytes.NewBuffer(make([]byte, 0, 98))
	binary.Write(req, binary.BigEndian, connID)
	binary.Write(req, binary.BigEndian, actionAnnounce)
	binary.Write(req, binary.BigEndian, transactionID)
	
	req.Write(t.InfoHash[:])
	req.Write([]byte(peerID))
	
	binary.Write(req, binary.BigEndian, int64(0)) // Downloaded
	binary.Write(req, binary.BigEndian, t.Info.Length) // Left (Assuming initial)
	binary.Write(req, binary.BigEndian, int64(0)) // Uploaded
	binary.Write(req, binary.BigEndian, int32(2)) // Event: Started (2)
	binary.Write(req, binary.BigEndian, uint32(0)) // IP default
	binary.Write(req, binary.BigEndian, uint32(rand.Int31())) // Key
	binary.Write(req, binary.BigEndian, int32(-1)) // Num Want (-1 default)
	binary.Write(req, binary.BigEndian, uint16(port)) 

	if _, err := conn.Write(req.Bytes()); err != nil {
		return nil, err
	}

	// Read Response
	// Minimum 20 bytes (Action(4) + TxID(4) + Interval(4) + Leechers(4) + Seeders(4))
	// + Peers (6 bytes each)
	resBuf := make([]byte, 4096) 
	n, err := conn.Read(resBuf)
	if err != nil {
		return nil, err
	}
	
	if n < 20 {
		return nil, fmt.Errorf("response too short")
	}
	
	r := bytes.NewReader(resBuf[:n])
	
	var action int32
	var resTxID int32
	
	binary.Read(r, binary.BigEndian, &action)
	binary.Read(r, binary.BigEndian, &resTxID)
	
	if action != actionAnnounce {
		return nil, fmt.Errorf("invalid announce action: %d", action)
	}
	if resTxID != transactionID {
		return nil, fmt.Errorf("transaction id mismatch in announce")
	}
	
	// Skip interval, leechers, seeders
	r.Seek(12, io.SeekCurrent)
	
	// Parse peers
	peers := []Peer{}
	for r.Len() >= 6 {
		var ipData [4]byte
		var portVal uint16
		r.Read(ipData[:])
		binary.Read(r, binary.BigEndian, &portVal)
		
		ip := net.IP(ipData[:])
		peers = append(peers, Peer{IP: ip, Port: portVal})
	}
	
	return peers, nil
}
