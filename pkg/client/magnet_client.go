package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"

	"net/url"
	"strings"

	"github.com/valhalla/go-torrent/pkg/bencode"
	"github.com/valhalla/go-torrent/pkg/dht"
	"github.com/valhalla/go-torrent/pkg/peer"
	"github.com/valhalla/go-torrent/pkg/torrent"
	"github.com/valhalla/go-torrent/pkg/tracker"
	"github.com/valhalla/go-torrent/pkg/ut_metadata"
)

// MagnetClient handles the lifecycle of downloading a torrent from a magnet link.
type MagnetClient struct {
	Client *Client
	DHT    *dht.DHT
}

// NewMagnetClientWrapper creates a wrapper around the client for magnet operations.
func NewMagnetClientWrapper(c *Client) *MagnetClient {
	return &MagnetClient{
		Client: c,
	}
}

// Start initiates the magnet download process.
func (mc *MagnetClient) Start() error {
	log.Println("Starting Magnet Download...")
	mc.Client.AddLog("Starting Magnet Download...")
	mc.Client.Status = "Initializing"

	// Decode InfoHash from hex string to [20]byte
	var infoHash [20]byte
	if len(mc.Client.InfoHash) == 40 {
		hex.Decode(infoHash[:], []byte(mc.Client.InfoHash))
	}

	// 1. Start DHT
	mc.Client.Status = "Starting DHT"
	mc.Client.AddLog("Initializing DHT on port 6881...")
	config := dht.DHTConfig{Port: 6881}
	d, err := dht.NewDHT(config)
	if err != nil {
		return err
	}
	mc.DHT = d
	d.Start()
	log.Println("DHT Started, bootstrapping...")
	mc.Client.AddLog("DHT Started. Bootstrapping with default nodes...")
	mc.Client.Status = "Bootstrapping DHT"
	mc.Client.Status = "Bootstrapping DHT"

	// 1b. Announce to Trackers (Concurrent)
	if len(mc.Client.Trackers) > 0 {
		mc.Client.AddLog(fmt.Sprintf("Announcing to %d trackers...", len(mc.Client.Trackers)))
		go mc.announceToTrackers(infoHash)
	}

	// 2. Lookup Peers for InfoHash
	// InfoHash already decoded above

	// Decode hex (simplified, assuming valid input from previous steps)
	// In real app, import encoding/hex

	// Wait for bootstrap (simple delay)
	time.Sleep(2 * time.Second)

	log.Println("Looking up peers in DHT...")
	mc.Client.AddLog(fmt.Sprintf("Searching for peers for InfoHash: %x", infoHash))
	mc.Client.Status = "Searching Peers"
	// This triggers the iterative lookup. In a real threaded DHT, this would fill the bucket or returning peers channel.
	// For now we assume typical bootstrap nodes return peers or we find some.
	d.GetPeers(infoHash)

	// 3. Peer Discovery & Metadata Fetching
	log.Println("Waiting for peers...")
	mc.Client.AddLog("Waiting for peers...")

	timeout := time.After(30 * time.Second)
	var metadataFetched bool

	for {
		select {
		case peerAddr := <-d.PeersFound:
			// If already downloading, just add peer to downloader
			if metadataFetched && mc.Client.Downloader != nil {
				host, portStr, _ := net.SplitHostPort(peerAddr)
				port, _ := net.LookupPort("tcp", portStr)
				ip := net.ParseIP(host)

				// Add to downloader
				log.Printf("Adding peer %s to downloader", peerAddr)
				mc.Client.Downloader.AddPeer(tracker.Peer{IP: ip, Port: uint16(port)})
				continue
			}

			log.Printf("Found peer: %s", peerAddr)
			mc.Client.AddLog(fmt.Sprintf("Found peer: %s", peerAddr))
			mc.Client.Status = "Connecting to Peer"

			// Parse addr
			host, portStr, _ := net.SplitHostPort(peerAddr)
			port, _ := net.LookupPort("tcp", portStr)
			ip := net.ParseIP(host)

			// Re-encode infohash from hex to bytes (assuming client has it stored or we parse it)
			// For now assuming we parsed it above (need to add that logic)
			var infoHashBytes [20]byte
			if len(mc.Client.InfoHash) == 40 {
				hex.Decode(infoHashBytes[:], []byte(mc.Client.InfoHash))
			}

			p := peer.NewPeer(ip, uint16(port), [20]byte{}) // PeerID unknown initially

			// Attempt Fetch
			mc.Client.AddLog(fmt.Sprintf("Attempting to fetch metadata from %s...", peerAddr))
			infoDict, err := mc.FetchMetadata(p, infoHashBytes)
			if err == nil {
				log.Println("Metadata FETCHED SUCCESSFULLY!")
				mc.Client.AddLog("Metadata fetched successfully!")
				mc.Client.Status = "Downloading"

				// Construct TorrentFile
				tf := &torrent.TorrentFile{
					InfoHash: infoHashBytes,
					Info:     infoDict,
				}

				// Start Downloader
				mc.Client.Downloader = NewDownloader(tf, mc.Client.PeerID, mc.Client.OutputDir)
				go mc.Client.Downloader.TrackSpeed()

				// Add this peer to downloader too!
				mc.Client.Downloader.AddPeer(tracker.Peer{IP: ip, Port: uint16(port)})

				metadataFetched = true
				// Do NOT return. Continue loop to find more peers.
			}

		case <-timeout:
			if metadataFetched {
				// If we have metadata, we don't care about timeout, just keep looking for peers forever (or until stop)
				// Reset timeout? Or just ignore.
				timeout = time.After(30 * time.Second) // Keep ticking
				continue
			}

			mc.Client.AddLog("Timeout waiting for peers.")
			mc.Client.Status = "Stalled (No Peers)"
			return fmt.Errorf("timeout waiting for peers (and metadata)")
		}
	}
}

// FetchMetadata attempts to fetch metadata from a peer
// FetchMetadata attempts to fetch metadata from a peer
func (mc *MagnetClient) FetchMetadata(p *peer.Peer, infoHash [20]byte) (torrent.InfoDictionary, error) {
	// 1. Connect & Handshake
	err := p.Connect(infoHash, mc.Client.PeerID, nil)
	if err != nil {
		return torrent.InfoDictionary{}, err
	}
	defer p.Conn.Close()

	// 2. Extended Handshake
	extHandshake := ut_metadata.NewHandshake(0)
	payload, err := ut_metadata.SerializeHandshake(extHandshake)
	if err != nil {
		return torrent.InfoDictionary{}, err
	}

	finalPayload := make([]byte, 1+len(payload))
	finalPayload[0] = ut_metadata.ExtensionHandshakeID
	copy(finalPayload[1:], payload)

	msg := &peer.Message{ID: peer.MsgExtended, Payload: finalPayload}
	if err := p.SendMessage(msg); err != nil {
		return torrent.InfoDictionary{}, err
	}

	// 3. Receive Extended Handshake & Request Metadata
	var utMetadataID int

	for i := 0; i < 20; i++ { // Increase attempts to 20
		msg, err := p.ReadMessage()
		if err != nil {
			return torrent.InfoDictionary{}, err
		}
		if msg == nil {
			continue
		}

		if msg.ID == peer.MsgExtended {
			extID := msg.Payload[0]
			if extID == ut_metadata.ExtensionHandshakeID {
				// We assume ut_metadata ID is 1 for now (simplified)
				utMetadataID = 1

				// Request Piece 0
				var buf bytes.Buffer
				reqMap := map[string]interface{}{
					"msg_type": 0,
					"piece":    0,
				}
				bencode.Encode(&buf, reqMap)
				reqPayload := ut_metadata.FormatExtendedMessage(uint8(utMetadataID), buf.Bytes())
				p.SendMessage(&peer.Message{ID: peer.MsgExtended, Payload: reqPayload})

			} else if int(extID) == ut_metadata.MetadataID || int(extID) == 1 {
				// Data!
				dataPayload := msg.Payload[1:]
				metaMsg, data, err := ut_metadata.ParseMetadataMessage(dataPayload)
				if err != nil {
					continue
				}

				if metaMsg.MsgType == ut_metadata.MsgTypeData {
					log.Println("GOT METADATA BYTES! Size:", len(data))

					// Decode Bencoded Info Dictionary
					r := bytes.NewReader(data)
					val, err := bencode.Decode(r)
					if err != nil {
						return torrent.InfoDictionary{}, fmt.Errorf("failed to decode metadata info: %v", err)
					}

					infoMap, ok := val.(map[string]interface{})
					if !ok {
						return torrent.InfoDictionary{}, fmt.Errorf("decoded metadata is not a dictionary")
					}

					// Parse into InfoDictionary struct
					infoDict, err := torrent.ParseInfo(infoMap)
					if err != nil {
						return torrent.InfoDictionary{}, fmt.Errorf("failed to parse into InfoDictionary: %v", err)
					}

					log.Printf("Parsed InfoDictionary: Name=%s, Length=%d, PiecesCount=%d", infoDict.Name, infoDict.Length, len(infoDict.Pieces)/20)
					return infoDict, nil
				}
			}
		}
	}

	return torrent.InfoDictionary{}, fmt.Errorf("metadata not found in first 20 messages")
}

func (mc *MagnetClient) announceToTrackers(infoHash [20]byte) {
	// Dummy TorrentFile for tracker compatibility
	tf := &torrent.TorrentFile{
		InfoHash: infoHash,
		Info:     torrent.InfoDictionary{Length: 0}, // Unknown length
	}

	for _, tr := range mc.Client.Trackers {
		go func(trURL string) {
			// Determine scheme
			u, err := url.Parse(trURL)
			if err != nil {
				return
			}

			var peers []tracker.Peer

			if strings.HasPrefix(u.Scheme, "udp") {
				udpClient := tracker.NewUDPClient()
				udpClient.Timeout = 5 * time.Second
				peers, err = udpClient.AnnounceUDP(trURL, tf, string(mc.Client.PeerID[:]), 6881)
			} else if strings.HasPrefix(u.Scheme, "http") {
				// Build URL manually or use helper? Helper expects tf.Announce to be set.
				tf.Announce = trURL
				fullURL, _ := tracker.BuildTrackerURL(tf, string(mc.Client.PeerID[:]), 6881, 0, 0, 0)
				httpClient := tracker.NewClient()
				httpClient.Timeout = 5 * time.Second
				peers, err = httpClient.AnnounceHTTP(fullURL)
			}

			if err != nil {
				mc.Client.AddLog(fmt.Sprintf("Tracker error %s: %v", u.Host, err))
				return
			}

			if len(peers) > 0 {
				mc.Client.AddLog(fmt.Sprintf("Tracker %s found %d peers", u.Host, len(peers)))
				// Feed to DHT PeersFound channel (reusing the discovery pipe)
				for _, p := range peers {
					mc.DHT.PeersFound <- fmt.Sprintf("%s:%d", p.IP.String(), p.Port)
				}
			}
		}(tr)
	}
}
