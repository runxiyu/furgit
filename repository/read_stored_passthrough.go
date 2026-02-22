package repository

import (
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadStoredHeader reads an object's type and declared content length.
func (repo *Repository) ReadStoredHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	return repo.objects.ReadHeader(id)
}

// ReadStoredSize reads an object's declared content length.
func (repo *Repository) ReadStoredSize(id objectid.ObjectID) (int64, error) {
	return repo.objects.ReadSize(id)
}

// ReadStoredBytesFull reads a full serialized object as "type size\0content".
func (repo *Repository) ReadStoredBytesFull(id objectid.ObjectID) ([]byte, error) {
	return repo.objects.ReadBytesFull(id)
}

// ReadStoredBytesContent reads an object's type and content bytes.
func (repo *Repository) ReadStoredBytesContent(id objectid.ObjectID) (objecttype.Type, []byte, error) {
	return repo.objects.ReadBytesContent(id)
}

// ReadStoredReaderFull reads a full serialized object stream.
//
// Caller must close the returned reader.
func (repo *Repository) ReadStoredReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	return repo.objects.ReadReaderFull(id)
}

// ReadStoredReaderContent reads an object's type, declared content length, and
// content stream.
//
// Caller must close the returned reader.
func (repo *Repository) ReadStoredReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	return repo.objects.ReadReaderContent(id)
}
