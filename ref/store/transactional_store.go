package store

// Transactioner begins atomic reference transactions.
//
// Implementations should only satisfy Transactioner
// when they can stage and commit reference updates
// atomically within that backend.
type Transactioner interface {
	// BeginTransaction creates one new mutable transaction.
	//
	// Labels: Deps-Borrowed, Life-Parent.
	BeginTransaction() (Transaction, error)
}
