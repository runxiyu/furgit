// Package objecttype provides object type constants and names.
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

const (
	typeNameBlob   = "blob"
	typeNameTree   = "tree"
	typeNameCommit = "commit"
	typeNameTag    = "tag"
)

// ParseName parses a canonical Git object type name.
func ParseName(name string) (Type, bool) {
	switch name {
	case typeNameBlob:
		return TypeBlob, true
	case typeNameTree:
		return TypeTree, true
	case typeNameCommit:
		return TypeCommit, true
	case typeNameTag:
		return TypeTag, true
	default:
		return TypeInvalid, false
	}
}

// Name returns the canonical Git object type name.
func Name(ty Type) (string, bool) {
	switch ty {
	case TypeBlob:
		return typeNameBlob, true
	case TypeTree:
		return typeNameTree, true
	case TypeCommit:
		return typeNameCommit, true
	case TypeTag:
		return typeNameTag, true
	case TypeInvalid, TypeFuture, TypeOfsDelta, TypeRefDelta:
		return "", false
	default:
		return "", false
	}
}
