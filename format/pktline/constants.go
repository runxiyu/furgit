package pktline

const (
	// DefaultPacketMax is a conservative packet size commonly used by
	// line-oriented protocol messages.
	DefaultPacketMax = 1000
	// LargePacketMax is the maximum on-wire packet size including the
	// 4-byte hexadecimal length header.
	LargePacketMax = 65520
	// LargePacketDataMax is the maximum payload size in one packet.
	LargePacketDataMax = LargePacketMax - 4
)
