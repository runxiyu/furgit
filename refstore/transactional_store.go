package refstore

// TransactionalStore begins atomic reference transactions.
//
// Not every readable reference store is writable. Implementations should only
// satisfy TransactionalStore when they can stage and commit reference updates
// atomically within that backend.
type TransactionalStore interface {
	// BeginTransaction creates one new mutable transaction.
	BeginTransaction() (Transaction, error)
}
