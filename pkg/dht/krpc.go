package dht

// KRPC definitions for DHT messages (BEP 5).

const (
	QueryPing      = "ping"
	QueryFindNode  = "find_node"
	QueryGetPeers  = "get_peers"
	QueryAnnounce  = "announce_peer"
)

// Msg represents a KRPC message.
type Msg struct {
	T []byte                 `bencode:"t"`           // Transaction ID
	Y string                 `bencode:"y"`           // Message Type: "q" (query), "r" (response), "e" (error)
	Q string                 `bencode:"q,omitempty"` // Query Method (only for queries)
	A map[string]interface{} `bencode:"a,omitempty"` // Arguments (only for queries)
	R map[string]interface{} `bencode:"r,omitempty"` // Return Values (only for responses)
	E []interface{}          `bencode:"e,omitempty"` // Error ([code, description])
}

// NewQuery creates a new query message.
func NewQuery(tid string, method string, args map[string]interface{}) Msg {
	return Msg{
		T: []byte(tid),
		Y: "q",
		Q: method,
		A: args,
	}
}

// NewResponse creates a new response message.
func NewResponse(tid string, values map[string]interface{}) Msg {
	return Msg{
		T: []byte(tid),
		Y: "r",
		R: values,
	}
}

// NewError creates a new error message.
func NewError(tid string, code int, msg string) Msg {
	return Msg{
		T: []byte(tid),
		Y: "e",
		E: []interface{}{code, msg},
	}
}

// Note: Serialization/Deserialization will be handled by the pkg/bencode package.
// We just need to ensure the bencode package can handle these structs constants.
// Since our bencode might be limited, we rely on map-based interaction as seen in ParseReader.
