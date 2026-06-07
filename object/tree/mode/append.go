package mode

import "strconv"

// Append appends the canonical octal encoding of the mode to dst.
func (mode Mode) Append(dst []byte) []byte {
	return strconv.AppendUint(dst, uint64(mode), 8)
}
