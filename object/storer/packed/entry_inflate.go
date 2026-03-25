package packed

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
)

// zlibReaderAt opens a zlib reader starting at data offset within pack.
func zlibReaderAt(pack *packFile, offset int) (io.ReadCloser, error) {
	if offset < 0 || offset > len(pack.data) {
		return nil, fmt.Errorf("objectstorer/packed: pack %q zlib offset out of bounds", pack.name)
	}

	return zlib.NewReader(bytes.NewReader(pack.data[offset:]))
}

// inflateAt inflates one entry payload from data offset.
func inflateAt(pack *packFile, offset int, expectedSize int64) ([]byte, error) {
	reader, err := zlibReaderAt(pack, offset)
	if err != nil {
		return nil, err
	}

	defer func() { _ = reader.Close() }()

	if expectedSize >= 0 {
		if expectedSize > int64(math.MaxInt) {
			return nil, fmt.Errorf(
				"objectstorer/packed: pack %q expected inflated size overflows int: %d",
				pack.name,
				expectedSize,
			)
		}

		body := make([]byte, int(expectedSize))

		_, err := io.ReadFull(reader, body)
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}
