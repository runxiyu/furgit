package object

import (
	"strconv"

	"codeberg.org/lindenii/furgit/objecttype"
)

func (tree *Tree) serialize() []byte {
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

	return body
}

// Serialize renders the raw object (header + body).
func (tree *Tree) Serialize() ([]byte, error) {
	body := tree.serialize()
	header, err := headerForType(objecttype.TypeTree, body)
	if err != nil {
		return nil, err
	}
	raw := make([]byte, len(header)+len(body))
	copy(raw, header)
	copy(raw[len(header):], body)
	return raw, nil
}
