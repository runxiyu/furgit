package objectstore

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadingStore reads Git objects by object ID.
//
// Methods may perform implementation-defined integrity verification beyond
// successfully producing their documented result.
//
// Labels: MT-Safe.
type ReadingStore interface {
	// ReadBytesFull reads a full serialized object as "type size\0content".
	//
	// In a valid repository, hashing this payload with the same algorithm yields
	// the requested object ID. Readers should treat this as a repository
	// invariant and should not re-verify it on every read.
	//
	// Labels: Life-Parent.
	ReadBytesFull(id objectid.ObjectID) ([]byte, error)

	// ReadBytesContent reads an object's type and content bytes.
	//
	// Labels: Life-Parent.
	ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error)

	// ReadReaderFull reads a full serialized object stream as "type size\0content".
	//
	// Labels: Life-Parent, Close-Caller.
	ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error)

	// ReadReaderContent reads an object's type, declared content length,
	// and content stream.
	//
	// Labels: Life-Parent, Close-Caller.
	ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error)

	// ReadSize reads an object's declared content length.
	//
	// This is equivalent to ReadHeader(...).size and may be cheaper than
	// ReadHeader when callers do not need object type.
	ReadSize(id objectid.ObjectID) (int64, error)

	// ReadHeader reads an object's type and declared content length.
	ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error)

	// Refresh updates any backend-local discovery/cache view of on-disk objects.
	//
	// Backends without dynamic discovery should do nothing and return nil.
	Refresh() error
}
