package ingest

import (
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objecttype"
)

// encodePackEntryHeader encodes one non-delta packed entry header.
func encodePackEntryHeader(ty objecttype.Type, size int64) []byte {
	var out [16]byte

	n := 0

	s, err := intconv.Int64ToUint64(size)
	if err != nil {
		panic(err)
	}

	c := (uint8(ty) << 4) | byte(s&0x0f)

	s >>= 4
	for s != 0 {
		out[n] = c | 0x80
		n++
		c = byte(s & 0x7f)
		s >>= 7
	}

	out[n] = c
	n++

	return append([]byte(nil), out[:n]...)
}
