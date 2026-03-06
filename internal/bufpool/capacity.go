package bufpool

// ensureCapacity grows the underlying buffer to accommodate the requested
// number of bytes. Growth doubles the capacity by default unless a larger
// expansion is needed. If the previous storage was pooled and not oversized,
// it is returned to the pool.
func (buf *Buffer) ensureCapacity(needed int) {
	if cap(buf.buf) >= needed {
		return
	}

	classIdx, classCap, pooled := classFor(needed)

	var newBuf []byte

	if pooled {
		//nolint:forcetypeassert
		raw := bufferPools[classIdx].Get().(*[]byte)
		if cap(*raw) < classCap {
			*raw = make([]byte, 0, classCap)
		}

		newBuf = (*raw)[:len(buf.buf)]
	} else {
		newBuf = make([]byte, len(buf.buf), classCap)
	}

	copy(newBuf, buf.buf)
	buf.returnToPool()

	buf.buf = newBuf
	if pooled {
		buf.pool = poolIndex(classIdx) //#nosec G115
	} else {
		buf.pool = unpooled
	}
}
