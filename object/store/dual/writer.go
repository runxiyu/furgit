package dual

import (
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// WriteBytesFull writes one full serialized object to the object side.
func (dual *Dual) WriteBytesFull(raw []byte) (id.ObjectID, error) {
	return dual.object.WriteBytesFull(raw) //nolint:wrapcheck
}

// WriteBytesContent writes one typed object content byte slice to the object side.
func (dual *Dual) WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error) {
	return dual.object.WriteBytesContent(ty, content) //nolint:wrapcheck
}

// WriteReaderFull writes one full serialized object stream to the object side.
func (dual *Dual) WriteReaderFull(src io.Reader) (id.ObjectID, error) {
	return dual.object.WriteReaderFull(src) //nolint:wrapcheck
}

// WriteReaderContent writes one typed object content stream to the object side.
func (dual *Dual) WriteReaderContent(ty typ.Type, size uint64, src io.Reader) (id.ObjectID, error) {
	return dual.object.WriteReaderContent(ty, size, src) //nolint:wrapcheck
}

// WritePack ingests one pack stream into the pack side.
func (dual *Dual) WritePack(src io.Reader, opts store.PackWriteOptions) error {
	return dual.pack.WritePack(src, opts) //nolint:wrapcheck
}
