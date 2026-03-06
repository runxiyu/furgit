package bufpool

const (
	// DefaultBufferCap is the minimum capacity a borrowed buffer will have.
	// Borrow() will allocate or retrieve a buffer with at least this capacity.
	DefaultBufferCap = 32 * 1024

	// maxPooledBuffer defines the maximum capacity of a buffer that may be
	// returned to the pool. Buffers larger than this will not be pooled to
	// avoid unbounded memory usage.
	maxPooledBuffer = 8 << 20
)
