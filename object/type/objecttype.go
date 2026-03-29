// Package objecttype provides Git object type tags and names.
package objecttype

// Type mirrors Git object type tags in packfiles.
type Type uint8

const (
	TypeInvalid  Type = 0
	TypeCommit   Type = 1
	TypeTree     Type = 2
	TypeBlob     Type = 3
	TypeTag      Type = 4
	TypeFuture   Type = 5
	TypeOfsDelta Type = 6
	TypeRefDelta Type = 7
)
