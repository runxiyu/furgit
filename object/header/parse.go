package objectheader

import (
	"bytes"
	"strconv"

	objecttype "lindenii.org/go/furgit/object/type"
)

// Parse parses a canonical loose-object header ("type size\\x00").
// It returns the parsed type, size, bytes consumed (including trailing NUL),
// and whether parsing succeeded.
func Parse(data []byte) (objecttype.Type, int64, int, bool) {
	space := bytes.IndexByte(data, ' ')
	if space <= 0 {
		return objecttype.TypeInvalid, 0, 0, false
	}

	nulRel := bytes.IndexByte(data[space+1:], 0)
	if nulRel < 0 {
		return objecttype.TypeInvalid, 0, 0, false
	}

	nul := space + 1 + nulRel

	ty, ok := objecttype.Parse(string(data[:space]))
	if !ok {
		return objecttype.TypeInvalid, 0, 0, false
	}

	sizeBytes := data[space+1 : nul]
	if len(sizeBytes) == 0 {
		return objecttype.TypeInvalid, 0, 0, false
	}

	size, err := strconv.ParseInt(string(sizeBytes), 10, 64)
	if err != nil || size < 0 {
		return objecttype.TypeInvalid, 0, 0, false
	}

	return ty, size, nul + 1, true
}
