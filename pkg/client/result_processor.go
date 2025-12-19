package client

import (
	"log"
)

// ProcessResults listens for completed pieces, writes them to disk, and broadcasts Have.
func (d *Downloader) ProcessResults() {
	fw := NewFileWriter(d.Output)
	
	for res := range d.Results {
		if res.Err != nil {
			log.Printf("Piece %d failed: %v", res.Index, res.Err)
			d.WorkQueue <- res.Index // Retry
			continue
		}
		
		// Write to disk
		// Calculate offset
		offset := int64(res.Index) * d.Torrent.Info.PieceLength
		
		// Assuming single file for simple writer usage again
		// If multi-file, we need to handle file boundaries.
		// Reusing straightforward single file logic for MVP extension
		filename := d.Torrent.Info.Name
		err := fw.Write(filename, offset, res.Buf)
		if err != nil {
			log.Printf("Failed to write piece %d: %v", res.Index, err)
			continue // Critical error really
		}
		
		log.Printf("Downloaded piece %d", res.Index)
		
		// Update Bitfield
		d.BitfieldMu.Lock()
		d.Bitfield.SetPiece(res.Index)
		d.BitfieldMu.Unlock()
		
		// Broadcast Have
		d.BroadcastHave(res.Index)
	}
}
