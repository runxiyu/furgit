package zlibx

import (
	"encoding/binary"
	"io"

	"git.sr.ht/~runxiyu/furgit/internal/adler32"
	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
	"git.sr.ht/~runxiyu/furgit/internal/flatex"
)

// Decompress inflates the provided zlib wrapped stream and returns the
// uncompressed data inside a pooled bufpool.Buffer.
func Decompress(src []byte) (bufpool.Buffer, error) {
	return DecompressDict(src, nil)
}

// DecompressDict is like Decompress but accepts a preset dictionary. The
// dictionary must match the checksum embedded in the stream if the dictionary
// flag is present.
func DecompressDict(src []byte, dict []byte) (bufpool.Buffer, error) {
	if len(src) < 6 {
		return bufpool.Buffer{}, io.ErrUnexpectedEOF
	}

	cmf := src[0]
	flg := src[1]
	if (cmf&0x0f != zlibDeflate) || (cmf>>4 > zlibMaxWindow) || (binary.BigEndian.Uint16(src[:2])%31 != 0) {
		return bufpool.Buffer{}, ErrHeader
	}

	offset := 2
	haveDict := flg&0x20 != 0
	if haveDict {
		if len(src) < offset+4 {
			return bufpool.Buffer{}, io.ErrUnexpectedEOF
		}
		if dict == nil {
			return bufpool.Buffer{}, ErrDictionary
		}
		checksum := binary.BigEndian.Uint32(src[offset : offset+4])
		if checksum != adler32.Checksum(dict) {
			return bufpool.Buffer{}, ErrDictionary
		}
		offset += 4
	}

	if len(src[offset:]) < 4 {
		return bufpool.Buffer{}, io.ErrUnexpectedEOF
	}

	deflateData := src[offset:]
	out, consumed, err := flatex.DecompressDict(deflateData, dict)
	if err != nil {
		return bufpool.Buffer{}, err
	}

	checksumPos := offset + consumed
	if checksumPos+4 > len(src) {
		out.Release()
		return bufpool.Buffer{}, io.ErrUnexpectedEOF
	}
	expected := binary.BigEndian.Uint32(src[checksumPos : checksumPos+4])
	if expected != adler32.Checksum(out.Bytes()) {
		out.Release()
		return bufpool.Buffer{}, ErrChecksum
	}
	return out, nil
}
