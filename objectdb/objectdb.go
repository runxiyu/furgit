// Package objectdb provides storage interfaces for Git objects.
package objectdb

import (
	"errors"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ErrObjectNotFound indicates that an object does not exist in a backend.
// TODO: This might need to be an interface or otherwise be able to encapsulate multiple concrete backends'
var ErrObjectNotFound = errors.New("objectdb: object not found")

// ObjectDB reads Git objects by object ID.
type ObjectDB interface {
	// ReadBytesFull reads a full serialized object as "type size\\x00content".
	// If hashed with the same algorithm it MUST match the object ID.
	ReadBytesFull(id objectid.ObjectID) ([]byte, error)
	// ReadBytesContent reads an object's type and content bytes.
	ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error)
	// ReadHeader reads an object's type and declared content length.
	ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error)
	// Close releases resources associated with the backend.
	Close() error
}
