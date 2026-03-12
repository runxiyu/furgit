package packfile

import "codeberg.org/lindenii/furgit/objecttype"

// IsBaseObjectType reports whether ty is one of the four canonical object
// types encoded directly in pack entries.
func IsBaseObjectType(ty objecttype.Type) bool {
	switch ty {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		return true
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return false
	default:
		return false
	}
}
