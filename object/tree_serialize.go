package object

import (
	"strconv"

	"codeberg.org/lindenii/furgit/internal/objectheader"
	"codeberg.org/lindenii/furgit/objecttype"
)

// SerializeWithoutHeader renders the raw tree body bytes.
func (tree *Tree) SerializeWithoutHeader() ([]byte, error) {
	var bodyLen int
	for _, entry := range tree.Entries {
		mode := strconv.FormatUint(uint64(entry.Mode), 8)
		bodyLen += len(mode) + 1 + len(entry.Name) + 1 + entry.ID.Size()
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

// SerializeWithHeader renders the raw object (header + body).
func (tree *Tree) SerializeWithHeader() ([]byte, error) {
	body, err := tree.SerializeWithoutHeader()
	if err != nil {
		return nil, err
	}
	header, ok := objectheader.Encode(objecttype.TypeTree, int64(len(body)))
	if !ok {
		return nil, ErrInvalidObject
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
