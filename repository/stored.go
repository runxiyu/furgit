package repository

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/object/stored"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadStored reads, parses, and wraps one object by ID.
func (repo *Repository) ReadStored(id objectid.ObjectID) (*stored.Stored[object.Object], error) {
	parsed, err := repo.readParsedObject(id)
	if err != nil {
		return nil, err
	}

	return stored.New(id, parsed), nil
}

// ReadStoredBlob reads and parses a blob object by ID.
func (repo *Repository) ReadStoredBlob(id objectid.ObjectID) (*stored.Stored[*object.Blob], error) {
	return readStoredAs[*object.Blob](repo, id)
}

// ReadStoredTree reads and parses a tree object by ID.
func (repo *Repository) ReadStoredTree(id objectid.ObjectID) (*stored.Stored[*object.Tree], error) {
	return readStoredAs[*object.Tree](repo, id)
}

// ReadStoredCommit reads and parses a commit object by ID.
func (repo *Repository) ReadStoredCommit(id objectid.ObjectID) (*stored.Stored[*object.Commit], error) {
	return readStoredAs[*object.Commit](repo, id)
}

// ReadStoredTag reads and parses a tag object by ID.
func (repo *Repository) ReadStoredTag(id objectid.ObjectID) (*stored.Stored[*object.Tag], error) {
	return readStoredAs[*object.Tag](repo, id)
}

// readParsedObject reads bytes content from storage and parses one object.
//
//nolint:ireturn
func (repo *Repository) readParsedObject(id objectid.ObjectID) (object.Object, error) {
	ty, content, err := repo.objects.ReadBytesContent(id)
	if err != nil {
		return nil, err
	}

	parsed, err := object.ParseObjectWithoutHeader(ty, content, repo.algo)
	if err != nil {
		tyName, ok := objecttype.Name(ty)
		if !ok {
			tyName = fmt.Sprintf("type %d", ty)
		}

		return nil, fmt.Errorf("repository: parse object %s (%s): %w", id, tyName, err)
	}

	return parsed, nil
}

func readStoredAs[T object.Object](repo *Repository, id objectid.ObjectID) (*stored.Stored[T], error) {
	parsed, err := repo.readParsedObject(id)
	if err != nil {
		return nil, err
	}

	typed, ok := parsed.(T)
	if !ok {
		return nil, fmt.Errorf("repository: expected %T object %s, got %v", *new(T), id, parsed.ObjectType())
	}

	return stored.New(id, typed), nil
}
