package typ

// Name returns the canonical Git object type name.
func (ty Type) Name() (string) {
	return ty.details().name
}
