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
	// An invalid object.
	ObjectTypeInvalid ObjectType = 0
	// A commit object.
	ObjectTypeCommit ObjectType = 1
	// A tree object.
	ObjectTypeTree ObjectType = 2
	// A blob object.
	ObjectTypeBlob ObjectType = 3
	// An annotated tag object.
	ObjectTypeTag ObjectType = 4
	// An object type reserved for future use.
	ObjectTypeFuture ObjectType = 5
	// A packfile offset delta object. This is not typically exposed.
	ObjectTypeOfsDelta ObjectType = 6
	// A packfile reference delta object. This is not typically exposed.
	ObjectTypeRefDelta ObjectType = 7
)

const (
	objectTypeNameBlob   = "blob"
	objectTypeNameTree   = "tree"
	objectTypeNameCommit = "commit"
	objectTypeNameTag    = "tag"
)

// Object represents a Git object.
type Object interface {
	// ObjectType returns the object's type.
	ObjectType() ObjectType
	// Serialize renders the object into its raw byte representation,
	// including the header (i.e., "type size\0").
	Serialize() ([]byte, error)
}

// StoredObject describes a Git object with a known hash, such as
// one read from storage.
type StoredObject interface {
	Object
	// Hash returns the object's hash.
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

// ReadObject resolves an ID.
func (repo *Repository) ReadObject(id Hash) (StoredObject, error) {
	ty, body, err := repo.looseRead(id)
	if err == nil {
		obj, parseErr := parseObjectBody(ty, id, body.Bytes(), repo)
		body.Release()
		return obj, parseErr
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	ty, body, err = repo.packRead(id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	obj, parseErr := parseObjectBody(ty, id, body.Bytes(), repo)
	body.Release()
	return obj, parseErr
}

// ReadObjectTypeRaw reads the object type and raw body.
func (repo *Repository) ReadObjectTypeRaw(id Hash) (ObjectType, []byte, error) {
	ty, body, err := repo.looseRead(id)
	if err == nil {
		return ty, body.Bytes(), nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ObjectTypeInvalid, nil, err
	}
	ty, body, err = repo.packRead(id)
	if errors.Is(err, ErrNotFound) {
		return ObjectTypeInvalid, nil, ErrNotFound
	}
	if err != nil {
		return ObjectTypeInvalid, nil, err
	}
	return ty, body.Bytes(), nil
	// note to self: It always feels wrong to not call .Release in places like
	// this but this is actually correct; we're returning the underlying buffer
	// to the user who should not be aware of our internal buffer pooling.
	// Releasing this buffer back to the pool would lead to a use-after-free;
	// not releasing it as we do here, means it gets GC'ed.
	// Copying into a newly allocated buffer is even worse as it incurs
	// unnecessary copy overhead.
}

// ReadObjectTypeSize reports the object type and size.
//
// Typicall, this is more efficient than reading the full object,
// as it avoids decompressing the entire object body.
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
