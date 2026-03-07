package refstore

// ReadWriteStore supports both reading and atomic transactional updates.
type ReadWriteStore interface {
	ReadingStore
	TransactionalStore
}
