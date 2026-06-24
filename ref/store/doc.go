// Package store provides interfaces for reference storage backends.
//
// Reference stores work directly with reference values,
// [ref.Direct] and [ref.Symbolic].
// Unlike object storage,
// they have no separate fetch layer
// to parse backend results into higher-level forms.
//
// The package separates read-only access
// from atomic transactions and non-atomic batches.
// Not every readable reference backend is writable,
// and not every writable backend offers the same update model.
//
// Concrete implementations generally inherit the contract
// documented by the interfaces they satisfy.
// Implementation docs focus on additional guarantees
// and implementation-specific behavior.
package store
