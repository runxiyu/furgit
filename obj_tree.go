package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Tree represents a Git tree object.
type Tree struct {
	Hash    Hash
	Entries []TreeEntry
}

// TreeEntry represents a single entry in a Git tree.
type TreeEntry struct {
	Mode uint32
	Name []byte
	ID   Hash
}

// ObjectType allows Tree to satisfy the Object interface.
func (tree *Tree) ObjectType() ObjectType {
	_ = tree
	return ObjectTypeTree
}

// parseTree decodes a tree body.
func parseTree(id Hash, body []byte, repo *Repository) (*Tree, error) {
	var entries []TreeEntry
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

		if i+repo.hashSize > len(body) {
			return nil, errors.New("furgit: tree: truncated child hash")
		}
		var child Hash
		copy(child.data[:], body[i:i+repo.hashSize])
		child.size = repo.hashSize
		i += repo.hashSize

		mode, err := strconv.ParseUint(string(modeBytes), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("furgit: tree: parse mode: %w", err)
		}

		entry := TreeEntry{
			Mode: uint32(mode),
			Name: append([]byte(nil), nameBytes...),
			ID:   child,
		}
		entries = append(entries, entry)
	}

	return &Tree{
		Hash:    id,
		Entries: entries,
	}, nil
}

// treeBody builds the entry list for a tree without the Git header.
func treeBody(t *Tree) []byte {
	var bodyLen int
	for _, e := range t.Entries {
		mode := strconv.FormatUint(uint64(e.Mode), 8)
		bodyLen += len(mode) + 1 + len(e.Name) + 1 + e.ID.size
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
		pos += copy(body[pos:], e.ID.data[:e.ID.size])
	}

	return body
}

// Serialize renders a Tree into canonical Git format.
func (tree *Tree) Serialize() ([]byte, error) {
	body := treeBody(tree)
	header, err := headerForType(ObjectTypeTree, body)
	if err != nil {
		return nil, err
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}

// Entry looks up a tree entry by name.
func (tree *Tree) Entry(name []byte) *TreeEntry {
	low, high := 0, len(tree.Entries)-1
	for low <= high {
		mid := (low + high) / 2
		cmp := bytes.Compare(tree.Entries[mid].Name, name)
		if cmp == 0 {
			return &tree.Entries[mid]
		} else if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return nil
}
