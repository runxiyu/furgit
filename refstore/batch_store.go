package refstore

// BatchStore begins non-atomic reference batches.
type BatchStore interface {
	// BeginBatch creates one new queued batch.
	BeginBatch() (Batch, error)
}
