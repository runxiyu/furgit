package ref

// Ref is a Git reference.
//
// Consider casting to [Detached] or [Symbolic].
type Ref interface {
	isRef()
	Name() string
}
