// Package refstore provides interfaces for reference storage backends.
//
// Ref stores directly use reference values. Unlike object storage, they
// do not have a separate fetch layer to parse backend results into
// higher-level forms.
//
// The package separates read-only access from atomic transactions and
// non-atomic batches. Not every readable ref backend is writable, and not
// every writable backend necessarily offers the same update model.
//
// Concrete implementations generally inherit the contract documented by the
// interfaces they satisfy. Implementation docs focus on additional guarantees
// and implementation-specific behavior.
package refstore
