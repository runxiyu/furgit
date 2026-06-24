package store

// Batcher begins non-atomic reference batches.
type Batcher interface {
	// BeginBatch creates one new queued batch.
	//
	// Labels: Deps-Borrowed, Life-Parent.
	BeginBatch() (Batch, error)
}
