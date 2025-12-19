package torrent

import (
	"crypto/sha1"
	"io"
	"os"

	"github.com/valhalla/go-torrent/pkg/bencode"
)

// Parse opens a .torrent file and parses it into a TorrentFile struct.
func Parse(path string) (*TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseReader(file)
}

// ParseReader parses a .torrent file from a reader.
func ParseReader(r io.Reader) (*TorrentFile, error) {
	// We need to read the whole content to hash the info dictionary later,
	// or we can decode and then re-encode the info dictionary to hash it.
	// For simplicity and correctness, let's decode the generic map first to extract raw info bytes if possible,
	// OR, we can just use our specialized decoder if we add struct tags support.
	
	// Since our simple bencode package (currently) returns map[string]interface{} for dicts,
	// we have two options:
	// 1. Use reflection to map the map[string]interface{} to the struct (like encoding/json).
	// 2. Manual mapping.
	// 3. Update bencode package to support Unmarshal into structs.
	
	// Given the task limitation, option 2 is tedious but robust. Option 3 is best.
	// Let's rely on the generic Decode and then map manually for now, 
	// or better, let's IMPROVE the bencode package to support `Unmarshal` in the next step.
	// For this specific step, I will stick to manual mapping to keep it simple and explicit,
	// as writing a full reflection-based Unmarshaler might be too much code for one turn.
	
	val, err := bencode.Decode(r)
	if err != nil {
		return nil, err
	}

	root, ok := val.(map[string]interface{})
	if !ok {
		return nil, os.ErrInvalid
	}

	t := &TorrentFile{}
	
	if announce, ok := root["announce"].(string); ok {
		t.Announce = announce
	}
	
	// Handle optional fields... (omitted for brevity in initial pass, can add later)

	infoMap, ok := root["info"].(map[string]interface{})
	if !ok {
		return nil, os.ErrInvalid
	}

	t.Info, err = parseInfo(infoMap)
	if err != nil {
		return nil, err
	}
	
	// Calculate InfoHash
	// For this, we need the exact bencoded bytes of the info dictionary.
	// Since we decoded everything, we don't have the raw bytes easily unless we re-encode.
	// Our bencode.Encode function sorts keys, so it should be deterministic and correct for hashing
	// IF the map contained exactly what was in the file.
	sha := sha1.New()
	err = bencode.Encode(sha, infoMap)
	if err != nil {
		return nil, err
	}
	copy(t.InfoHash[:], sha.Sum(nil))

	return t, nil
}

func parseInfo(m map[string]interface{}) (InfoDictionary, error) {
	info := InfoDictionary{}
	
	if name, ok := m["name"].(string); ok {
		info.Name = name
	}
	
	if pl, ok := m["piece length"].(int); ok {
		info.PieceLength = int64(pl)
	}
	
	if pieces, ok := m["pieces"].(string); ok {
		info.Pieces = pieces
	}
	
	if length, ok := m["length"].(int); ok {
		info.Length = int64(length)
	}
	
	// Multi-file parsing
	if files, ok := m["files"].([]interface{}); ok {
		for _, f := range files {
			fMap, ok := f.(map[string]interface{})
			if !ok {
				continue
			}
			fileInfo := FileInfo{}
			if l, ok := fMap["length"].(int); ok {
				fileInfo.Length = int64(l)
			}
			if p, ok := fMap["path"].([]interface{}); ok {
				for _, pPart := range p {
					if s, ok := pPart.(string); ok {
						fileInfo.Path = append(fileInfo.Path, s)
					}
				}
			}
			info.Files = append(info.Files, fileInfo)
		}
	}
	
	return info, nil
}
