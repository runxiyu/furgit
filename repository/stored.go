package repository

import (
	"fmt"

	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstored"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadStored reads, parses, and wraps one object by ID.
//
//nolint:ireturn
func (repo *Repository) ReadStored(id objectid.ObjectID) (objectstored.StoredObject, error) {
	parsed, err := repo.readParsedObject(id)
	if err != nil {
		return nil, err
	}

	switch parsed := parsed.(type) {
	case *object.Blob:
		return objectstored.NewStoredBlob(id, parsed), nil
	case *object.Tree:
		return objectstored.NewStoredTree(id, parsed), nil
	case *object.Commit:
		return objectstored.NewStoredCommit(id, parsed), nil
	case *object.Tag:
		return objectstored.NewStoredTag(id, parsed), nil
	default:
		return nil, fmt.Errorf("repository: unsupported parsed object type %T", parsed)
	}
}

// ReadStoredBlob reads and parses a blob object by ID.
func (repo *Repository) ReadStoredBlob(id objectid.ObjectID) (*objectstored.StoredBlob, error) {
	stored, err := repo.ReadStored(id)
	if err != nil {
		return nil, err
	}

	blob, ok := stored.(*objectstored.StoredBlob)
	if !ok {
		return nil, fmt.Errorf("repository: expected blob object %s, got %v", id, stored.Object().ObjectType())
	}

	return blob, nil
}

// ReadStoredTree reads and parses a tree object by ID.
func (repo *Repository) ReadStoredTree(id objectid.ObjectID) (*objectstored.StoredTree, error) {
	stored, err := repo.ReadStored(id)
	if err != nil {
		return nil, err
	}

	tree, ok := stored.(*objectstored.StoredTree)
	if !ok {
		return nil, fmt.Errorf("repository: expected tree object %s, got %v", id, stored.Object().ObjectType())
	}

	return tree, nil
}

// ReadStoredCommit reads and parses a commit object by ID.
func (repo *Repository) ReadStoredCommit(id objectid.ObjectID) (*objectstored.StoredCommit, error) {
	stored, err := repo.ReadStored(id)
	if err != nil {
		return nil, err
	}

	commit, ok := stored.(*objectstored.StoredCommit)
	if !ok {
		return nil, fmt.Errorf("repository: expected commit object %s, got %v", id, stored.Object().ObjectType())
	}

	return commit, nil
}

// ReadStoredTag reads and parses a tag object by ID.
func (repo *Repository) ReadStoredTag(id objectid.ObjectID) (*objectstored.StoredTag, error) {
	stored, err := repo.ReadStored(id)
	if err != nil {
		return nil, err
	}

	tag, ok := stored.(*objectstored.StoredTag)
	if !ok {
		return nil, fmt.Errorf("repository: expected tag object %s, got %v", id, stored.Object().ObjectType())
	}

	return tag, nil
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
