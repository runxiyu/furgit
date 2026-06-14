package tree_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
)

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			size := objectFormat.Size()

			record := func(mode, name string, idLen int) []byte {
				var b bytes.Buffer
				b.WriteString(mode)
				b.WriteByte(' ')
				b.WriteString(name)
				b.WriteByte(0)
				b.Write(make([]byte, idLen))

				return b.Bytes()
			}

			for _, tc := range []struct {
				name string
				body []byte
			}{
				{name: "malformed-mode", body: record("10064x", "file", size)},
				{name: "zero-padded-mode", body: record("0100644", "file", size)},
				{name: "unsupported-mode", body: record("100640", "file", size)},
				{name: "empty-name", body: record("100644", "", size)},
				{name: "slash-name", body: record("100644", "a/b", size)},
				{name: "truncated-id", body: record("100644", "file", size-1)},
				{name: "missing-mode-terminator", body: []byte("100644")},
				{name: "missing-name-terminator", body: []byte("100644 file")},
				{name: "unsorted", body: append(record("100644", "b", size), record("100644", "a", size)...)},
				{name: "duplicate", body: append(record("100644", "a", size), record("100644", "a", size)...)},
				{name: "conflicting-tree-blob", body: append(record("100644", "foo", size), record("40000", "foo", size)...)},
				{name: "conflicting-tree-blob-nonadjacent", body: append(append(record("100644", "foo", size), record("100644", "foo.c", size)...), record("40000", "foo", size)...)},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					_, err := tree.Parse(tc.body, objectFormat)
					if !errors.Is(err, tree.ErrInvalidTree) {
						t.Fatalf("Parse error = %v, want ErrInvalidTree", err)
					}
				})
			}
		})
	}
}
