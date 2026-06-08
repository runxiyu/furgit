package loose

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// decodeAll inflates the full loose object payload from file.
func decodeAll(file *os.File) ([]byte, error) {
	zr, err := zlib.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("object/store/loose: %w", err)
	}

	defer func() { _ = zr.Close() }()

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("object/store/loose: %w", err)
	}

	return data, nil
}

// parseRaw parses a loose object payload in "type size\x00content" format.
func parseRaw(raw []byte) (typ.Type, []byte, error) {
	ty, size, consumed, err := header.Parse(raw)
	if err != nil {
		return typ.TypeUnknown, nil, fmt.Errorf("%w: %w", store.ErrInvalidObject, err)
	}

	content := raw[consumed:]
	if uint64(len(content)) != size {
		return typ.TypeUnknown, nil, fmt.Errorf("%w: header size/content mismatch", store.ErrInvalidObject)
	}

	return ty, content, nil
}

// readHeader reads and parses a loose object header from br,
// and returns the raw header bytes including the trailing NUL.
func readHeader(br *bufio.Reader) ([]byte, typ.Type, uint64, error) {
	headerBytes, err := br.ReadSlice(0)
	if err != nil {
		return nil, typ.TypeUnknown, 0, fmt.Errorf("object/store/loose: %w", err)
	}

	ty, size, _, err := header.Parse(headerBytes)
	if err != nil {
		return nil, typ.TypeUnknown, 0, fmt.Errorf("%w: %w", store.ErrInvalidObject, err)
	}

	return headerBytes, ty, size, nil
}
