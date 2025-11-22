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
func Decompress(src []byte) (*bufpool.Buffer, error) {
	return DecompressSized(src, 0)
}

// DecompressSized inflates the provided zlib stream, using sizeHint to
// preallocate the output buffer when known (e.g. packfile entries).
func DecompressSized(src []byte, sizeHint int) (*bufpool.Buffer, error) {
	return DecompressDictSized(src, nil, sizeHint)
}

// DecompressDict is like Decompress but accepts a preset dictionary. The
// dictionary must match the checksum embedded in the stream if the dictionary
// flag is present.
func DecompressDict(src []byte, dict []byte) (*bufpool.Buffer, error) {
	return DecompressDictSized(src, dict, 0)
}

// DecompressDictSized is like DecompressDict but allows providing an expected
// uncompressed size to avoid buffer growth copies.
func DecompressDictSized(src []byte, dict []byte, sizeHint int) (*bufpool.Buffer, error) {
	if len(src) < 6 {
		return nil, io.ErrUnexpectedEOF
	}

	cmf := src[0]
	flg := src[1]
	if (cmf&0x0f != zlibDeflate) || (cmf>>4 > zlibMaxWindow) || (binary.BigEndian.Uint16(src[:2])%31 != 0) {
		return nil, ErrHeader
	}

	offset := 2
	haveDict := flg&0x20 != 0
	if haveDict {
		if len(src) < offset+4 {
			return nil, io.ErrUnexpectedEOF
		}
		if dict == nil {
			return nil, ErrDictionary
		}
		checksum := binary.BigEndian.Uint32(src[offset : offset+4])
		if checksum != adler32.Checksum(dict) {
			return nil, ErrDictionary
		}
		offset += 4
	}

	if len(src[offset:]) < 4 {
		return nil, io.ErrUnexpectedEOF
	}

	deflateData := src[offset:]
	out, consumed, err := flatex.DecompressDictSized(deflateData, dict, sizeHint)
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
