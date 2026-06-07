package object

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/blob"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

// ErrSizeMismatch indicates a mismatch
// between the size declared in the object header
// and the size of the object body.
var ErrSizeMismatch = errors.New("object: size mismatch")

// ParseWithHeader parses a loose object
// in "type size\x00body" format.
//
//nolint:ireturn
func ParseWithHeader(raw []byte, objectFormat id.ObjectFormat) (Object, error) {
	ty, size, headerLen, err := header.Parse(raw)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	body := raw[headerLen:]
	if uint64(len(body)) != size {
		return nil, fmt.Errorf("%w: header declares %d bytes, body has %d", ErrSizeMismatch, size, len(body))
	}

	return ParseWithoutHeader(ty, body, objectFormat)
}

// ParseWithoutHeader parses a typed object body.
//
//nolint:ireturn
func ParseWithoutHeader(ty typ.Type, body []byte, objectFormat id.ObjectFormat) (Object, error) {
	switch ty {
	case typ.TypeBlob:
		return blob.Parse(body) //nolint:wrapcheck
	case typ.TypeTree:
		panic("TODO")
	case typ.TypeCommit:
		return commit.Parse(body, objectFormat) //nolint:wrapcheck
	case typ.TypeTag:
		return tag.Parse(body, objectFormat) //nolint:wrapcheck
	case typ.TypeUnknown:
		return nil, typ.ErrInvalidType
	default:
		return nil, typ.ErrInvalidType
	}
}
