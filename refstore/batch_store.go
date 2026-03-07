package refstore

// BatchStore begins non-atomic reference batches.
type BatchStore interface {
	// BeginBatch creates one new immediate-apply batch.
	BeginBatch() (Batch, error)
}
