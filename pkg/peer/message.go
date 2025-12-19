package peer

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageID uint8

const (
	MsgChoke         MessageID = 0
	MsgUnchoke       MessageID = 1
	MsgInterested    MessageID = 2
	MsgNotInterested MessageID = 3
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
)

// Message stores ID and payload of a message
type Message struct {
	ID      MessageID
	Payload []byte
}

// FormatRequest creates a REQUEST message payload.
func FormatRequest(index, begin, length int) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return payload
}

// FormatHave creates a HAVE message payload.
func FormatHave(index int) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	return payload
}

// ParseHave parses a HAVE message payload.
func ParseHave(payload []byte) (int, error) {
	if len(payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got %d", len(payload))
	}
	return int(binary.BigEndian.Uint32(payload)), nil
}

// ParsePiece parses a PIECE message payload.
func ParsePiece(index int, buf []byte, msg Message) (int, []byte, error) {
	if msg.ID != MsgPiece {
		return 0, nil, fmt.Errorf("expected PIECE (7), got %d", msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, nil, fmt.Errorf("payload too short")
	}
	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, nil, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	data := msg.Payload[8:]
	return begin, data, nil
}

// ParseRequest parses a REQUEST message payload.
func ParseRequest(msg Message) (int, int, int, error) {
	if msg.ID != MsgRequest {
		return 0, 0, 0, fmt.Errorf("expected REQUEST (6), got %d", msg.ID)
	}
	if len(msg.Payload) != 12 {
		return 0, 0, 0, fmt.Errorf("expected payload length 12, got %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	length := int(binary.BigEndian.Uint32(msg.Payload[8:12]))
	return index, begin, length, nil
}

// FormatPiece creates a PIECE message payload.
func FormatPiece(index, begin int, block []byte) []byte {
	payload := make([]byte, 8+len(block))
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], block)
	return payload
}

// Serialize writes bits to w. Format: <length prefix><message ID><payload>
func (m *Message) Serialize(w io.Writer) error {
	if m == nil {
		// Keep-alive message (len=0)
		_, err := w.Write([]byte{0, 0, 0, 0})
		return err
	}

	length := uint32(len(m.Payload) + 1) // +1 for ID
	buf := make([]byte, 4+length)
	
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	
	_, err := w.Write(buf)
	return err
}

// ReadMessage reads a message from a stream.
func ReadMessage(r io.Reader) (*Message, error) {
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lenBuf)
	if err != nil {
		return nil, err
	}
	
	length := binary.BigEndian.Uint32(lenBuf)
	
	// Keep-alive message
	if length == 0 {
		return nil, nil // Nil message represents keep-alive
	}
	
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	
	return &Message{
		ID:      MessageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}
