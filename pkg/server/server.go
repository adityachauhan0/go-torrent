package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/valhalla/go-torrent/pkg/client"
	"github.com/valhalla/go-torrent/pkg/torrent"
)

// Server represents the HTTP server for the torrent client.
type Server struct {
	Port    string
	Manager *client.Manager
}

// NewServer creates a new HTTP server.
func NewServer(port string, m *client.Manager) *Server {
	return &Server{
		Port:    port,
		Manager: m,
	}
}

// corsMiddleware adds CORS headers to the response.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			return
		}
		next(w, r)
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	http.HandleFunc("/api/stats", corsMiddleware(s.handleStats))
	http.HandleFunc("/api/torrents", corsMiddleware(s.handleTorrents))
	http.HandleFunc("/api/torrents/add", corsMiddleware(s.handleAddTorrent))
	http.HandleFunc("/api/torrents/upload", corsMiddleware(s.handleUploadTorrent))
	http.HandleFunc("/api/torrents/remove", corsMiddleware(s.handleRemoveTorrent))

	return http.ListenAndServe(":"+s.Port, nil)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	down, up, active := s.Manager.GetGlobalStats()
	stats := map[string]interface{}{
		"download_speed_bps": down,
		"upload_speed_bps":   up,
		"active_torrents":    active,
	}
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleTorrents(w http.ResponseWriter, r *http.Request) {
	stats := s.Manager.GetTorrents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

type AddTorrentRequest struct {
	MagnetLink string `json:"magnet_link"`
}

func (s *Server) handleAddTorrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddTorrentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MagnetLink == "" {
		http.Error(w, "Magnet link is required", http.StatusBadRequest)
		return
	}

	magnet, err := torrent.ParseMagnetURI(req.MagnetLink)
	if err != nil {
		http.Error(w, "Invalid magnet link: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create and start magnet client
	// Use "NewMagnetClient" which we will define, or manually construct.
	c := client.NewMagnetClient(
		fmt.Sprintf("%x", magnet.InfoHash),
		magnet.DisplayName,
		magnet.Trackers,
		".", // Output current dir
	)

	if c.Name == "" {
		c.Name = "Unknown Torrent"
	}

	s.Manager.Add(c)

	// Start it asynchronously
	go func() {
		c.StartMagnet(magnet.Trackers)
	}()

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRemoveTorrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "Hash required", http.StatusBadRequest)
		return
	}

	s.Manager.Remove(hash)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUploadTorrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("torrent_file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file
	// Ensure torrents dir exists (done via mkdir command, but safe to check? skip for now)
	filePath := fmt.Sprintf("./torrents/%s", handler.Filename)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Start Client
	c, err := client.New(filePath, ".")
	if err != nil {
		http.Error(w, "Failed to create client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.Manager.Add(c)

	// Start async
	go func() {
		c.Run()
	}()

	w.WriteHeader(http.StatusOK)
}
