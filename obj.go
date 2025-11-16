package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// ObjType mirrors Git's object type tags.
type ObjType uint8

const (
	ObjInvalid  ObjType = 0
	ObjCommit   ObjType = 1
	ObjTree     ObjType = 2
	ObjBlob     ObjType = 3
	ObjTag      ObjType = 4
	ObjFuture   ObjType = 5
	ObjOfsDelta ObjType = 6
	ObjRefDelta ObjType = 7
)

const (
	objNameBlob   = "blob"
	objNameTree   = "tree"
	objNameCommit = "commit"
	objNameTag    = "tag"
)

// Object describes any Git object variant.
type Object interface {
	ObjType() ObjType
}

func computeRawHash(data []byte, hashSize int) Hash {
	hashFunc := hashFuncs[hashSize]
	return hashFunc(data)
}

func headerForType(ty ObjType, body []byte) ([]byte, error) {
	var tyStr string
	switch ty {
	case ObjBlob:
		tyStr = objNameBlob
	case ObjTree:
		tyStr = objNameTree
	case ObjCommit:
		tyStr = objNameCommit
	case ObjTag:
		tyStr = objNameTag
	case ObjInvalid, ObjFuture, ObjOfsDelta, ObjRefDelta:
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

func verifyRawObject(buf []byte, want Hash, hashSize int) bool {
	return computeRawHash(buf, hashSize) == want
}

func verifyTypedObject(ty ObjType, body []byte, want Hash, hashSize int) bool {
	header, err := headerForType(ty, body)
	if err != nil {
		return false
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return computeRawHash(raw, hashSize) == want
}

func parseObjectBody(ty ObjType, id Hash, body []byte, hashSize int) (Object, error) {
	switch ty {
	case ObjBlob:
		return parseBlob(id, body)
	case ObjTree:
		return parseTree(id, body, hashSize)
	case ObjCommit:
		return parseCommit(id, body, hashSize)
	case ObjTag:
		return parseTag(id, body, hashSize)
	case ObjInvalid, ObjFuture, ObjOfsDelta, ObjRefDelta:
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
func (repo *Repository) ReadObjectTypeSize(id Hash) (ObjType, int64, error) {
	ty, size, err := repo.looseTypeSize(id)
	if err == nil {
		return ty, size, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ObjInvalid, 0, err
	}
	loc, err := repo.packIndexFind(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ObjInvalid, 0, ErrInvalidObject
		}
		return ObjInvalid, 0, err
	}
	return repo.packTypeSizeAtLocation(loc, nil)
}
