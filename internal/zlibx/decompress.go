package zlibx

import (
	"encoding/binary"
	"io"

	"codeberg.org/lindenii/furgit/internal/adler32"
	"codeberg.org/lindenii/furgit/internal/bufpool"
	"codeberg.org/lindenii/furgit/internal/flatex"
)

func Decompress(src []byte) (bufpool.Buffer, error) {
	out, _, err := DecompressSized(src, 0)
	return out, err
}

func DecompressSized(src []byte, sizeHint int) (bufpool.Buffer, int, error) {
	if len(src) < 6 {
		return bufpool.Buffer{}, 0, io.ErrUnexpectedEOF
	}

	cmf := src[0]
	flg := src[1]
	if (cmf&0x0f != zlibDeflate) || (cmf>>4 > zlibMaxWindow) || (binary.BigEndian.Uint16(src[:2])%31 != 0) {
		return bufpool.Buffer{}, 0, ErrHeader
	}

	offset := 2
	if flg&0x20 != 0 {
		return bufpool.Buffer{}, 0, ErrHeader
	}

	if len(src[offset:]) < 4 {
		return bufpool.Buffer{}, 0, io.ErrUnexpectedEOF
	}

	deflateData := src[offset:]
	out, consumed, err := flatex.DecompressSized(deflateData, sizeHint)
	if err != nil {
		return bufpool.Buffer{}, 0, err
	}

	checksumPos := offset + consumed
	if checksumPos+4 > len(src) {
		out.Release()
		return bufpool.Buffer{}, 0, io.ErrUnexpectedEOF
	}
	expected := binary.BigEndian.Uint32(src[checksumPos : checksumPos+4])
	if expected != adler32.Checksum(out.Bytes()) {
		out.Release()
		return bufpool.Buffer{}, 0, ErrChecksum
	}
	return out, checksumPos + 4, nil
}
