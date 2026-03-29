package object

import (
	"fmt"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// ParseObjectWithHeader parses a loose object in "type size\x00body" format.
//
//nolint:ireturn
func ParseObjectWithHeader(raw []byte, algo objectid.Algorithm) (Object, error) {
	ty, size, headerLen, ok := objectheader.Parse(raw)
	if !ok {
		return nil, fmt.Errorf("object: malformed object header")
	}

	body := raw[headerLen:]
	if int64(len(body)) != size {
		return nil, fmt.Errorf("object: size mismatch: header says %d bytes, body has %d", size, len(body))
	}

	return ParseObjectWithoutHeader(ty, body, algo)
}
