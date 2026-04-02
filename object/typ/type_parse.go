package typ

// Parse parses a canonical Git object type name.
func Parse(name string) (Type, bool) {
	ty, ok := typeByName[name]

	return ty, ok
}
