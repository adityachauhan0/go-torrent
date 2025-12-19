package client

import (
	"log"

	"github.com/valhalla/go-torrent/pkg/peer"
)

// HandleRequest processes a peer's request for a block.
func (d *Downloader) HandleRequest(p *peer.Peer, msg *peer.Message) {
	index, begin, length, err := peer.ParseRequest(*msg)
	if err != nil {
		log.Printf("Failed to parse request: %v", err)
		return
	}

	// Validation: Check if we have the piece
	d.BitfieldMu.Lock()
	has := d.Bitfield.HasPiece(index)
	d.BitfieldMu.Unlock()
	
	if !has {
		// We don't have it, ignore or send choke?
		// BEP 3 says just ignore if choked, but if unchoked and don't have it, maybe close or ignore.
		return 
	}
	
	if length > 131072 { // Sanity check (usually 16KB requests)
		log.Printf("Request too large: %d", length)
		return
	}

	// Read from disk
	// We need to map piece index to file(s).
	// For single file torrents, it's easy.
	// For multi-file, it's complex.
	// Assuming `FileWriter` handles "Piece Index -> File mapping" OR we do it here.
	// The current FileWriter takes a `path`. We need to abstract this.
	
	// Refactor needed: FileWriter should probably read by "Piece Index + Offset" or "Global Offset".
	// But `FileWriter` currently takes `path`.
	
	// Let's implement a helper in Downloader or FileWriter to read by Piece Index.
	// For MVP/Single File (Alice/Alpine), we can assume single file?
	// Alpine is single file (tar.gz inside torrent?). 
	// Wait, internal structure of Downloader needs to know global mapping.
	
	// TODO: Implement ReadPiece in Downloader
	buf, err := d.ReadPiece(index, begin, length)
	if err != nil {
		log.Printf("Failed to read piece %d: %v", index, err)
		return
	}

	// Send Piece
	payload := peer.FormatPiece(index, begin, buf)
	p.SendMessage(&peer.Message{ID: peer.MsgPiece, Payload: payload})
}

// ReadPiece reads a block from the storage.
func (d *Downloader) ReadPiece(index, begin, length int) ([]byte, error) {
	// Simple Single File implementation for now
	if len(d.Torrent.Info.Files) == 0 {
		// Single file
		// Global offset = index * pieceLength + begin
		offset := int64(index)*d.Torrent.Info.PieceLength + int64(begin)
		filename := d.Torrent.Info.Name
		
		fw := NewFileWriter(d.Output) // Should reuse existing instance ideally
		return fw.Read(filename, offset, length)
	}
	
	// Multi-file not implemented for Read yet
	return nil, nil // Error
}
