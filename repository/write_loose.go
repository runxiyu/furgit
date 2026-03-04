package repository

import (
	"bytes"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// WriteLooseBytesFull writes one loose object from raw "type size\0content".
func (repo *Repository) WriteLooseBytesFull(raw []byte) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteReaderFull(bytes.NewReader(raw))
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose full bytes: %w", err)
	}

	return id, nil
}

// WriteLooseBytesContent writes one loose object from typed content bytes.
func (repo *Repository) WriteLooseBytesContent(ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteReaderContent(ty, int64(len(content)), bytes.NewReader(content))
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose content bytes: %w", err)
	}

	return id, nil
}

// WriteLooseReaderFull writes one loose object from raw bytes
// "type size\0content" read from src.
func (repo *Repository) WriteLooseReaderFull(src io.Reader) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteReaderFull(src)
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose full reader: %w", err)
	}

	return id, nil
}

// WriteLooseReaderContent writes one loose object from typed content bytes read
// from src. src must provide exactly size bytes.
func (repo *Repository) WriteLooseReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error) {
	id, err := repo.objectsLooseForWritingOnly.WriteReaderContent(ty, size, src)
	if err != nil {
		return objectid.ObjectID{}, fmt.Errorf("repository: write loose content reader: %w", err)
	}

	return id, nil
}
