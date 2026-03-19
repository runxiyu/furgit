// Package resolve resolves stored Git objects by exact type, by peeling
// tree-ish or commit-ish references, and by path within trees.
//
// A Resolver does not take ownership of the underlying object store.
package resolve
