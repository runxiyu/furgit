// Package apply applies Git delta instruction streams.
package apply

import "fmt"

// Apply applies one Git delta instruction stream to base.
func Apply(base, delta []byte) ([]byte, error) {
	pos := 0

	srcSize, err := readVarint(delta, &pos)
	if err != nil {
		return nil, err
	}

	dstSize, err := readVarint(delta, &pos)
	if err != nil {
		return nil, err
	}

	if srcSize != len(base) {
		return nil, fmt.Errorf("delta/apply: delta source size mismatch: got %d want %d", srcSize, len(base))
	}

	out := make([]byte, dstSize)
	outPos := 0

	for pos < len(delta) {
		op := delta[pos]
		pos++

		//nolint:nestif
		if op&0x80 != 0 {
			off := 0

			if op&0x01 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy offset")
				}

				off |= int(delta[pos])
				pos++
			}

			if op&0x02 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy offset")
				}

				off |= int(delta[pos]) << 8
				pos++
			}

			if op&0x04 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy offset")
				}

				off |= int(delta[pos]) << 16
				pos++
			}

			if op&0x08 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy offset")
				}

				off |= int(delta[pos]) << 24
				pos++
			}

			n := 0

			if op&0x10 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy size")
				}

				n |= int(delta[pos])
				pos++
			}

			if op&0x20 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy size")
				}

				n |= int(delta[pos]) << 8
				pos++
			}

			if op&0x40 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("delta/apply: malformed delta copy size")
				}

				n |= int(delta[pos]) << 16
				pos++
			}

			if n == 0 {
				n = 0x10000
			}

			if off < 0 || n < 0 || off+n > len(base) || outPos+n > len(out) {
				return nil, fmt.Errorf("delta/apply: delta copy out of bounds")
			}

			copy(out[outPos:outPos+n], base[off:off+n])
			outPos += n

			continue
		}

		if op == 0 {
			return nil, fmt.Errorf("delta/apply: invalid delta opcode 0")
		}

		n := int(op)
		if pos+n > len(delta) || outPos+n > len(out) {
			return nil, fmt.Errorf("delta/apply: delta insert out of bounds")
		}

		copy(out[outPos:outPos+n], delta[pos:pos+n])
		outPos += n
		pos += n
	}

	if outPos != len(out) {
		return nil, fmt.Errorf("delta/apply: delta output size mismatch: got %d want %d", outPos, len(out))
	}

	return out, nil
}

// readVarint parses one Git delta varint and advances pos.
func readVarint(buf []byte, pos *int) (int, error) {
	value := 0
	shift := uint(0)

	for {
		if *pos >= len(buf) {
			return 0, fmt.Errorf("delta/apply: malformed delta varint")
		}

		b := buf[*pos]
		*pos++

		value |= int(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}

		shift += 7
		if shift > 63 {
			return 0, fmt.Errorf("delta/apply: delta varint overflow")
		}
	}

	return value, nil
}
