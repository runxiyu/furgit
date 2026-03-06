package bufpool

// Append copies the provided bytes onto the end of the buffer, growing its
// capacity if required. If src is empty, the method does nothing.
//
// The receiver retains ownership of the data; the caller may reuse src freely.
func (buf *Buffer) Append(src []byte) {
	if len(src) == 0 {
		return
	}

	start := len(buf.buf)
	buf.ensureCapacity(start + len(src))
	buf.buf = buf.buf[:start+len(src)]
	copy(buf.buf[start:], src)
}
