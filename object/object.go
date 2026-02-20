// Package object provides Git object models and codecs.
package object

import (
	"errors"

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
	SerializeWithoutHeader() ([]byte, error)
	SerializeWithHeader() ([]byte, error)
}
