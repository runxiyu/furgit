package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/blob"
	"codeberg.org/lindenii/furgit/object/commit"
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tag"
	"codeberg.org/lindenii/furgit/object/tree"
	"codeberg.org/lindenii/furgit/object/typ"
)

// ParseWithHeader parses a loose object in "type size\x00body" format.
//
//nolint:ireturn
func ParseWithHeader(raw []byte, algo id.Algorithm) (Object, error) {
	ty, size, headerLen, err := header.Parse(raw)
	if err != nil {
		return nil, err
	}

	body := raw[headerLen:]
	if uint64(len(body)) != size {
		return nil, fmt.Errorf("object: size mismatch: header says %d bytes, body has %d", size, len(body))
	}

	return ParseWithoutHeader(ty, body, algo)
}

// ParseWithoutHeader parses a typed object body.
//
//nolint:ireturn
func ParseWithoutHeader(ty typ.Type, body []byte, algo id.Algorithm) (Object, error) {
	switch ty {
	case typ.TypeBlob:
		return blob.Parse(body)
	case typ.TypeTree:
		return tree.Parse(body, algo)
	case typ.TypeCommit:
		return commit.Parse(body, algo)
	case typ.TypeTag:
		return tag.Parse(body, algo)
	default:
		return nil, fmt.Errorf("object: unsupported object type %d", ty)
	}
}
