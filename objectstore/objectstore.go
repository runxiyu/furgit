// Package objectstore provides storage interfaces for Git objects.
package objectstore

import (
	"errors"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ErrObjectNotFound indicates that an object does not exist in a backend.
// TODO: This might need to be an interface or otherwise be able to encapsulate multiple concrete backends'
var ErrObjectNotFound = errors.New("objectstore: object not found")

// ObjectStore reads Git objects by object ID.
type ObjectStore interface {
	// ReadBytesFull reads a full serialized object as "type size\\x00content".
	// If hashed with the same algorithm it MUST match the object ID.
	ReadBytesFull(id objectid.ObjectID) ([]byte, error)
	// ReadBytesContent reads an object's type and content bytes.
	ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error)
	// ReadReaderFull reads a full serialized object stream as "type size\\x00content".
	// Caller must close the returned reader.
	ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error)
	// ReadReaderContent reads an object's type and content stream.
	// Caller must close the returned reader.
	ReadReaderContent(id objectid.ObjectID) (objecttype.Type, io.ReadCloser, error)
	// ReadHeader reads an object's type and declared content length.
	ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error)
	// Close releases resources associated with the backend.
	Close() error
}
