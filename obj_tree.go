package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Tree represents a Git tree object.
type Tree[T HashType] struct {
	Hash    Hash[T]
	Entries []TreeEntry[T]
}

// TreeEntry represents a single entry in a Git tree.
type TreeEntry[T HashType] struct {
	Mode uint32
	Name []byte
	ID   Hash[T]
}

// ObjType allows Tree to satisfy the Object interface.
func (*Tree[T]) ObjType() ObjType {
	return ObjTree
}

// parseTree decodes a tree body.
func parseTree[T HashType](id Hash[T], body []byte) (*Tree[T], error) {
	var entries []TreeEntry[T]
	hashSize := hashLen[T]()
	i := 0
	for i < len(body) {
		space := bytes.IndexByte(body[i:], ' ')
		if space < 0 {
			return nil, errors.New("furgit: tree: missing mode terminator")
		}
		modeBytes := body[i : i+space]
		i += space + 1

		nul := bytes.IndexByte(body[i:], 0)
		if nul < 0 {
			return nil, errors.New("furgit: tree: missing name terminator")
		}
		nameBytes := body[i : i+nul]
		i += nul + 1

		if i+hashSize > len(body) {
			return nil, errors.New("furgit: tree: truncated child hash")
		}
		var child Hash[T]
		copy(child.Slice(), body[i:i+hashSize])
		i += hashSize

		mode, err := strconv.ParseUint(string(modeBytes), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("furgit: tree: parse mode: %w", err)
		}

		entry := TreeEntry[T]{
			Mode: uint32(mode),
			Name: append([]byte(nil), nameBytes...),
			ID:   child,
		}
		entries = append(entries, entry)
	}

	return &Tree[T]{
		Hash:    id,
		Entries: entries,
	}, nil
}

// treeBody builds the entry list for a tree without the Git header.
func treeBody[T HashType](t *Tree[T]) []byte {
	hashSize := hashLen[T]()
	var bodyLen int
	for _, e := range t.Entries {
		mode := strconv.FormatUint(uint64(e.Mode), 8)
		bodyLen += len(mode) + 1 + len(e.Name) + 1 + hashSize
	}

	body := make([]byte, bodyLen)
	pos := 0
	for _, e := range t.Entries {
		mode := strconv.FormatUint(uint64(e.Mode), 8)
		pos += copy(body[pos:], []byte(mode))
		body[pos] = ' '
		pos++
		pos += copy(body[pos:], e.Name)
		body[pos] = 0
		pos++
		pos += copy(body[pos:], e.ID.Slice()[:hashSize])
	}

	return body
}

// Serialize renders a Tree into canonical Git format.
func (t *Tree[T]) Serialize() ([]byte, error) {
	body := treeBody(t)
	header, err := headerForType(ObjTree, body)
	if err != nil {
		return nil, err
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}

// Entry looks up a tree entry by name.
func (t *Tree[T]) Entry(name []byte) *TreeEntry[T] {
	low, high := 0, len(t.Entries)-1
	for low <= high {
		mid := (low + high) / 2
		cmp := bytes.Compare(t.Entries[mid].Name, name)
		if cmp == 0 {
			return &t.Entries[mid]
		} else if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return nil
}
