package loose

import (
	"bufio"
	"compress/zlib"
	"errors"
	"io"
	"os"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objecttype"
)

// decodeAll inflates the full loose object payload from file.
func decodeAll(file *os.File) ([]byte, error) {
	zr, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	return io.ReadAll(zr)
}

// parseRaw parses a loose object payload in "type size\0content" format.
func parseRaw(raw []byte) (objecttype.Type, []byte, error) {
	ty, _, headerLen, ok := objectheader.Parse(raw)
	if !ok {
		return objecttype.TypeInvalid, nil, errors.New("objectstore/loose: malformed object header")
	}
	return ty, raw[headerLen:], nil
}

// readHeader reads and parses a loose object header from br.
// br must be positioned at the start of decoded loose object bytes.
func readHeader(br *bufio.Reader) (objecttype.Type, int64, error) {
	header, err := br.ReadSlice(0)
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}
	ty, size, _, ok := objectheader.Parse(header)
	if !ok {
		return objecttype.TypeInvalid, 0, errors.New("objectstore/loose: malformed object header")
	}
	return ty, size, nil
}
