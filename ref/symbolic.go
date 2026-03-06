package ref

// Symbolic points to another reference name.
type Symbolic struct {
	RefName string
	Target  string
}

// Name returns the fully-qualified reference name.
func (ref Symbolic) Name() string {
	return ref.RefName
}

func (Symbolic) isRef() {}
