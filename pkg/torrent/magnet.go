package torrent

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

// MagnetLink represents the data extracted from a magnet URI.
type MagnetLink struct {
	InfoHash    [20]byte
	DisplayName string
	Trackers    []string
}

// ParseMagnetURI parses a magnet link and returns a MagnetLink struct.
// Format: magnet:?xt=urn:btih:<hash>&dn=<name>&tr=<tracker>
func ParseMagnetURI(uri string) (*MagnetLink, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "magnet" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	q := u.Query()
	xt := q.Get("xt")
	if !strings.HasPrefix(xt, "urn:btih:") {
		return nil, fmt.Errorf("invalid xt parameter: %s", xt)
	}

	hashStr := strings.TrimPrefix(xt, "urn:btih:")
	var infoHash [20]byte

	// Hash can be hex or base32. For now, let's assume hex and support base32 if needed.
	// Most magnet links use hex.
	if len(hashStr) == 40 {
		h, err := hex.DecodeString(hashStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hex hash: %v", err)
		}
		copy(infoHash[:], h)
	} else if len(hashStr) == 32 {
		// Base32 support TODO
		return nil, fmt.Errorf("base32 hash not supported yet")
	} else {
		return nil, fmt.Errorf("invalid hash length: %d", len(hashStr))
	}

	m := &MagnetLink{
		InfoHash:    infoHash,
		DisplayName: q.Get("dn"),
		Trackers:    q["tr"],
	}

	return m, nil
}
