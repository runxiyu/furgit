package tree

import (
	"bytes"
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree/mode"
)

// Parse decodes a tree object body into a fully materialized Tree.
//
// It enforces canonical modes, nonempty slash-free names,
// correctly sized object IDs, and strictly increasing Git tree order.
// It does not enforce fsck-level name policy
// (for example ".", "..", ".git", or platform-specific aliases).
func Parse(body []byte, objectFormat id.ObjectFormat) (*Tree, error) {
	tree := new(Tree)
	idSize := objectFormat.Size()
	seen := make(map[string]struct{})

	i := 0
	for i < len(body) {
		space := bytes.IndexByte(body[i:], ' ')
		if space < 0 {
			return nil, fmt.Errorf("%w: missing mode terminator at offset %d", ErrInvalidTree, i)
		}

		entryMode, err := mode.Parse(body[i : i+space])
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidTree, err)
		}

		i += space + 1

		nul := bytes.IndexByte(body[i:], 0)
		if nul < 0 {
			return nil, fmt.Errorf("%w: missing name terminator at offset %d", ErrInvalidTree, i)
		}

		name := string(body[i : i+nul])
		i += nul + 1

		if err := validateName(name); err != nil {
			return nil, err
		}

		idEnd := i + idSize
		if idEnd > len(body) {
			return nil, fmt.Errorf("%w: truncated object id at offset %d", ErrInvalidTree, i)
		}

		oid, err := objectFormat.FromBytes(body[i:idEnd])
		if err != nil {
			return nil, fmt.Errorf("%w: object id at offset %d: %w", ErrInvalidTree, i, err)
		}

		i = idEnd

		entry := Entry{Mode: entryMode, Name: name, ID: oid}

		if len(tree.entries) > 0 {
			prev := tree.entries[len(tree.entries)-1]
			if nameCompare(prev.Name, prev.Mode == mode.Directory, entry.Name, entryMode == mode.Directory) >= 0 {
				return nil, fmt.Errorf("%w: entry %q out of order or duplicated", ErrInvalidTree, entry.Name)
			}
		}

		if _, dup := seen[entry.Name]; dup {
			return nil, fmt.Errorf("%w: duplicate entry name %q", ErrInvalidTree, entry.Name)
		}

		seen[entry.Name] = struct{}{}

		tree.entries = append(tree.entries, entry)
	}

	return tree, nil
}
