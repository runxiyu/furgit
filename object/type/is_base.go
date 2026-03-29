package objecttype

// IsBaseObject reports whether ty is one of the four canonical Git object
// types encoded directly in pack entries.
func (ty Type) IsBaseObject() bool {
	return ty.details().isBaseObject
}
