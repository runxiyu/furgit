package typ

// Parse parses a canonical Git object type name.
func Parse(name string) (Type, error) {
	ty, ok := typeByName[name]

	if !ok {
		return 0, ErrInvalidType
	}

	return ty, nil
}
