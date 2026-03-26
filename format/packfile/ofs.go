package packfile

import "fmt"

// ParseOfsDeltaDistance parses one ofs-delta backward distance.
func ParseOfsDeltaDistance(buf []byte) (uint64, int, error) {
	if len(buf) == 0 {
		return 0, 0, fmt.Errorf("packfile: malformed ofs-delta distance")
	}

	b := buf[0]
	dist := uint64(b & 0x7f)

	consumed := 1
	for b&0x80 != 0 {
		if consumed >= len(buf) {
			return 0, 0, fmt.Errorf("packfile: malformed ofs-delta distance")
		}

		b = buf[consumed]
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}

	return dist, consumed, nil
}
