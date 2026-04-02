package intconv

// SignExtendByteToUint32 sign-extends b as a signed 8-bit integer into uint32.
func SignExtendByteToUint32(b byte) uint32 {
	if b&0x80 == 0 {
		return uint32(b)
	}

	return 0xFFFFFF00 | uint32(b)
}
