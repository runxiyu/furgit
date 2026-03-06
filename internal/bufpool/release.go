package bufpool

// Release returns the buffer to the global pool if it originated from the
// pool and its capacity is no larger than maxPooledBuffer. After release, the
// Buffer becomes invalid and should not be used further.
//
// Releasing a non-pooled buffer has no effect beyond clearing its internal
// storage.
func (buf *Buffer) Release() {
	if buf.buf == nil {
		return
	}

	buf.returnToPool()
	buf.buf = nil
	buf.pool = unpooled
}
