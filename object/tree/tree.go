// Package tree provides representations, parsers, and serializers for tree objects.
package tree

// Tree represents a fully materialized Git tree object.
//
// Labels: MT-Unsafe.
type Tree struct {
	// Entries must be sorted by TreeEntryNameCompare.
	// Use the Tree methods to preserve ordering and copy semantics rather than
	// modifying the slice directly.
	Entries []TreeEntry
}
