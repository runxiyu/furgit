package typ

// Type represents a Git object type.
//
// The values currently mirror what's found in the Git packfile format.
//
// TODO: Revisit this.
type Type uint8

const (
	// TypeInvalid represents an invalid Git object type.
	TypeInvalid Type = 0

	// TypeCommit represents a Git commit.
	TypeCommit Type = 1

	// TypeTree represents a Git tree.
	TypeTree Type = 2

	// TypeBlob represents a Git blob.
	TypeBlob Type = 3

	// TypeTag represents a Git tag.
	TypeTag Type = 4

	// TypeFuture is reserved for the future, just like in the packfile format.
	TypeFuture Type = 5

	// TypeOfsDelta is reserved for internal use in packfile handlers.
	//
	// TODO: Revisit this.
	TypeOfsDelta Type = 6

	// TypeRefDelta is reserved for internal use in packfile handlers.
	//
	// TODO: Revisit this.
	TypeRefDelta Type = 7
)
