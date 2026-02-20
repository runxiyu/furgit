// Package object provides Git object models and codecs.
package object

import (
	"bytes"
	"errors"
	"strconv"

	"codeberg.org/lindenii/furgit/objecttype"
)

var (
	// ErrInvalidObject indicates malformed serialized data.
	ErrInvalidObject = errors.New("object: invalid object encoding")
	// ErrNotFound indicates missing entries in in-memory lookups.
	ErrNotFound = errors.New("object: not found")
)

// Object is a Git object that can serialize itself.
type Object interface {
	ObjectType() objecttype.Type
	Serialize() ([]byte, error)
}

func headerForType(ty objecttype.Type, body []byte) ([]byte, error) {
	tyStr, ok := objecttype.Name(ty)
	if !ok {
		return nil, ErrInvalidObject
	}
	size := strconv.Itoa(len(body))
	var buf bytes.Buffer
	buf.Grow(len(tyStr) + len(size) + 2)
	buf.WriteString(tyStr)
	buf.WriteByte(' ')
	buf.WriteString(size)
	buf.WriteByte(0)
	return buf.Bytes(), nil
}
