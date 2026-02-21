// Package objectstore provides interfaces for object storage backends.
package objectstore

import (
	"errors"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ErrObjectNotFound indicates that an object does not exist in a backend.
// TODO: This might need to be an interface or otherwise be able to encapsulate multiple concrete backends'.
var ErrObjectNotFound = errors.New("objectstore: object not found")

// Store reads Git objects by object ID.
type Store interface {
	// ReadBytesFull reads a full serialized object as "type size\\x00content".
	//
	// In a valid repository, hashing this payload with the same algorithm yields
	// the requested object ID. Readers should treat this as a repository
	// invariant and should not re-verify it on every read.
	ReadBytesFull(id objectid.ObjectID) ([]byte, error)
	// ReadBytesContent reads an object's type and content bytes.
	ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error)
	// ReadReaderFull reads a full serialized object stream as "type size\\x00content".
	// Caller must close the returned reader.
	ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error)
	// ReadReaderContent reads an object's type, declared content length,
	// and content stream.
	// Caller must close the returned reader.
	ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error)
	// ReadSize reads an object's declared content length.
	//
	// This is equivalent to ReadHeader(...).size and may be cheaper than
	// ReadHeader when callers do not need object type.
	ReadSize(id objectid.ObjectID) (int64, error)
	// ReadHeader reads an object's type and declared content length.
	ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error)
	// Close releases resources associated with the backend.
	Close() error
}
