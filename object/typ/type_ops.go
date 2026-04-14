package typ

// Name returns the canonical Git object type name.
func (ty Type) Name() (string, bool) {
	details := ty.details()
	if details.name == "" {
		return "", false
	}

	return details.name, true
}
