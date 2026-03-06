package pktline

// PacketType identifies the kind of pkt-line frame.
type PacketType uint8

const (
	// PacketData is a regular data frame whose payload is application-defined.
	PacketData PacketType = iota
	// PacketFlush is control frame 0000 and marks end of a message.
	PacketFlush
	// PacketDelim is control frame 0001 and separates sections in protocol v2.
	PacketDelim
	// PacketResponseEnd is control frame 0002 and marks response end on stateless v2 transports.
	PacketResponseEnd
)
