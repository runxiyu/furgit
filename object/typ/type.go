package typ

// Type represents a Git object type.
type Type uint8

const (
	// Unknown represents an unknown or unset Git object type.
	Unknown Type = iota

	// Blob represents a Git blob.
	Blob

	// Tree represents a Git tree.
	Tree

	// Commit represents a Git commit.
	Commit

	// Tag represents a Git tag.
	Tag
)
