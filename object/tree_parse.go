package object

import (
	"bytes"
	"fmt"
	"strconv"

	"codeberg.org/lindenii/furgit/oid"
)

// ParseTree decodes a tree object body.
func ParseTree(body []byte, algo oid.Algorithm) (*Tree, error) {
	if algo.Size() == 0 {
		return nil, ErrInvalidObject
	}

	var entries []TreeEntry
	i := 0
	for i < len(body) {
		space := bytes.IndexByte(body[i:], ' ')
		if space < 0 {
			return nil, fmt.Errorf("object: tree: missing mode terminator")
		}
		modeBytes := body[i : i+space]
		i += space + 1

		nul := bytes.IndexByte(body[i:], 0)
		if nul < 0 {
			return nil, fmt.Errorf("object: tree: missing name terminator")
		}
		nameBytes := body[i : i+nul]
		i += nul + 1

		idEnd := i + algo.Size()
		if idEnd > len(body) {
			return nil, fmt.Errorf("object: tree: truncated child object id")
		}
		id, err := oid.FromBytes(algo, body[i:idEnd])
		if err != nil {
			return nil, err
		}
		i = idEnd

		mode, err := strconv.ParseUint(string(modeBytes), 8, 32)
		if err != nil {
			return nil, fmt.Errorf("object: tree: parse mode: %w", err)
		}

		entries = append(entries, TreeEntry{
			Mode: FileMode(mode),
			Name: append([]byte(nil), nameBytes...),
			ID:   id,
		})
	}

	return &Tree{Entries: entries}, nil
}
