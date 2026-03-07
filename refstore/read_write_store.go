package refstore

// ReadWriteStore supports reading, atomic transactions, and non-atomic batches.
type ReadWriteStore interface {
	ReadingStore
	TransactionalStore
	BatchStore
}
