package header

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/lgo/intconv"
)

// ErrInvalidHeader indicates a malformed loose-object header.
var ErrInvalidHeader = errors.New("object/header: invalid header")

// Parse parses a canonical loose-object header ("type size\x00").
func Parse(data []byte) (ty typ.Type, size int, consumed int, err error) {
	space := bytes.IndexByte(data, ' ')
	if space <= 0 {
		return 0, 0, 0, fmt.Errorf("%w: missing ' ' type/size separator", ErrInvalidHeader)
	}

	nulRel := bytes.IndexByte(data[space+1:], 0)
	if nulRel < 0 {
		return 0, 0, 0, fmt.Errorf("%w: missing NUL terminator", ErrInvalidHeader)
	}

	nul := space + 1 + nulRel

	ty, err = typ.Parse(string(data[:space]))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: type %q: %w", ErrInvalidHeader, data[:space], err)
	}

	sizeBytes := data[space+1 : nul]
	if len(sizeBytes) == 0 {
		return 0, 0, 0, fmt.Errorf("%w: empty size field", ErrInvalidHeader)
	}

	sizeU, err := strconv.ParseUint(string(sizeBytes), 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: size %q: %w", ErrInvalidHeader, sizeBytes, err)
	}

	size, err = intconv.Uint64ToInt(sizeU)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: size %q: %w", ErrInvalidHeader, sizeBytes, err)
	}

	return ty, size, nul + 1, nil
}
