package furgit

import (
	"io"

	"codeberg.org/lindenii/furgit/internal/bufpool"
)

func packDeltaReadOfsDistance(data []byte) (uint64, int, error) {
	if len(data) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b := data[0]
	dist := uint64(b & 0x7f)
	consumed := 1
	for (b & 0x80) != 0 {
		if consumed >= len(data) {
			return 0, 0, io.ErrUnexpectedEOF
		}
		b = data[consumed]
		consumed++
		dist += 1
		dist <<= 7
		dist |= uint64(b & 0x7f)
	}
	return dist, consumed, nil
}

func packDeltaApply(base, delta bufpool.Buffer) (bufpool.Buffer, error) {
	pos := 0
	baseBytes := base.Bytes()
	deltaBytes := delta.Bytes()
	srcSize, err := packVarintRead(deltaBytes, &pos)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	dstSize, err := packVarintRead(deltaBytes, &pos)
	if err != nil {
		return bufpool.Buffer{}, err
	}
	if srcSize != len(baseBytes) {
		return bufpool.Buffer{}, ErrInvalidObject
	}
	out := bufpool.Borrow(dstSize)
	out.Resize(dstSize)
	outBytes := out.Bytes()
	outPos := 0

	for pos < len(deltaBytes) {
		op := deltaBytes[pos]
		pos++
		switch {
		case op&0x80 != 0:
			off := 0
			n := 0
			if op&0x01 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos])
				pos++
			}
			if op&0x02 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 8
				pos++
			}
			if op&0x04 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 16
				pos++
			}
			if op&0x08 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				off |= int(deltaBytes[pos]) << 24
				pos++
			}
			if op&0x10 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos])
				pos++
			}
			if op&0x20 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos]) << 8
				pos++
			}
			if op&0x40 != 0 {
				if pos >= len(deltaBytes) {
					out.Release()
					return bufpool.Buffer{}, ErrInvalidObject
				}
				n |= int(deltaBytes[pos]) << 16
				pos++
			}
			if n == 0 {
				n = 0x10000
			}
			if off+n > len(baseBytes) || outPos+n > len(outBytes) {
				out.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			copy(outBytes[outPos:], baseBytes[off:off+n])
			outPos += n
		case op != 0:
			n := int(op)
			if pos+n > len(deltaBytes) || outPos+n > len(outBytes) {
				out.Release()
				return bufpool.Buffer{}, ErrInvalidObject
			}
			copy(outBytes[outPos:], deltaBytes[pos:pos+n])
			pos += n
			outPos += n
		default:
			out.Release()
			return bufpool.Buffer{}, ErrInvalidObject
		}
	}

	if outPos != len(outBytes) {
		out.Release()
		return bufpool.Buffer{}, ErrInvalidObject
	}
	return out, nil
}

func packVarintRead(buf []byte, pos *int) (int, error) {
	res := 0
	shift := 0
	for {
		if *pos >= len(buf) {
			return 0, ErrInvalidObject
		}
		b := buf[*pos]
		*pos++
		res |= int(b&0x7f) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}
	return res, nil
}
