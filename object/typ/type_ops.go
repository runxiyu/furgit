package typ

// IsBaseObject reports whether ty is
// one of the four ordinary Git object types.
//
// TODO: This should probably be removed.
func (ty Type) IsBaseObject() bool {
	return ty.details().isBaseObject
}

// Name returns the canonical Git object type name.
func (ty Type) Name() (string, bool) {
	details := ty.details()
	if details.name == "" {
		return "", false
	}

	return details.name, true
}
