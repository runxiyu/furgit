package zlibx

import (
	"encoding/binary"
	"io"

	"git.sr.ht/~runxiyu/furgit/internal/adler32"
	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
	"git.sr.ht/~runxiyu/furgit/internal/flatex"
)

func Decompress(src []byte) (*bufpool.Buffer, error) {
	return DecompressSized(src, 0)
}

func DecompressSized(src []byte, sizeHint int) (*bufpool.Buffer, error) {
	if len(src) < 6 {
		return nil, io.ErrUnexpectedEOF
	}

	cmf := src[0]
	flg := src[1]
	if (cmf&0x0f != zlibDeflate) || (cmf>>4 > zlibMaxWindow) || (binary.BigEndian.Uint16(src[:2])%31 != 0) {
		return nil, ErrHeader
	}

	offset := 2
	if flg&0x20 != 0 {
		return nil, ErrHeader
	}

	if len(src[offset:]) < 4 {
		return nil, io.ErrUnexpectedEOF
	}

	deflateData := src[offset:]
	out, consumed, err := flatex.DecompressSized(deflateData, sizeHint)
	if err != nil {
		return nil, err
	}

	checksumPos := offset + consumed
	if checksumPos+4 > len(src) {
		out.Release()
		return nil, io.ErrUnexpectedEOF
	}
	expected := binary.BigEndian.Uint32(src[checksumPos : checksumPos+4])
	if expected != adler32.Checksum(out.Bytes()) {
		out.Release()
		return nil, ErrChecksum
	}
	return out, nil
}
