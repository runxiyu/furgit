package bufpool

// Bytes returns the underlying byte slice that represents the current contents
// of the buffer. Modifying the returned slice modifies the Buffer itself.
func (buf *Buffer) Bytes() []byte {
	return buf.buf
}
