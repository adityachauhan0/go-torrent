package peer

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/valhalla/go-torrent/pkg/bitfield"
	"github.com/valhalla/go-torrent/pkg/p2p"
)

// Peer represents a connected peer.
type Peer struct {
	Conn     *p2p.Client
	Choked   bool
	Bitfield bitfield.Bitfield
	PeerID   [20]byte
	IP       net.IP
	Port     uint16
}

func NewPeer(ip net.IP, port uint16, peerID [20]byte) *Peer {
	return &Peer{
		IP:     ip,
		Port:   port,
		PeerID: peerID,
		Choked: true, // Peers start choked
	}
}

// Connect establishes a TCP connection, performs the handshake, and returns the peer.
func (p *Peer) Connect(infoHash, clientPeerID [20]byte, bf bitfield.Bitfield) error {
	address := net.JoinHostPort(p.IP.String(), fmt.Sprintf("%d", p.Port))
	conn, err := p2p.New(address, 30*time.Second)
	if err != nil {
		return err
	}
	p.Conn = conn

	// Handshake
	h := NewHandshake(infoHash, clientPeerID)
	if err := h.Serialize(p.Conn); err != nil {
		p.Conn.Close()
		return err
	}

	resHandshake, err := ReadHandshake(p.Conn)
	if err != nil {
		p.Conn.Close()
		return err
	}

	if resHandshake.InfoHash != infoHash {
		p.Conn.Close()
		return fmt.Errorf("info hash mismatch")
	}

	// Send Bitfield
	if len(bf) > 0 {
		if err := p.SendBitfield(bf); err != nil {
			p.Conn.Close()
			return err
		}
	}

	// Should verify PeerID if we know it (tracker might give it)
	// For now, accept whatever unless we want to enforce it.
	
	log.Printf("Connected to peer %s", p.IP)
	return nil
}

// ReadMessage reads the next message from the peer.
func (p *Peer) ReadMessage() (*Message, error) {
	return ReadMessage(p.Conn)
}

// SendMessage sends a message to the peer.
func (p *Peer) SendMessage(msg *Message) error {
	return msg.Serialize(p.Conn)
}

// SendUnchoke sends an Unchoke message.
func (p *Peer) SendUnchoke() error {
	msg := &Message{ID: MsgUnchoke}
	return p.SendMessage(msg)
}

// SendInterested sends an Interested message.
func (p *Peer) SendInterested() error {
	msg := &Message{ID: MsgInterested}
	return p.SendMessage(msg)
}

// SendRequest sends a Request message.
func (p *Peer) SendRequest(index, begin, length int) error {
	payload := FormatRequest(index, begin, length)
	msg := &Message{ID: MsgRequest, Payload: payload}
	return p.SendMessage(msg)
}
// SendBitfield sends a Bitfield message.
func (p *Peer) SendBitfield(bf bitfield.Bitfield) error {
	msg := &Message{ID: MsgBitfield, Payload: bf}
	return p.SendMessage(msg)
}
