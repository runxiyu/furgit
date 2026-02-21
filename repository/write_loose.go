package repository

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// WriteLooseBytesFull writes one loose object from raw "type size\0content".
func (repo *Repository) WriteLooseBytesFull(raw []byte) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteBytesFull(raw)
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose full bytes: %w", err)
	}
	return id, nil
}

// WriteLooseBytesContent writes one loose object from typed content bytes.
func (repo *Repository) WriteLooseBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteBytesContent(ty, content)
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose content bytes: %w", err)
	}
	return id, nil
}

// WriteLooseWriterFull returns a writer for one full serialized object stream.
//
// The caller must close the writer, then call finalize to publish the object.
func (repo *Repository) WriteLooseWriterFull() (io.WriteCloser, func() (objectid.ObjectID, error), error) {
	writer, finalize, err := repo.objectsLooseForWritingOnly.WriteWriterFull()
	if err != nil {
		return nil, nil, fmt.Errorf("repository: create loose full writer: %w", err)
	}
	return writer, finalize, nil
}

// WriteLooseWriterContent returns a writer for one typed object content stream.
//
// The caller must write exactly size bytes, close the writer, then call
// finalize to publish the object.
func (repo *Repository) WriteLooseWriterContent(ty objecttype.Type, size int64) (io.WriteCloser, func() (objectid.ObjectID, error), error) {
	writer, finalize, err := repo.objectsLooseForWritingOnly.WriteWriterContent(ty, size)
	if err != nil {
		return nil, nil, fmt.Errorf("repository: create loose content writer: %w", err)
	}
	return writer, finalize, nil
}
