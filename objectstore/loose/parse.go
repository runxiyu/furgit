package loose

import (
	"bufio"
	"errors"
	"io"
	"os"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/object/header"
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
	ty, size, headerLen, ok := header.Parse(raw)
	if !ok {
		return objecttype.TypeInvalid, nil, errors.New("objectstore/loose: malformed object header")
	}

	content := raw[headerLen:]
	if int64(len(content)) != size {
		return objecttype.TypeInvalid, nil, errors.New("objectstore/loose: object header size/content mismatch")
	}

	return ty, content, nil
}

// readHeader reads and parses a loose object header from br, and returns
// the raw header bytes including the trailing NUL.
func readHeader(br *bufio.Reader) ([]byte, objecttype.Type, int64, error) {
	h, err := br.ReadSlice(0)
	if err != nil {
		return nil, objecttype.TypeInvalid, 0, err
	}

	ty, size, _, ok := header.Parse(h)
	if !ok {
		return nil, objecttype.TypeInvalid, 0, errors.New("objectstore/loose: malformed object header")
	}

	return h, ty, size, nil
}
