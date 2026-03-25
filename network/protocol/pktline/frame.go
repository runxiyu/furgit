package pktline

// Frame is one decoded pkt-line frame.
//
// For PacketData, Payload holds frame bytes (possibly empty for 0004).
// For control frames, Payload is nil.
type Frame struct {
	Type    PacketType
	Payload []byte
}
