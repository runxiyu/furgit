package loose

import (
	"bufio"
	"errors"
	"io"
	"os"

	"lindenii.org/go/furgit/internal/compress/zlib"
	objectheader "lindenii.org/go/furgit/object/header"
	objecttype "lindenii.org/go/furgit/object/type"
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
	ty, size, headerLen, ok := objectheader.Parse(raw)
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
	header, err := br.ReadSlice(0)
	if err != nil {
		return nil, objecttype.TypeInvalid, 0, err
	}

	ty, size, _, ok := objectheader.Parse(header)
	if !ok {
		return nil, objecttype.TypeInvalid, 0, errors.New("objectstore/loose: malformed object header")
	}

	return header, ty, size, nil
}
