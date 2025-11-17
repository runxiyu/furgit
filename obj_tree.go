package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

// Tree represents a Git tree object.
type Tree struct {
	// Entries represents the entries in the tree.
	Entries []TreeEntry
}

// StoredTree represents a tree stored in the object database.
type StoredTree struct {
	Tree
	hash Hash
}

// Hash returns the hash of the stored tree.
func (sTree *StoredTree) Hash() Hash {
	return sTree.hash
}

// FileMode represents the mode of a file in a Git tree.
type FileMode uint32

const (
	// FileModeDir represents a directory (tree) in a Git tree.
	FileModeDir FileMode = 0o40000
	// FileModeRegular represents a regular file (blob) in a Git tree.
	FileModeRegular FileMode = 0o100644
	// FileModeExecutable represents an executable file (blob) in a Git tree.
	FileModeExecutable FileMode = 0o100755
	// FileModeSymlink represents a symbolic link (blob) in a Git tree.
	FileModeSymlink FileMode = 0o120000
	// FileModeGitlink represents a Git link (submodule) in a Git tree.
	FileModeGitlink FileMode = 0o160000
)

// TreeEntry represents a single entry in a Git tree.
type TreeEntry struct {
	// Mode represents the file mode of the entry.
	Mode FileMode
	// Name represents the name of the entry.
	Name []byte
	// ID represents the hash of the entry. This is typically
	// either a blob or a tree.
	ID Hash
}

// ObjectType returns the object type of the tree.
//
// It always returns ObjectTypeTree.
func (tree *Tree) ObjectType() ObjectType {
	_ = tree
	return ObjectTypeTree
}

// parseTree decodes a tree body.
func parseTree(id Hash, body []byte, repo *Repository) (*StoredTree, error) {
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
			Mode: FileMode(mode),
			Name: append([]byte(nil), nameBytes...),
			ID:   child,
		}
		entries = append(entries, entry)
	}

	return &StoredTree{
		hash: id,
		Tree: Tree{
			Entries: entries,
		},
	}, nil
}

// treeBody builds the entry list for a tree without the Git header.
func (tree *Tree) serialize() []byte {
	var bodyLen int
	for _, e := range tree.Entries {
		mode := strconv.FormatUint(uint64(e.Mode), 8)
		bodyLen += len(mode) + 1 + len(e.Name) + 1 + e.ID.size
	}

	body := make([]byte, bodyLen)
	pos := 0
	for _, e := range tree.Entries {
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

// Serialize renders the tree into its raw byte representation,
// including the header (i.e., "type size\0").
func (tree *Tree) Serialize() ([]byte, error) {
	body := tree.serialize()
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
//
// Lookups are not recursive.
// It returns nil if no such entry exists.
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

// EntryRecursive looks up a tree entry by path.
//
// Lookups are recursive.
func (sTree *StoredTree) EntryRecursive(repo *Repository, path [][]byte) (*TreeEntry, error) {
	if len(path) == 0 {
		return nil, errors.New("furgit: tree: empty path")
	}

	currentTree := sTree
	for i, part := range path {
		entry := currentTree.Entry(part)
		if entry == nil {
			return nil, ErrNotFound
		}
		if i == len(path)-1 {
			return entry, nil
		}
		obj, err := repo.ReadObject(entry.ID)
		if err != nil {
			return nil, err
		}
		nextTree, ok := obj.(*StoredTree)
		if !ok {
			return nil, fmt.Errorf("furgit: tree: expected tree object at %s, got %T", part, obj)
			// TODO: It may be useful to check the mode instead of reporting
			// an object type error.
		}
		currentTree = nextTree
	}

	return nil, ErrNotFound
}
