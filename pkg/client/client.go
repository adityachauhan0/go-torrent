package client

import (
	"log"
	"math/rand"
	"net/url"
	"sync"

	"github.com/valhalla/go-torrent/pkg/torrent"
	"github.com/valhalla/go-torrent/pkg/tracker"
)

type Client struct {
	TorrentPath string
	OutputDir   string
	PeerID      [20]byte

	// New fields for Manager
	InfoHash   string
	Name       string
	Downloader *Downloader
	Trackers   []string

	// Status and Logs
	Status string
	Logs   []string
	LogMu  sync.Mutex
}

func (c *Client) AddLog(msg string) {
	c.LogMu.Lock()
	defer c.LogMu.Unlock()
	c.Logs = append(c.Logs, msg)
	// Keep logs manageable
	if len(c.Logs) > 100 {
		c.Logs = c.Logs[len(c.Logs)-100:]
	}
}

func (c *Client) GetLogs() []string {
	c.LogMu.Lock()
	defer c.LogMu.Unlock()
	// Return a copy to avoid race on append
	logs := make([]string, len(c.Logs))
	copy(logs, c.Logs)
	return logs
}

func (c *Client) DownloadSpeed() float64 {
	if c.Downloader == nil {
		return 0
	}
	return c.Downloader.Speed
}

func (c *Client) UploadSpeed() float64 {
	return 0 // TODO: Implement upload
}

func (c *Client) Progress() float64 {
	if c.Downloader == nil {
		return 0
	}
	if c.Downloader.Torrent.Info.Length == 0 {
		return 0
	}
	return float64(c.Downloader.Downloaded) / float64(c.Downloader.Torrent.Info.Length) * 100
}

func (c *Client) Downloaded() int64 {
	if c.Downloader == nil {
		return 0
	}
	return c.Downloader.Downloaded
}

func (c *Client) PeerCount() int {
	if c.Downloader == nil {
		return 0
	}
	return len(c.Downloader.ActivePeers)
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

// NewMagnetClient creates a client for a magnet link.
func NewMagnetClient(infoHash string, name string, trackers []string, outputDir string) *Client {
	var peerID [20]byte
	rand.Read(peerID[:])

	return &Client{
		InfoHash:  infoHash,
		Name:      name,
		OutputDir: outputDir,
		PeerID:    peerID,
		Trackers:  trackers,
	}
}

// StartMagnet starts the download process for a magnet link.
func (c *Client) StartMagnet(trackers []string) error {
	mc := NewMagnetClientWrapper(c)
	return mc.Start()
}
