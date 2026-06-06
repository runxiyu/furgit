package ref

// Ref is a Git reference.
//
// Consider casting to [Direct] or [Symbolic].
type Ref interface {
	isRef()
	Name() string
}
