package bufpool

// Resize adjusts the length of the buffer to n bytes. If n exceeds the current
// capacity, the underlying storage is grown. If n is negative, it is treated
// as zero.
//
// The buffer's new contents beyond the previous length are undefined.
func (buf *Buffer) Resize(n int) {
	if n < 0 {
		n = 0
	}

	buf.ensureCapacity(n)
	buf.buf = buf.buf[:n]
}
