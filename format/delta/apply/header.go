package apply

import (
	"fmt"
	"io"
)

// ReadHeaderSizes reads the first two varints in one inflated delta stream.
func ReadHeaderSizes(reader io.Reader) (int, int, error) {
	// Two Git varints are read here. Each can take up to 10 bytes.
	var buf [20]byte
	n := 0

	for {
		if n >= len(buf) {
			return 0, 0, fmt.Errorf("format/delta/apply: malformed delta varint")
		}
		if _, err := io.ReadFull(reader, buf[n:n+1]); err != nil {
			return 0, 0, fmt.Errorf("format/delta/apply: malformed delta varint: %w", err)
		}
		n++
		if buf[n-1]&0x80 == 0 {
			break
		}
	}
	pos := 0
	srcSize, err := readVarint(buf[:n], &pos)
	if err != nil {
		return 0, 0, err
	}

	for {
		if n >= len(buf) {
			return 0, 0, fmt.Errorf("format/delta/apply: malformed delta varint")
		}
		if _, err := io.ReadFull(reader, buf[n:n+1]); err != nil {
			return 0, 0, fmt.Errorf("format/delta/apply: malformed delta varint: %w", err)
		}
		n++
		if buf[n-1]&0x80 == 0 {
			break
		}
	}
	dstSize, err := readVarint(buf[:n], &pos)
	if err != nil {
		return 0, 0, err
	}
	return srcSize, dstSize, nil
}
