package packed

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objecttype"
)

// deltaResolveContent resolves one object's content bytes from its pack location.
func (store *Store) deltaResolveContent(start location) (objecttype.Type, []byte, error) {
	plan, err := store.deltaPlanFor(start)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	baseType, out, err := store.deltaResolveBase(plan)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}
	for i := len(plan.frames) - 1; i >= 0; i-- {
		frame := plan.frames[i]
		pack, err := store.openPack(frame.packName)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
		delta, err := inflateAt(pack, frame.dataOffset, -1)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
		out, err = applyDelta(out, delta)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
	}
	if int64(len(out)) != plan.declaredSize {
		return objecttype.TypeInvalid, nil, fmt.Errorf(
			"objectstore/packed: resolved content size mismatch: got %d want %d",
			len(out),
			plan.declaredSize,
		)
	}
	return baseType, out, nil
}

// applyDelta applies one Git delta instruction stream to base.
func applyDelta(base, delta []byte) ([]byte, error) {
	pos := 0
	srcSize, err := readDeltaVarint(delta, &pos)
	if err != nil {
		return nil, err
	}
	dstSize, err := readDeltaVarint(delta, &pos)
	if err != nil {
		return nil, err
	}
	if srcSize != len(base) {
		return nil, fmt.Errorf("objectstore/packed: delta source size mismatch: got %d want %d", srcSize, len(base))
	}

	out := make([]byte, dstSize)
	outPos := 0
	for pos < len(delta) {
		op := delta[pos]
		pos++
		if op&0x80 != 0 {
			off := 0
			if op&0x01 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy offset")
				}
				off |= int(delta[pos])
				pos++
			}
			if op&0x02 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy offset")
				}
				off |= int(delta[pos]) << 8
				pos++
			}
			if op&0x04 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy offset")
				}
				off |= int(delta[pos]) << 16
				pos++
			}
			if op&0x08 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy offset")
				}
				off |= int(delta[pos]) << 24
				pos++
			}

			n := 0
			if op&0x10 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy size")
				}
				n |= int(delta[pos])
				pos++
			}
			if op&0x20 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy size")
				}
				n |= int(delta[pos]) << 8
				pos++
			}
			if op&0x40 != 0 {
				if pos >= len(delta) {
					return nil, fmt.Errorf("objectstore/packed: malformed delta copy size")
				}
				n |= int(delta[pos]) << 16
				pos++
			}
			if n == 0 {
				n = 0x10000
			}
			if off < 0 || n < 0 || off+n > len(base) || outPos+n > len(out) {
				return nil, fmt.Errorf("objectstore/packed: delta copy out of bounds")
			}
			copy(out[outPos:outPos+n], base[off:off+n])
			outPos += n
			continue
		}

		if op == 0 {
			return nil, fmt.Errorf("objectstore/packed: invalid delta opcode 0")
		}
		n := int(op)
		if pos+n > len(delta) || outPos+n > len(out) {
			return nil, fmt.Errorf("objectstore/packed: delta insert out of bounds")
		}
		copy(out[outPos:outPos+n], delta[pos:pos+n])
		outPos += n
		pos += n
	}
	if outPos != len(out) {
		return nil, fmt.Errorf("objectstore/packed: delta output size mismatch: got %d want %d", outPos, len(out))
	}
	return out, nil
}

// readDeltaVarint parses one Git delta varint and advances pos.
func readDeltaVarint(buf []byte, pos *int) (int, error) {
	value := 0
	shift := uint(0)
	for {
		if *pos >= len(buf) {
			return 0, fmt.Errorf("objectstore/packed: malformed delta varint")
		}
		b := buf[*pos]
		*pos++
		value |= int(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
		if shift > 63 {
			return 0, fmt.Errorf("objectstore/packed: delta varint overflow")
		}
	}
	return value, nil
}
