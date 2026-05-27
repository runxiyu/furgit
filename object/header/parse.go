package header

import (
	"bytes"
	"errors"
	"strconv"

	"lindenii.org/go/furgit/object/typ"
)

var ErrInvalidHeader = errors.New("object/header: invalid header")

// Parse parses a canonical loose-object header ("type size\x00").
func Parse(data []byte) (ty typ.Type, size uint64, consumed int, err error) {
	space := bytes.IndexByte(data, ' ')
	if space <= 0 {
		return 0, 0, 0, ErrInvalidHeader
	}

	nulRel := bytes.IndexByte(data[space+1:], 0)
	if nulRel < 0 {
		return 0, 0, 0, ErrInvalidHeader
	}

	nul := space + 1 + nulRel

	ty, err = typ.Parse(string(data[:space]))
	if err != nil {
		return 0, 0, 0, ErrInvalidHeader
	}

	sizeBytes := data[space+1 : nul]
	if len(sizeBytes) == 0 {
		return 0, 0, 0, ErrInvalidHeader
	}

	size, err = strconv.ParseUint(string(sizeBytes), 10, 64)
	if err != nil {
		return 0, 0, 0, ErrInvalidHeader
	}

	return ty, size, nul + 1, nil
}
