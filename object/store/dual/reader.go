package dual

import (
	"io"

	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// ReadBytesFull reads a full serialized object as "type size\0content" from
// either store.
func (dual *Dual) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	return dual.reader.ReadBytesFull(id)
}

// ReadBytesContent reads an object's type and content bytes from either store.
func (dual *Dual) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	return dual.reader.ReadBytesContent(id)
}

// ReadReaderFull reads a full serialized object stream as "type size\0content"
// from either store.
func (dual *Dual) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	return dual.reader.ReadReaderFull(id)
}

// ReadReaderContent reads an object's type, declared content length, and
// content stream from either store.
func (dual *Dual) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	return dual.reader.ReadReaderContent(id)
}

// ReadSize reads an object's declared content length from either store.
func (dual *Dual) ReadSize(id objectid.ObjectID) (int64, error) {
	return dual.reader.ReadSize(id)
}

// ReadHeader reads an object's type and declared content length from either
// store.
func (dual *Dual) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	return dual.reader.ReadHeader(id)
}

// Refresh refreshes both underlying stores and the combined read view.
func (dual *Dual) Refresh() error {
	err := dual.object.Refresh()
	if err != nil {
		return err
	}

	err = dual.pack.Refresh()
	if err != nil {
		return err
	}

	return dual.reader.Refresh()
}
