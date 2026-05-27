package trees

import "lindenii.org/go/furgit/object/tree"

// Entry is one recursive tree difference at a path.
type Entry struct {
	// Path is the slash-separated path relative to the diff root.
	Path []byte
	// Kind is the difference kind for this path.
	Kind EntryKind
	// Old is the old tree entry (nil when Kind is EntryKindAdded).
	Old *tree.TreeEntry
	// New is the new tree entry (nil when Kind is EntryKindDeleted).
	New *tree.TreeEntry
}
