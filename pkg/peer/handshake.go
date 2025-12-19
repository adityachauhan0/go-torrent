package peer

import (
	"fmt"
	"io"
)

// Handshake is a special message used to establish a connection.
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// NewHandshake creates a new handshake for the given infohash and peerID.
func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Serialize writes the handshake to w.
func (h *Handshake) Serialize(w io.Writer) error {
	pstrlen := len(h.Pstr)
	buf := make([]byte, 1+pstrlen+8+20+20) // 49 + pstrlen
	
	buf[0] = byte(pstrlen)
	copy(buf[1:1+pstrlen], h.Pstr)
	// Leave 8 reserved bytes as 0
	copy(buf[1+pstrlen+8:], h.InfoHash[:])
	copy(buf[1+pstrlen+8+20:], h.PeerID[:])
	
	_, err := w.Write(buf)
	return err
}

// ReadHandshake reads a handshake from r.
func ReadHandshake(r io.Reader) (*Handshake, error) {
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, err
	}
	
	pstrlen := int(lenBuf[0])
	if pstrlen == 0 {
		return nil, fmt.Errorf("pstrlen cannot be 0")
	}

	buf := make([]byte, pstrlen+8+20+20)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	
	pstr := string(buf[0:pstrlen])
	
	var infoHash [20]byte
	var peerID [20]byte
	
	copy(infoHash[:], buf[pstrlen+8 : pstrlen+8+20])
	copy(peerID[:], buf[pstrlen+8+20 : pstrlen+8+20+20])
	
	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
