package ut_metadata

import (
	"bytes"

	"github.com/valhalla/go-torrent/pkg/bencode"
)

const (
	// ExtensionHandshakeID is the extended message ID for the handshake
	ExtensionHandshakeID = 0
	// MetadataID is the identifier we will use for the ut_metadata extension
	MetadataID = 1
)

// ExtensionHandshake represents the payload for the extension handshake.
// BEP 10: 'm' dictionary maps extension names to message IDs.
type ExtensionHandshake struct {
	M            map[string]int `bencode:"m"`
	MetadataSize int            `bencode:"metadata_size,omitempty"` // For BEP 9
}

// NewHandshake creates a new extension handshake payload supported by this client.
func NewHandshake(metadataSize int) ExtensionHandshake {
	return ExtensionHandshake{
		M: map[string]int{
			"ut_metadata": MetadataID,
		},
		MetadataSize: metadataSize,
	}
}

// SerializeHandshake bencodes the handshake map.
// Note: We need to wrap it in a proper wire message format later.
// The wire format for extended messages is: <len><20><extended_msg_id><payload>
func SerializeHandshake(h ExtensionHandshake) ([]byte, error) {
	// Our bencode package encodes structs if they are mapped to map[string]interface{}.
	// Since our current bencode package is simple (per analysis), we might need to convert manually
	// or rely on `bencode.Encode` handling map[string]interface{}.

	m := map[string]interface{}{
		"m": map[string]interface{}{
			"ut_metadata": MetadataID,
		},
	}
	if h.MetadataSize > 0 {
		m["metadata_size"] = h.MetadataSize
	}

	var buf bytes.Buffer
	err := bencode.Encode(&buf, m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FormatExtendedMessage creates the full byte array for a peer message.
// msgID: The extension ID (0 for handshake, or the one assigned by peer for other messages)
func FormatExtendedMessage(msgID uint8, payload []byte) []byte {
	// payload length + 1 (for extended msg ID)
	buf := make([]byte, 1+len(payload))
	buf[0] = byte(msgID)
	copy(buf[1:], payload)

	return buf
}
