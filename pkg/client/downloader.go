package client

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/valhalla/go-torrent/pkg/bitfield"
	"github.com/valhalla/go-torrent/pkg/peer"
	"github.com/valhalla/go-torrent/pkg/torrent"
	"github.com/valhalla/go-torrent/pkg/tracker"
)

type Downloader struct {
	Torrent     *torrent.TorrentFile
	PeerID      [20]byte
	Output      string
	Peers       []peer.Peer
	ActivePeers []*peer.Peer
	PeersMu     sync.Mutex // Protect ActivePeers
	BitfieldMu  sync.Mutex // Protect Bitfield
	WorkQueue   chan int
	Results     chan *PieceResult
	Bitfield    bitfield.Bitfield

	// Stats
	Downloaded int64
	Uploaded   int64
	Speed      float64 // Bps
}

func (d *Downloader) BroadcastHave(index int) {
	d.PeersMu.Lock()
	defer d.PeersMu.Unlock()

	payload := peer.FormatHave(index)
	msg := &peer.Message{ID: peer.MsgHave, Payload: payload}

	for _, p := range d.ActivePeers {
		go p.SendMessage(msg) // Send async to not block
	}
}

type PieceResult struct {
	Index int
	Buf   []byte
	Err   error
}

func NewDownloader(tf *torrent.TorrentFile, peerID [20]byte, output string) *Downloader {
	d := &Downloader{
		Torrent:   tf,
		PeerID:    peerID,
		Output:    output,
		WorkQueue: make(chan int, len(tf.Info.Pieces)/20), // Each piece hash is 20 bytes
		Results:   make(chan *PieceResult),
	}

	// Initialize bitfield logic
	numPieces := len(tf.Info.Pieces) / 20
	byteSize := (numPieces + 7) / 8
	d.Bitfield = make(bitfield.Bitfield, byteSize)

	for i := 0; i < numPieces; i++ {
		d.WorkQueue <- i
	}

	return d
}

// Start spawns workers for peers.
func (d *Downloader) Start(peers []tracker.Peer) {
	go d.TrackSpeed() // Start speed tracker
	for _, p := range peers {
		go d.downloadWorker(p.IP, p.Port)
	}
}

func (d *Downloader) AddPeer(p tracker.Peer) {
	go d.downloadWorker(p.IP, p.Port)
}

const BlockSize = 16384

func (d *Downloader) downloadWorker(ip net.IP, port uint16) {
	p := peer.NewPeer(ip, port, [20]byte{})
	err := p.Connect(d.Torrent.InfoHash, d.PeerID, d.Bitfield)
	if err != nil {
		log.Printf("Failed to handshake with %s: %v", ip, err)
		return
	}
	defer p.Conn.Close()

	// Register peer
	d.PeersMu.Lock()
	d.ActivePeers = append(d.ActivePeers, p)
	d.PeersMu.Unlock()

	// Deregister on exit (optional but good)

	p.SendUnchoke()
	p.SendInterested()

	for index := range d.WorkQueue {
		// Ensure peer has piece
		// if !p.Bitfield.HasPiece(index) {
		// 	d.WorkQueue <- index
		// 	continue
		// }

		// Calculate piece length
		length := d.Torrent.Info.PieceLength
		// Last piece might be shorter
		if index == len(d.Torrent.Info.Pieces)/20-1 {
			length = d.Torrent.Info.Length % d.Torrent.Info.PieceLength
			if length == 0 {
				length = d.Torrent.Info.PieceLength
			}
		}

		buf := make([]byte, length)
		downloaded := 0
		requested := 0
		backlog := 0 // Pipelining depth

		for downloaded < int(length) {
			if !p.Choked {
				for backlog < 5 && requested < int(length) {
					blockSize := BlockSize
					if requested+blockSize > int(length) {
						blockSize = int(length) - requested
					}

					err := p.SendRequest(index, requested, blockSize)
					if err != nil {
						d.WorkQueue <- index
						return
					}
					backlog++
					requested += blockSize
				}
			}

			msg, err := p.ReadMessage()
			if err != nil {
				d.WorkQueue <- index
				return
			}

			if msg == nil { // keep-alive
				continue
			}

			switch msg.ID {
			case peer.MsgUnchoke:
				p.Choked = false
			case peer.MsgChoke:
				p.Choked = true
			case peer.MsgHave:
				// Update bitfield
				idx, _ := peer.ParseHave(msg.Payload)
				p.Bitfield.SetPiece(idx)
			case peer.MsgPiece:
				n, data, err := peer.ParsePiece(index, msg.Payload, *msg)
				if err != nil {
					log.Printf("Error parsing piece: %v", err)
					continue
				}
				copy(buf[n:], data)
				downloaded += len(data)

				// Track stats
				d.PeersMu.Lock()
				d.Downloaded += int64(len(data))
				d.PeersMu.Unlock()

				backlog--
			case peer.MsgRequest:
				d.HandleRequest(p, msg)
			}
		}

		// Integrity check (SHA1) would go here

		d.Results <- &PieceResult{Index: index, Buf: buf}
	}
}

func (d *Downloader) TrackSpeed() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastDownloaded int64
	for range ticker.C {
		d.PeersMu.Lock()
		current := d.Downloaded
		diff := current - lastDownloaded
		d.Speed = float64(diff) / 2.0 // bps
		lastDownloaded = current
		d.PeersMu.Unlock()
	}
}
