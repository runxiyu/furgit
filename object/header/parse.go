package header

import (
	"bytes"
	"strconv"

	"codeberg.org/lindenii/furgit/object/typ"
)

// Parse parses a canonical loose-object header ("type size\x00").
func Parse(data []byte) (ty typ.Type, size int64, consumed int, ok bool) {
	space := bytes.IndexByte(data, ' ')
	if space <= 0 {
		return typ.TypeInvalid, 0, 0, false
	}

	nulRel := bytes.IndexByte(data[space+1:], 0)
	if nulRel < 0 {
		return typ.TypeInvalid, 0, 0, false
	}

	nul := space + 1 + nulRel

	ty, ok = typ.Parse(string(data[:space]))
	if !ok {
		return typ.TypeInvalid, 0, 0, false
	}

	sizeBytes := data[space+1 : nul]
	if len(sizeBytes) == 0 {
		return typ.TypeInvalid, 0, 0, false
	}

	size, err := strconv.ParseInt(string(sizeBytes), 10, 64)
	if err != nil || size < 0 {
		return typ.TypeInvalid, 0, 0, false
	}

	return ty, size, nul + 1, true
}
