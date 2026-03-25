// Package tree provides representations, parsers, and serializers for tree objects.
package tree

// Tree represents a Git tree object.
type Tree struct {
	Entries []TreeEntry
}
