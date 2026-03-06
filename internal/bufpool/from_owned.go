package bufpool

// FromOwned constructs a Buffer from a caller-owned byte slice. The resulting
// Buffer does not participate in pooling and will never be returned to the
// internal pool when released.
func FromOwned(buf []byte) Buffer {
	return Buffer{buf: buf, pool: unpooled}
}
