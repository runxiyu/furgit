package tree_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tree"
	"lindenii.org/go/furgit/object/tree/mode"
)

func TestInsertRejects(t *testing.T) {
	t.Parallel()

	zero := id.SupportedObjectFormats()[0].Zero()

	for _, tc := range []struct {
		name  string
		entry tree.Entry
	}{
		{name: "empty-name", entry: tree.Entry{Mode: mode.Regular, Name: "", ID: zero}},
		{name: "slash-name", entry: tree.Entry{Mode: mode.Regular, Name: "a/b", ID: zero}},
		{name: "nul-name", entry: tree.Entry{Mode: mode.Regular, Name: "a\x00b", ID: zero}},
		{name: "invalid-mode", entry: tree.Entry{Mode: mode.Mode(0o100640), Name: "file", ID: zero}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr tree.Tree

			err := tr.Insert(tc.entry)
			if !errors.Is(err, tree.ErrInvalidTree) {
				t.Fatalf("Insert error = %v, want ErrInvalidTree", err)
			}
		})
	}
}

func TestInsertRejectsConflict(t *testing.T) {
	t.Parallel()

	zero := id.SupportedObjectFormats()[0].Zero()

	for _, tc := range []struct {
		name   string
		first  tree.Entry
		second tree.Entry
	}{
		{
			name:   "same-mode",
			first:  tree.Entry{Mode: mode.Regular, Name: "file", ID: zero},
			second: tree.Entry{Mode: mode.Regular, Name: "file", ID: zero},
		},
		{
			name:   "blob-then-tree",
			first:  tree.Entry{Mode: mode.Regular, Name: "name", ID: zero},
			second: tree.Entry{Mode: mode.Directory, Name: "name", ID: zero},
		},
		{
			name:   "tree-then-blob",
			first:  tree.Entry{Mode: mode.Directory, Name: "name", ID: zero},
			second: tree.Entry{Mode: mode.Regular, Name: "name", ID: zero},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr tree.Tree

			err := tr.Insert(tc.first)
			if err != nil {
				t.Fatalf("Insert(first): %v", err)
			}

			err = tr.Insert(tc.second)
			if !errors.Is(err, tree.ErrInvalidTree) {
				t.Fatalf("Insert(second) error = %v, want ErrInvalidTree", err)
			}
		})
	}
}

func TestEntriesIsCopy(t *testing.T) {
	t.Parallel()

	zero := id.SupportedObjectFormats()[0].Zero()

	var tr tree.Tree

	err := tr.Insert(tree.Entry{Mode: mode.Regular, Name: "file", ID: zero})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	entries := tr.Entries()
	entries[0].Name = "mutated"

	again := tr.Entries()
	if again[0].Name != "file" {
		t.Fatalf("Entries()[0].Name = %q, want %q", again[0].Name, "file")
	}
}
