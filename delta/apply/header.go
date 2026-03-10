package apply

import (
	"fmt"
	"io"
)

// ReadHeaderSizes reads the first two varints in one inflated delta stream.
//
// Callers that continue reading the same stream should pass their own buffered
// byte reader and keep using that same reader afterwards.
func ReadHeaderSizes(reader io.ByteReader) (int, int, error) {
	srcSize, err := readVarintFromByteReader(reader)
	if err != nil {
		return 0, 0, err
	}

	dstSize, err := readVarintFromByteReader(reader)
	if err != nil {
		return 0, 0, err
	}

	return srcSize, dstSize, nil
}

// readVarintFromByteReader parses one Git delta varint from reader.
func readVarintFromByteReader(reader io.ByteReader) (int, error) {
	value := 0
	shift := uint(0)

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("delta/apply: malformed delta varint: %w", err)
		}

		value |= int(b&0x7f) << shift
		if b&0x80 == 0 {
			return value, nil
		}

		shift += 7
		if shift > 63 {
			return 0, fmt.Errorf("delta/apply: delta varint overflow")
		}
	}
}
