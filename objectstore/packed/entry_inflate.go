package packed

import (
	"bytes"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/internal/zlib"
)

// zlibReaderAt opens a zlib reader starting at data offset within pack.
func zlibReaderAt(pack *packFile, offset int) (io.ReadCloser, error) {
	if offset < 0 || offset > len(pack.data) {
		return nil, fmt.Errorf("objectstore/packed: pack %q zlib offset out of bounds", pack.name)
	}
	return zlib.NewReader(bytes.NewReader(pack.data[offset:]))
}

// inflateAt inflates one entry payload from data offset.
//
// When expectedSize is non-negative, the inflated length must match.
func inflateAt(pack *packFile, offset int, expectedSize int64) ([]byte, error) {
	reader, err := zlibReaderAt(pack, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if expectedSize >= 0 && int64(len(body)) != expectedSize {
		return nil, fmt.Errorf(
			"objectstore/packed: pack %q inflated size mismatch: got %d want %d",
			pack.name,
			len(body),
			expectedSize,
		)
	}
	return body, nil
}
