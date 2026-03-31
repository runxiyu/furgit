package refstore

// Batcher begins non-atomic reference batches.
type Batcher interface {
	// BeginBatch creates one new queued batch.
	//
	// Labels: Life-Parent.
	BeginBatch() (Batch, error)
}
