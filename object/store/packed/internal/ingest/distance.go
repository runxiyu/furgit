package ingest

import (
	"fmt"
	"io"
)

// readOfsDistanceFromStream reads one ofs-delta encoded distance.
func readOfsDistanceFromStream(reader io.ByteReader) (uint64, int, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return 0, 0, fmt.Errorf("read ofs distance first byte: %w", err)
	}

	dist := uint64(first & 0x7f)
	consumed := 1

	b := first
	for b&0x80 != 0 {
		b, err = reader.ReadByte()
		if err != nil {
			return 0, 0, fmt.Errorf("read ofs distance continuation: %w", err)
		}

		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}

	return dist, consumed, nil
}
