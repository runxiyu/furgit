package sideband64k

import "codeberg.org/lindenii/furgit/protocol/pktline"

const (
	// PacketMax is the maximum on-wire pkt-line size used by side-band-64k.
	PacketMax = pktline.LargePacketMax
	// DataMax is the maximum sideband payload size excluding the 1-byte band designator.
	DataMax = pktline.LargePacketDataMax - 1
)
