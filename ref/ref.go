// Package ref provides general, detached, and symbolic references.
package ref

// Ref is a Git reference.
//
// Implementations must be in this package.
type Ref interface {
	isRef()
	Name() string
}
