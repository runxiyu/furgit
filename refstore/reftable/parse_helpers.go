package reftable

import "fmt"

// readUint24 reads a 24-bit big-endian unsigned integer.
func readUint24(b []byte) uint32 {
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// alignUp rounds pos up to the next multiple of blockSize.
func alignUp(pos, blockSize int) int {
	rem := pos % blockSize
	if rem == 0 {
		return pos
	}

	return pos + (blockSize - rem)
}

// readVarint decodes one reftable/ofs-delta style varint.
func readVarint(buf []byte, off, end int) (uint64, int, error) {
	if off >= end {
		return 0, 0, fmt.Errorf("unexpected EOF")
	}

	b := buf[off]
	val := uint64(b & 0x7f)

	off++
	for b&0x80 != 0 {
		if off >= end {
			return 0, 0, fmt.Errorf("unexpected EOF")
		}

		b = buf[off]
		off++
		val = ((val + 1) << 7) | uint64(b&0x7f)
	}

	return val, off, nil
}
