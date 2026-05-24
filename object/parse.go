package object

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object/blob"
	"codeberg.org/lindenii/furgit/object/commit"
	"codeberg.org/lindenii/furgit/object/header"
	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/typ"
	// "codeberg.org/lindenii/furgit/object/tag"
	// "codeberg.org/lindenii/furgit/object/tree"
)

// SizeMismatchError indicates a mismatch
// between the size expected from the object header
// and the size of the object.
type SizeMismatchError struct {
	Expected int
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
func ParseWithHeader(raw []byte, algo id.Algorithm) (Object, error) {
	ty, size, headerLen, err := header.Parse(raw)
	if err != nil {
		return nil, err
	}

	body := raw[headerLen:]
	if uint64(len(body)) != size {
		return nil, SizeMismatchError{Expected: int(size), Got: len(body)}
	}

	return ParseWithoutHeader(ty, body, algo)
}

// ParseWithoutHeader parses a typed object body.
//
//nolint:ireturn
func ParseWithoutHeader(ty typ.Type, body []byte, algo id.Algorithm) (Object, error) {
	switch ty {
	case typ.TypeBlob:
		return blob.Parse(body) //nolint:wrapcheck
	case typ.TypeTree:
		panic("TODO")
	case typ.TypeCommit:
		return commit.Parse(body, algo) //nolint:wrapcheck
	case typ.TypeTag:
		panic("TODO")
	default:
		return nil, typ.ErrInvalidType
	}
}
