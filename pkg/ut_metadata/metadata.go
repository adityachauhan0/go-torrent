package ut_metadata

import (
	"bytes"
	"fmt"
	"io"
	
	"github.com/valhalla/go-torrent/pkg/bencode"
)

const (
	MsgTypeRequest = 0
	MsgTypeData    = 1
	MsgTypeReject  = 2
)

// MetadataMessage represents the BEP 9 message request/response.
type MetadataMessage struct {
	MsgType int `bencode:"msg_type"`
	Piece   int `bencode:"piece"`
	// TotalSize is only present in 'data' messages
	TotalSize int `bencode:"total_size,omitempty"` 
}

// ParseMetadataMessage decodes the bencoded dictionary at the start of the payload.
func ParseMetadataMessage(payload []byte) (MetadataMessage, []byte, error) {
	reader := bytes.NewReader(payload)
	val, err := bencode.Decode(reader)
	if err != nil {
		return MetadataMessage{}, nil, err
	}
	
	msgMap, ok := val.(map[string]interface{})
	if !ok {
		return MetadataMessage{}, nil, fmt.Errorf("metadata message is not a dictionary")
	}
	
	var msg MetadataMessage
	if t, ok := msgMap["msg_type"].(int); ok {
		msg.MsgType = t
	} else {
		return MetadataMessage{}, nil, fmt.Errorf("missing msg_type")
	}
	
	if p, ok := msgMap["piece"].(int); ok {
		msg.Piece = p
	}
	
	if ts, ok := msgMap["total_size"].(int); ok {
		msg.TotalSize = ts
	}
	
	// The rest of the payload is the actual data (if MsgTypeData)
	// We need to know how many bytes we read to key the offset.
	// Since bencode.Decode reads from the reader, we can check current offset.
	offset, _ := reader.Seek(0, io.SeekCurrent)
	data := payload[offset:]
	
	return msg, data, nil
}
