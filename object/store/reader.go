package store

import (
	"errors"
	"io"

	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/typ"
)

// ObjectReader reads Git objects by object ID.
//
// Methods may perform implementation-defined integrity verification
// beyond successfully producing their documented result.
//
// Labels: MT-Safe.
type ObjectReader interface {
	// ReadBytesFull reads a full serialized object as "type size\0content".
	//
	// In a valid repository,
	// hashing this payload with the same object format
	// yields the requested object ID.
	// Users should treat this as an invariant;
	// implementations should not re-verify it on every read.
	//
	// Labels: Life-Parent.
	ReadBytesFull(id id.ObjectID) ([]byte, error)

	// ReadBytesContent reads an object's type and content bytes.
	//
	// Labels: Life-Parent.
	ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error)

	// ReadReaderFull reads a full serialized object stream
	// as "type size\0content".
	//
	// Labels: Life-Parent, Close-Caller.
	ReadReaderFull(id id.ObjectID) (io.ReadCloser, error)

	// ReadReaderContent reads an object's type,
	// declared content length, and content stream.
	//
	// Labels: Life-Parent, Close-Caller.
	ReadReaderContent(id id.ObjectID) (typ.Type, int64, io.ReadCloser, error)

	// ReadSize reads an object's declared content length.
	//
	// This is equivalent to ReadHeader(...).size;
	// for some implementations, this may be cheaper than ReadHeader
	// when callers do not need object type.
	ReadSize(id id.ObjectID) (int64, error)

	// ReadHeader reads an object's type and declared content length.
	ReadHeader(id id.ObjectID) (typ.Type, int64, error)

	// Refresh updates any backend-local discovery/cache view of on-disk objects.
	//
	// Backends without dynamic discovery should do nothing and return nil.
	Refresh() error
}

// ErrObjectNotFound indicates that
// an object does not exist in a backend.
// This error must only be produced by object stores,
// when it has no specified object ID,
// but no other unexpected conditions were encountered.
var ErrObjectNotFound = errors.New("objectstore: object not found")

// This is a sentinel with no details,
// because it could be a frequent occurrence,
// and allocating frequently on expected error paths
// would be extremely harmful to performance.
// Sometime, I will audit this again.
// TODO
