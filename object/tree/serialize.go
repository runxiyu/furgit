package tree

import (
	"errors"
	"strconv"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// BytesWithoutHeader renders the raw tree body bytes.
func (tree *Tree) BytesWithoutHeader() ([]byte, error) {
	var bodyLen int

	for _, entry := range tree.Entries {
		mode := strconv.FormatUint(uint64(entry.Mode), 8)
		bodyLen += len(mode) + 1 + len(entry.Name) + 1 + entry.ID.Algorithm().Size()
	}

	body := make([]byte, bodyLen)
	pos := 0

	for _, entry := range tree.Entries {
		mode := strconv.FormatUint(uint64(entry.Mode), 8)
		pos += copy(body[pos:], mode)
		body[pos] = ' '
		pos++
		pos += copy(body[pos:], entry.Name)
		body[pos] = 0
		pos++
		id := entry.ID.Bytes()
		pos += copy(body[pos:], id)
	}

	return body, nil
}

// BytesWithHeader renders the raw object (header + body).
func (tree *Tree) BytesWithHeader() ([]byte, error) {
	body, err := tree.BytesWithoutHeader()
	if err != nil {
		return nil, err
	}

	header, ok := objectheader.Encode(objecttype.TypeTree, int64(len(body)))
	if !ok {
		return nil, errors.New("object: tree: failed to encode object header")
	}

	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)

	return raw, nil
}
