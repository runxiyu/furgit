package refstore

// Transactioner begins atomic reference transactions.
//
// Not every readable reference store is writable. Implementations should only
// satisfy Transactioner when they can stage and commit reference updates
// atomically within that backend.
type Transactioner interface {
	// BeginTransaction creates one new mutable transaction.
	//
	// Labels: Life-Parent.
	BeginTransaction() (Transaction, error)
}
