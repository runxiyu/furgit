package object

import objecttype "codeberg.org/lindenii/furgit/object/type"

// TypeFor returns the Git object type for T when T is one of the standard
// parsed object types.
func TypeFor[T Object]() (objecttype.Type, bool) {
	switch any(*new(T)).(type) {
	case *Blob:
		return objecttype.TypeBlob, true
	case *Tree:
		return objecttype.TypeTree, true
	case *Commit:
		return objecttype.TypeCommit, true
	case *Tag:
		return objecttype.TypeTag, true
	default:
		return 0, false
	}
}
