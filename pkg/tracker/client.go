package tracker

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/valhalla/go-torrent/pkg/torrent"
)

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"` // Compact format usually
	// Full list format support would require interface{} and type assertion
}

// BuildTrackerURL creates the URL to query the HTTP tracker.
func BuildTrackerURL(t *torrent.TorrentFile, peerID string, port uint16, uploaded, downloaded, left int64) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{peerID},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{strconv.FormatInt(uploaded, 10)},
		"downloaded": []string{strconv.FormatInt(downloaded, 10)},
		"left":       []string{strconv.FormatInt(left, 10)},
		"compact":    []string{"1"},
		"event":      []string{"started"}, // Can be started, stopped, completed
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// Client handles communication with trackers.
type Client struct {
	Timeout time.Duration
}

func NewClient() *Client {
	return &Client{
		Timeout: 15 * time.Second,
	}
}

// AnnounceHTTP sends an announce request to an HTTP tracker.
func (c *Client) AnnounceHTTP(announceURL string) ([]Peer, error) {
	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Get(announceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned non-200 status: %d", resp.StatusCode)
	}

	// We need to parse the response
	// The response is bencoded.
	// We'll use our bencode package (assuming we expanded it or use manual parsing).
	// For this step, I'll return a placeholder or need to implement the parsing properly.
	return ParseTrackerResponse(resp.Body)
}
