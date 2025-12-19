package tracker

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/valhalla/go-torrent/pkg/bencode"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), fmt.Sprintf("%d", p.Port))
}

// ParseTrackerResponse decodes the HTTP response body and extracts peers.
func ParseTrackerResponse(r io.Reader) ([]Peer, error) {
	val, err := bencode.Decode(r)
	if err != nil {
		return nil, err
	}

	root, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("tracker response is not a dict")
	}

	if failure, ok := root["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker failed: %s", failure)
	}

	// Helper to handle both string (compact) and list (dictionary) models for peers
	peersRaw, ok := root["peers"]
	if !ok {
		return nil, fmt.Errorf("peers not found in tracker response")
	}

	var peers []Peer

	switch v := peersRaw.(type) {
	case string:
		// Compact model: 6 bytes per peer (4 ip + 2 port)
		peersBytes := []byte(v)
		if len(peersBytes)%6 != 0 {
			return nil, fmt.Errorf("invalid peer list length: %d", len(peersBytes))
		}
		
		numPeers := len(peersBytes) / 6
		for i := 0; i < numPeers; i++ {
			offset := i * 6
			ip := net.IP(peersBytes[offset : offset+4])
			port := binary.BigEndian.Uint16(peersBytes[offset+4 : offset+6])
			peers = append(peers, Peer{IP: ip, Port: port})
		}
	
	case []interface{}:
		// Dictionary model
		for _, pVal := range v {
			pMap, ok := pVal.(map[string]interface{})
			if !ok {
				continue
			}
			
			ipStr, ok := pMap["ip"].(string)
			if !ok {
				continue
			}
			
			portInt, ok := pMap["port"].(int)
			if !ok {
				continue
			}
			
			peers = append(peers, Peer{
				IP:   net.ParseIP(ipStr),
				Port: uint16(portInt),
			})
		}
	
	default:
		return nil, fmt.Errorf("unsupported peers format: %T", peersRaw)
	}

	return peers, nil
}
