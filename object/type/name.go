package objecttype

const (
	typeNameBlob   = "blob"
	typeNameTree   = "tree"
	typeNameCommit = "commit"
	typeNameTag    = "tag"
)

// Parse parses a canonical Git object type name.
func Parse(name string) (Type, bool) {
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
func (ty Type) Name() (string, bool) {
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
