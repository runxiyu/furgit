package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// ObjectType mirrors Git's object type tags.
type ObjectType uint8

const (
	ObjectTypeInvalid  ObjectType = 0
	ObjectTypeCommit   ObjectType = 1
	ObjectTypeTree     ObjectType = 2
	ObjectTypeBlob     ObjectType = 3
	ObjectTypeTag      ObjectType = 4
	ObjectTypeFuture   ObjectType = 5
	ObjectTypeOfsDelta ObjectType = 6
	ObjectTypeRefDelta ObjectType = 7
)

const (
	objectTypeNameBlob   = "blob"
	objectTypeNameTree   = "tree"
	objectTypeNameCommit = "commit"
	objectTypeNameTag    = "tag"
)

// Object describes any Git object variant.
type Object interface {
	ObjectType() ObjectType
}

// StoredObject describes a Git object with a known hash.
type StoredObject interface {
	Object
	Hash() Hash
}

func headerForType(ty ObjectType, body []byte) ([]byte, error) {
	var tyStr string
	switch ty {
	case ObjectTypeBlob:
		tyStr = objectTypeNameBlob
	case ObjectTypeTree:
		tyStr = objectTypeNameTree
	case ObjectTypeCommit:
		tyStr = objectTypeNameCommit
	case ObjectTypeTag:
		tyStr = objectTypeNameTag
	case ObjectTypeInvalid, ObjectTypeFuture, ObjectTypeOfsDelta, ObjectTypeRefDelta:
		return nil, fmt.Errorf("furgit: object: unsupported type %d", ty)
	default:
		return nil, fmt.Errorf("furgit: object: unsupported type %d", ty)
	}
	size := strconv.Itoa(len(body))
	var buf bytes.Buffer
	buf.Grow(len(tyStr) + len(size) + 1)
	buf.WriteString(tyStr)
	buf.WriteByte(' ')
	buf.WriteString(size)
	buf.WriteByte(0)
	return buf.Bytes(), nil
}

func parseObjectBody(ty ObjectType, id Hash, body []byte, repo *Repository) (StoredObject, error) {
	switch ty {
	case ObjectTypeBlob:
		return parseBlob(id, body)
	case ObjectTypeTree:
		return parseTree(id, body, repo)
	case ObjectTypeCommit:
		return parseCommit(id, body, repo)
	case ObjectTypeTag:
		return parseTag(id, body, repo)
	case ObjectTypeInvalid, ObjectTypeFuture, ObjectTypeOfsDelta, ObjectTypeRefDelta:
		return nil, fmt.Errorf("furgit: object: unsupported type %d", ty)
	default:
		return nil, fmt.Errorf("furgit: object: unknown type %d", ty)
	}
}

// ReadObject resolves an ID by consulting loose then packed storage.
func (repo *Repository) ReadObject(id Hash) (Object, error) {
	obj, err := repo.looseRead(id)
	if err == nil {
		return obj, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	obj, err = repo.packRead(id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidObject
	}
	return obj, err
}

// ReadObjectTypeSize reports the object type and size without inflating the body.
func (repo *Repository) ReadObjectTypeSize(id Hash) (ObjectType, int64, error) {
	ty, size, err := repo.looseTypeSize(id)
	if err == nil {
		return ty, size, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ObjectTypeInvalid, 0, err
	}
	loc, err := repo.packIndexFind(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ObjectTypeInvalid, 0, ErrInvalidObject
		}
		return ObjectTypeInvalid, 0, err
	}
	return repo.packTypeSizeAtLocation(loc, nil)
}
