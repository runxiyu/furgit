// Package object provides Git object models and codecs.
package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	// ErrInvalidObject indicates malformed serialized data.
	ErrInvalidObject = errors.New("object: invalid object encoding")
	// ErrNotFound indicates missing entries in in-memory lookups.
	ErrNotFound = errors.New("object: not found")
)

// Type mirrors Git object type tags.
type Type uint8

const (
	TypeInvalid  Type = 0
	TypeCommit   Type = 1
	TypeTree     Type = 2
	TypeBlob     Type = 3
	TypeTag      Type = 4
	TypeFuture   Type = 5
	TypeOfsDelta Type = 6
	TypeRefDelta Type = 7
)

const (
	typeNameBlob   = "blob"
	typeNameTree   = "tree"
	typeNameCommit = "commit"
	typeNameTag    = "tag"
)

// Object is a Git object that can serialize itself.
type Object interface {
	ObjectType() Type
	Serialize() ([]byte, error)
}

// ParseTypeName parses a canonical Git object type name.
func ParseTypeName(name string) (Type, error) {
	switch name {
	case typeNameBlob:
		return TypeBlob, nil
	case typeNameTree:
		return TypeTree, nil
	case typeNameCommit:
		return TypeCommit, nil
	case typeNameTag:
		return TypeTag, nil
	default:
		return TypeInvalid, ErrInvalidObject
	}
}

func typeName(ty Type) (string, error) {
	switch ty {
	case TypeBlob:
		return typeNameBlob, nil
	case TypeTree:
		return typeNameTree, nil
	case TypeCommit:
		return typeNameCommit, nil
	case TypeTag:
		return typeNameTag, nil
	default:
		return "", fmt.Errorf("object: unsupported type %d", ty)
	}
}

func headerForType(ty Type, body []byte) ([]byte, error) {
	tyStr, err := typeName(ty)
	if err != nil {
		return nil, err
	}
	size := strconv.Itoa(len(body))
	var buf bytes.Buffer
	buf.Grow(len(tyStr) + len(size) + 2)
	buf.WriteString(tyStr)
	buf.WriteByte(' ')
	buf.WriteString(size)
	buf.WriteByte(0)
	return buf.Bytes(), nil
}
