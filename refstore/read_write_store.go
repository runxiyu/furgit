package refstore

// ReadWriteStore supports reading, atomic transactions, and immediate batches.
type ReadWriteStore interface {
	ReadingStore
	TransactionalStore
	BatchStore
}
