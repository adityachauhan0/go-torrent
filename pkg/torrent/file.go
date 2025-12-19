package torrent

// SingleFile represents a file in a single-file torrent or a file inside a multi-file torrent.
type FileInfo struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
}

// InfoDictionary represents the 'info' dictionary in the torrent file.
type InfoDictionary struct {
	PieceLength int64      `bencode:"piece length"`
	Pieces      string     `bencode:"pieces"`
	Private     int        `bencode:"private,omitempty"` // Optional
	Name        string     `bencode:"name"`
	Length      int64      `bencode:"length,omitempty"` // Only for single file
	Files       []FileInfo `bencode:"files,omitempty"`  // Only for multi file
}

// TorrentFile represents the top-level structure of a .torrent file.
type TorrentFile struct {
	Announce     string         `bencode:"announce"`
	AnnounceList [][]string     `bencode:"announce-list,omitempty"` // BEP 12: Multitracker
	CreationDate int64          `bencode:"creation date,omitempty"`
	Comment      string         `bencode:"comment,omitempty"`
	CreatedBy    string         `bencode:"created by,omitempty"`
	Info         InfoDictionary `bencode:"info"`
	InfoHash     [20]byte       `bencode:"-"` // Calculated, not present in file
}
