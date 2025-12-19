package bencode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// Decode reads bencoded data from the reader and returns the Go representation.
// Supported types:
// string -> string
// integer -> int
// list -> []interface{}
// dict -> map[string]interface{}
func Decode(r io.Reader) (interface{}, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	b, err := br.Peek(1)
	if err != nil {
		return nil, err
	}

	switch {
	case b[0] == 'i':
		return decodeInt(br)
	case b[0] == 'l':
		return decodeList(br)
	case b[0] == 'd':
		return decodeDict(br)
	case b[0] >= '0' && b[0] <= '9':
		return decodeString(br)
	default:
		return nil, fmt.Errorf("invalid bencode prefix: %v", b[0])
	}
}

func decodeInt(br *bufio.Reader) (int, error) {
	// Consumes 'i'
	_, err := br.ReadByte()
	if err != nil {
		return 0, err
	}

	// Read until 'e'
	bytes, err := br.ReadBytes('e')
	if err != nil {
		return 0, err
	}

	// Remove 'e'
	strVal := string(bytes[:len(bytes)-1])
	
	if strVal == "-0" {
		return 0, errors.New("invalid integer: -0")
	}
	if len(strVal) > 1 && strVal[0] == '0' {
		return 0, errors.New("invalid integer: leading zero")
	}

	return strconv.Atoi(strVal)
}

func decodeString(br *bufio.Reader) (string, error) {
	// Read length until ':'
	lenBytes, err := br.ReadBytes(':')
	if err != nil {
		return "", err
	}

	strLen := string(lenBytes[:len(lenBytes)-1])
	length, err := strconv.Atoi(strLen)
	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", errors.New("negative string length")
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(br, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func decodeList(br *bufio.Reader) ([]interface{}, error) {
	// Consume 'l'
	_, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	list := make([]interface{}, 0)

	for {
		b, err := br.Peek(1)
		if err != nil {
			return nil, err
		}

		if b[0] == 'e' {
			// consume 'e'
			br.ReadByte()
			return list, nil
		}

		val, err := Decode(br)
		if err != nil {
			return nil, err
		}
		list = append(list, val)
	}
}

func decodeDict(br *bufio.Reader) (map[string]interface{}, error) {
	// Consume 'd'
	_, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	dict := make(map[string]interface{})

	for {
		b, err := br.Peek(1)
		if err != nil {
			return nil, err
		}

		if b[0] == 'e' {
			br.ReadByte()
			return dict, nil
		}

		// Keys must be strings
		keyVal, err := Decode(br)
		if err != nil {
			return nil, err
		}

		key, ok := keyVal.(string)
		if !ok {
			return nil, fmt.Errorf("dict key must be string, got %T", keyVal)
		}

		val, err := Decode(br)
		if err != nil {
			return nil, err
		}

		dict[key] = val
	}
}

// Encode writes the bencoded representation of data to w.
func Encode(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case string:
		fmt.Fprintf(w, "%d:%s", len(v), v)
	case int:
		fmt.Fprintf(w, "i%de", v)
	case []interface{}:
		fmt.Fprint(w, "l")
		for _, item := range v {
			if err := Encode(w, item); err != nil {
				return err
			}
		}
		fmt.Fprint(w, "e")
	case map[string]interface{}:
		fmt.Fprint(w, "d")
		// Keys must be sorted strings
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		for _, k := range keys {
			if err := Encode(w, k); err != nil { // encode key
				return err
			}
			if err := Encode(w, v[k]); err != nil { // encode value
				return err
			}
		}
		fmt.Fprint(w, "e")
	default:
		return fmt.Errorf("unsupported type: %T", data)
	}
	return nil
}
