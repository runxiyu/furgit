package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/objectheader"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/oid"
)

// ParseObjectWithoutHeader parses a typed object body.
func ParseObjectWithoutHeader(ty objecttype.Type, body []byte, algo oid.Algorithm) (Object, error) {
	switch ty {
	case objecttype.TypeBlob:
		return ParseBlob(body)
	case objecttype.TypeTree:
		return ParseTree(body, algo)
	case objecttype.TypeCommit:
		return ParseCommit(body, algo)
	case objecttype.TypeTag:
		return ParseTag(body, algo)
	default:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	}
}

// ParseObjectWithHeader parses a loose object in "type size\\x00body" format.
func ParseObjectWithHeader(raw []byte, algo oid.Algorithm) (Object, error) {
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
