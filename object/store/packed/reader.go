package packed

import (
	"io"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads a full serialized object as "type size\x00content".
func (packed *Packed) ReadBytesFull(objectID id.ObjectID) ([]byte, error)

// ReadBytesContent reads an object's type and content bytes.
func (packed *Packed) ReadBytesContent(objectID id.ObjectID) (typ.Type, []byte, error)

// ReadReaderFull reads a full serialized object stream as "type size\x00content".
func (packed *Packed) ReadReaderFull(objectID id.ObjectID) (io.ReadCloser, error)

// ReadReaderContent reads an object's type, declared content length,
// and content stream.
func (packed *Packed) ReadReaderContent(objectID id.ObjectID) (typ.Type, uint64, io.ReadCloser, error)

// ReadSize reads an object's declared content length.
func (packed *Packed) ReadSize(objectID id.ObjectID) (uint64, error)

// ReadHeader reads an object's type and declared content length.
func (packed *Packed) ReadHeader(objectID id.ObjectID) (typ.Type, uint64, error)

// Refresh updates the packed-store view of on-disk pack/index candidates.
func (packed *Packed) Refresh() error
