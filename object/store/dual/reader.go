package dual

import (
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads a full serialized object from the combined view.
func (dual *Dual) ReadBytesFull(id id.ObjectID) ([]byte, error) {
	return dual.reader.ReadBytesFull(id) //nolint:wrapcheck
}

// ReadBytesContent reads an object's type and content bytes from the combined view.
func (dual *Dual) ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error) {
	return dual.reader.ReadBytesContent(id) //nolint:wrapcheck
}

// ReadReaderFull reads a full serialized object stream from the combined view.
func (dual *Dual) ReadReaderFull(id id.ObjectID) (io.ReadCloser, error) {
	return dual.reader.ReadReaderFull(id) //nolint:wrapcheck
}

// ReadReaderContent reads an object's type, declared content length,
// and content stream from the combined view.
func (dual *Dual) ReadReaderContent(id id.ObjectID) (typ.Type, int, io.ReadCloser, error) {
	return dual.reader.ReadReaderContent(id) //nolint:wrapcheck
}

// ReadSize reads an object's declared content length from the combined view.
func (dual *Dual) ReadSize(id id.ObjectID) (int, error) {
	return dual.reader.ReadSize(id) //nolint:wrapcheck
}

// ReadHeader reads an object's type and declared content length from the combined view.
func (dual *Dual) ReadHeader(id id.ObjectID) (typ.Type, int, error) {
	return dual.reader.ReadHeader(id) //nolint:wrapcheck
}

// Refresh refreshes both sides.
func (dual *Dual) Refresh() error {
	return dual.reader.Refresh() //nolint:wrapcheck
}
