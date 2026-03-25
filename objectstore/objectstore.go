// Package objectstore provides interfaces for object storage backends.
package objectstore

import (
	"errors"
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ErrObjectNotFound indicates that an object does not exist in a backend.
// This error MUST only be used in situations where the object store has
// no specified object ID, but no other unexpected conditions were
// encountered. In particular, it is not suitable for situations where one
// object references another (such as a tree referencing a blob) but
// the latter does not exist; these situations should use a separate
// error (TODO).
var ErrObjectNotFound = errors.New("objectstore: object not found")

// Store reads Git objects by object ID.
//
// Unless an implementation explicitly documents otherwise, values returned by
// Store methods are only valid until the store is closed.
type Store interface {
	// ReadBytesFull reads a full serialized object as "type size\0content".
	//
	// In a valid repository, hashing this payload with the same algorithm yields
	// the requested object ID. Readers should treat this as a repository
	// invariant and should not re-verify it on every read.
	//
	// Any read-time integrity verification beyond producing this payload is
	// implementation-defined.
	ReadBytesFull(id objectid.ObjectID) ([]byte, error)

	// ReadBytesContent reads an object's type and content bytes.
	//
	// Any read-time integrity verification beyond producing this payload is
	// implementation-defined.
	ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error)

	// ReadReaderFull reads a full serialized object stream as "type size\0content".
	//
	// Caller must close the returned reader.
	// The returned reader is only valid until the store is closed.
	//
	// Any read-time integrity verification performed while producing the stream
	// is implementation-defined.
	ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error)

	// ReadReaderContent reads an object's type, declared content length,
	// and content stream.
	//
	// Caller must close the returned reader.
	// The returned reader is only valid until the store is closed.
	//
	// Any read-time integrity verification performed while producing the stream
	// is implementation-defined.
	ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error)

	// ReadSize reads an object's declared content length.
	//
	// This is equivalent to ReadHeader(...).size and may be cheaper than
	// ReadHeader when callers do not need object type.
	//
	// Any read-time integrity verification performed to produce the size is
	// implementation-defined.
	ReadSize(id objectid.ObjectID) (int64, error)

	// ReadHeader reads an object's type and declared content length.
	//
	// Any read-time integrity verification performed to produce the header is
	// implementation-defined.
	ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error)

	// Refresh updates any backend-local discovery/cache view of on-disk objects.
	//
	// Backends without dynamic discovery should return nil.
	Refresh() error

	// Close releases resources associated with the backend.
	//
	// Repeated calls to Close are undefined behavior unless the implementation
	// explicitly documents otherwise.
	Close() error
}

// type Cursor any
//
// Then make all read functions accept and provide a Cursor
// nil must always be accepted and would exhibit the same behavior as right now
// Non-nil behavior is implementation-defined: e.g., pack selection
