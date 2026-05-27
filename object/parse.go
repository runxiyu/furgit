package object

import (
	"fmt"

	"lindenii.org/go/furgit/object/blob"
	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// SizeMismatchError indicates a mismatch
// between the size expected from the object header
// and the size of the object.
type SizeMismatchError struct {
	Expected uint64
	Got      int
}

func (sizeMismatchError SizeMismatchError) Error() string {
	return fmt.Sprintf(
		"object: size mismatch: header says %d bytes, but got %d from body",
		sizeMismatchError.Expected, sizeMismatchError.Got,
	)
}

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
		return nil, SizeMismatchError{Expected: size, Got: len(body)}
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
		panic("TODO")
	default:
		return nil, typ.ErrInvalidType
	}
}
