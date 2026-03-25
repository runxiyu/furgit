// Package tree provides representations, parsers, and serializers for tree objects.
package tree

// Tree represents a Git tree object.
type Tree struct {
	// Entries must be sorted by TreeEntryNameCompare.
	// You are strongly advised to use the methods for manipulation
	// rather than modifying the slice yourself.
	Entries []TreeEntry
}
