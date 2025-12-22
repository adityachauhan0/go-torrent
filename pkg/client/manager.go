package client

import (
	"sync"
)

type TorrentStats struct {
	Hash       string   `json:"hash"`
	Name       string   `json:"name"`
	State      string   `json:"state"` // "Downloading", "Seeding", "Paused"
	Progress   float64  `json:"progress"`
	SpeedIn    float64  `json:"speed_in"`
	SpeedOut   float64  `json:"speed_out"`
	Peers      int      `json:"peers"`
	Downloaded int64    `json:"downloaded"`
	Logs       []string `json:"logs"`
}

type Manager struct {
	Clients map[string]*Client
	mu      sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		Clients: make(map[string]*Client),
	}
}

func (m *Manager) Add(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Map by InfoHash hex string
	infoHash := c.InfoHash
	m.Clients[infoHash] = c
}

func (m *Manager) Remove(hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Stop the client first
	delete(m.Clients, hash)
}

func (m *Manager) GetTorrents() []TorrentStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := make([]TorrentStats, 0, len(m.Clients))
	for hash, c := range m.Clients {
		s := TorrentStats{
			Hash:       hash,
			Name:       c.Name,
			State:      c.Status, // Real granular status
			Progress:   c.Progress(),
			SpeedIn:    c.DownloadSpeed(),
			Peers:      c.PeerCount(),
			Downloaded: c.Downloaded(),
			Logs:       c.GetLogs(),
		}
		if s.State == "" {
			s.State = "Downloading"
		}
		if s.Progress >= 100 {
			s.State = "Seeding"
		}
		stats = append(stats, s)
	}
	return stats
}

func (m *Manager) GetGlobalStats() (float64, float64, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var totalDown, totalUp float64
	var active int

	for _, c := range m.Clients {
		totalDown += c.DownloadSpeed()
		totalUp += c.UploadSpeed() // Placeholder
		active++
	}
	return totalDown, totalUp, active
}
