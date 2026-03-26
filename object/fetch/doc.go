// Package fetch resolves stored Git objects by exact type, by peeling
// tree-ish or commit-ish references, and by path within trees.
//
// A Fetcher does not take ownership of the underlying object store.
package fetch
