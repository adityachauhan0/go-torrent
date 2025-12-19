package client

import (
	"log"
	"math/rand"
	"net/url"

	"github.com/valhalla/go-torrent/pkg/torrent"
	"github.com/valhalla/go-torrent/pkg/tracker"
)

type Client struct {
	TorrentPath string
	OutputDir   string
	PeerID      [20]byte
}

func New(torrentPath, outputDir string) (*Client, error) {
	var peerID [20]byte
	rand.Read(peerID[:]) // Random peer ID for now
	
	return &Client{
		TorrentPath: torrentPath,
		OutputDir:   outputDir,
		PeerID:      peerID,
	}, nil
}

func (c *Client) Run() error {
	// Parse torrent file
	tf, err := torrent.Parse(c.TorrentPath)
	if err != nil {
		return err
	}
	log.Printf("Parsed torrent: %s (InfoHash: %x)", tf.Info.Name, tf.InfoHash)

	// Parse announce URL to check scheme
	announceURL, err := url.Parse(tf.Announce)
	if err != nil {
		return err
	}

	var peersInfo []tracker.Peer

	if announceURL.Scheme == "udp" {
		log.Printf("Using UDP tracker: %s", tf.Announce)
		udpClient := tracker.NewUDPClient()
		peersInfo, err = udpClient.AnnounceUDP(tf.Announce, tf, string(c.PeerID[:]), 6881)
	} else {
		log.Printf("Using HTTP tracker: %s", tf.Announce)
		// Build HTTP Tracker URL
		// Left needs calculation (total size - downloaded)
		tURL, err := tracker.BuildTrackerURL(tf, string(c.PeerID[:]), 6881, 0, 0, tf.Info.Length) 
		if err != nil {
			return err
		}
		
		httpClient := tracker.NewClient()
		peersInfo, err = httpClient.AnnounceHTTP(tURL)
	}

	if err != nil {
		log.Printf("Error getting peers from tracker: %v", err)
	}
	log.Printf("Found %d peers", len(peersInfo))

	// Start Download
	d := NewDownloader(tf, c.PeerID, c.OutputDir)
	go d.ProcessResults()
	d.Start(peersInfo)
	
	select {}
}
