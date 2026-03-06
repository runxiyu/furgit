package bufpool

// Borrow retrieves a Buffer suitable for storing up to capHint bytes.
// The returned Buffer may come from an internal sync.Pool.
//
// If capHint is smaller than DefaultBufferCap, it is automatically raised
// to DefaultBufferCap. If no pooled buffer has sufficient capacity, a new
// unpooled buffer is allocated.
//
// The caller must call Release() when finished using the returned Buffer.
func Borrow(capHint int) Buffer {
	if capHint < DefaultBufferCap {
		capHint = DefaultBufferCap
	}

	classIdx, classCap, pooled := classFor(capHint)
	if !pooled {
		newBuf := make([]byte, 0, capHint)

		return Buffer{buf: newBuf, pool: unpooled}
	}
	//nolint:forcetypeassert
	buf := bufferPools[classIdx].Get().(*[]byte)
	if cap(*buf) < classCap {
		*buf = make([]byte, 0, classCap)
	}

	slice := (*buf)[:0]

	return Buffer{buf: slice, pool: poolIndex(classIdx)} //#nosec G115
}
