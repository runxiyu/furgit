package packed

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/store/packed/internal/reading"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

var _ objectstore.Reader = (*Store)(nil)

// ReadBytesFull reads a full serialized object as "type size\0content".
func (store *Store) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	return store.reader.ReadBytesFull(id)
}

// ReadBytesContent reads an object's type and content bytes.
func (store *Store) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	return store.reader.ReadBytesContent(id)
}

// ReadReaderFull reads a full serialized object stream as "type size\0content".
func (store *Store) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	return store.reader.ReadReaderFull(id)
}

// ReadReaderContent reads an object's type, declared content length, and
// content stream.
func (store *Store) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	return store.reader.ReadReaderContent(id)
}

// ReadSize reads an object's declared content length.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	return store.reader.ReadSize(id)
}

// ReadHeader reads an object's type and declared content length.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	return store.reader.ReadHeader(id)
}

// Refresh updates the packed-store view of on-disk pack/index candidates.
func (store *Store) Refresh() error {
	return store.reader.Refresh()
}

func (opts Options) toReadingOptions() reading.Options {
	var refreshPolicy reading.RefreshPolicy

	switch opts.RefreshPolicy {
	case RefreshPolicyOnMissing:
		refreshPolicy = reading.RefreshPolicyOnMissing
	case RefreshPolicyNever:
		refreshPolicy = reading.RefreshPolicyNever
	default:
		refreshPolicy = reading.RefreshPolicy(opts.RefreshPolicy)
	}

	return reading.Options{
		RefreshPolicy: refreshPolicy,
	}
}
