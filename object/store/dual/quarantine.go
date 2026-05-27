package dual

import (
	"io"

	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objectmix "lindenii.org/go/furgit/object/store/mix"
	objecttype "lindenii.org/go/furgit/object/type"
)

// quarantine is one coordinated dual quarantine over both stores.
type quarantine struct {
	objectQ objectstore.ObjectQuarantine
	packQ   objectstore.PackQuarantine
	reader  objectstore.Reader
}

var (
	_ objectstore.ObjectQuarantine = (*quarantine)(nil)
	_ objectstore.PackQuarantine   = (*quarantine)(nil)
	_ objectstore.Quarantine       = (*quarantine)(nil)
)

func newQuarantine(
	objectQ objectstore.ObjectQuarantine,
	packQ objectstore.PackQuarantine,
) *quarantine {
	return &quarantine{
		objectQ: objectQ,
		packQ:   packQ,
		reader:  objectmix.New(objectQ, packQ),
	}
}

// ReadBytesFull reads a full serialized object as "type size\0content" from
// either quarantined store.
func (quarantine *quarantine) ReadBytesFull(id objectid.ObjectID) ([]byte, error) {
	return quarantine.reader.ReadBytesFull(id)
}

// ReadBytesContent reads an object's type and content bytes from either
// quarantined store.
func (quarantine *quarantine) ReadBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	return quarantine.reader.ReadBytesContent(id)
}

// ReadReaderFull reads a full serialized object stream as
// "type size\0content" from either quarantined store.
func (quarantine *quarantine) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	return quarantine.reader.ReadReaderFull(id)
}

// ReadReaderContent reads an object's type, declared content length, and
// content stream from either quarantined store.
func (quarantine *quarantine) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	return quarantine.reader.ReadReaderContent(id)
}

// ReadSize reads an object's declared content length from either quarantined
// store.
func (quarantine *quarantine) ReadSize(id objectid.ObjectID) (int64, error) {
	return quarantine.reader.ReadSize(id)
}

// ReadHeader reads an object's type and declared content length from either
// quarantined store.
func (quarantine *quarantine) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	return quarantine.reader.ReadHeader(id)
}

// Refresh refreshes both quarantined stores and the combined quarantined reader.
func (quarantine *quarantine) Refresh() error {
	err := quarantine.objectQ.Refresh()
	if err != nil {
		return err
	}

	err = quarantine.packQ.Refresh()
	if err != nil {
		return err
	}

	return quarantine.reader.Refresh()
}

// WriteReaderContent writes one typed object content stream to the quarantined
// object-wise store.
func (quarantine *quarantine) WriteReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error) {
	return quarantine.objectQ.WriteReaderContent(ty, size, src)
}

// WriteReaderFull writes one full serialized object stream as
// "type size\0content" to the quarantined object-wise store.
func (quarantine *quarantine) WriteReaderFull(src io.Reader) (objectid.ObjectID, error) {
	return quarantine.objectQ.WriteReaderFull(src)
}

// WriteBytesContent writes one typed object content byte slice to the
// quarantined object-wise store.
func (quarantine *quarantine) WriteBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	return quarantine.objectQ.WriteBytesContent(ty, content)
}

// WriteBytesFull writes one full serialized object byte slice as
// "type size\0content" to the quarantined object-wise store.
func (quarantine *quarantine) WriteBytesFull(raw []byte) (objectid.ObjectID, error) {
	return quarantine.objectQ.WriteBytesFull(raw)
}

// WritePack ingests one pack stream into the quarantined pack-wise store.
func (quarantine *quarantine) WritePack(src io.Reader, opts objectstore.PackWriteOptions) error {
	return quarantine.packQ.WritePack(src, opts)
}
