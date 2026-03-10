package server

import "codeberg.org/lindenii/furgit/protocol/pktline"

// FrameType identifies one low-level v0/v1 server pkt-line frame type.
type FrameType = pktline.PacketType

const (
	// FrameData is one data pkt-line.
	FrameData = pktline.PacketData
	// FrameFlush is one flush-pkt.
	FrameFlush = pktline.PacketFlush
	// FrameDelim is one delim-pkt.
	FrameDelim = pktline.PacketDelim
	// FrameResponseEnd is one response-end-pkt.
	FrameResponseEnd = pktline.PacketResponseEnd
)

// Frame is one decoded low-level pkt-line frame.
type Frame = pktline.Frame
