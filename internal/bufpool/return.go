package bufpool

func (buf *Buffer) returnToPool() {
	if buf.pool == unpooled {
		return
	}

	tmp := buf.buf[:0]
	bufferPools[int(buf.pool)].Put(&tmp)
}
