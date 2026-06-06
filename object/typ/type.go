package typ

// Type represents a Git object type.
type Type uint8

const (
	// TypeUnknown represents an unknown or unset Git object type.
	TypeUnknown Type = iota

	// TypeBlob represents a Git blob.
	TypeBlob

	// TypeTree represents a Git tree.
	TypeTree

	// TypeCommit represents a Git commit.
	TypeCommit

	// TypeTag represents a Git tag.
	TypeTag
)
