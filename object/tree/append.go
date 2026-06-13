package tree

import (
	"slices"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/tree/mode"
	"lindenii.org/go/furgit/object/typ"
)

// AppendWithoutHeader renders the raw tree body bytes.
//
// It trusts the package invariant that the entries are valid and ordered,
// so it does not revalidate them.
func (tree *Tree) AppendWithoutHeader(dst []byte) ([]byte, error) {
	var bodyLen int
	for _, entry := range tree.entries {
		bodyLen += mode.MaxModeDigits + 1 + len(entry.Name) + 1 + entry.ID.ObjectFormat().Size()
	}

	dst = slices.Grow(dst, bodyLen)

	for _, entry := range tree.entries {
		dst = entry.Mode.Append(dst)
		dst = append(dst, ' ')
		dst = append(dst, entry.Name...)
		dst = append(dst, 0)
		dst = append(dst, entry.ID.RawBytes()...)
	}

	return dst, nil
}

// AppendWithHeader renders the raw object (header + body).
func (tree *Tree) AppendWithHeader(dst []byte) ([]byte, error) {
	// TODO: Try to not allocate?
	body, err := tree.AppendWithoutHeader(nil)
	if err != nil {
		return dst, err
	}

	dst = header.Append(dst, typ.Tree, len(body))

	return append(dst, body...), nil
}
