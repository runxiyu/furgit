package dual

import (
	"io"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// WriteReaderContent writes one typed object content stream to the object-wise
// store.
func (dual *Dual) WriteReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error) {
	return dual.object.WriteReaderContent(ty, size, src)
}

// WriteReaderFull writes one full serialized object stream as
// "type size\0content" to the object-wise store.
func (dual *Dual) WriteReaderFull(src io.Reader) (objectid.ObjectID, error) {
	return dual.object.WriteReaderFull(src)
}

// WriteBytesContent writes one typed object content byte slice to the
// object-wise store.
func (dual *Dual) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	return dual.object.WriteBytesContent(ty, content)
}

// WriteBytesFull writes one full serialized object byte slice as
// "type size\0content" to the object-wise store.
func (dual *Dual) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	return dual.object.WriteBytesFull(raw)
}
