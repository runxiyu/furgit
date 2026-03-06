package bufpool

// Buffer is a growable byte container that optionally participates in a
// memory pool. A Buffer may be obtained through Borrow() or constructed
// directly from owned data via FromOwned().
//
// A Buffer's underlying slice may grow as needed. When finished with a
// pooled buffer, the caller should invoke Release() to return it to the pool.
//
// Buffers must not be copied after first use; doing so can cause double-returns
// to the pool and data races.
//
// In general, pass Buffer around when used internally, and directly .Bytes() when
// returning output across our API boundary. It is neither necessary nor efficient
// to copy/append the .Bytes() to a newly-allocated slice; in cases where we do
// want the raw byte slice out of our API boundary, it is perfectly acceptable to
// simply not call Release().
//
//go:nocopy
type Buffer struct {
	_    struct{} // for nocopy
	buf  []byte
	pool poolIndex
}
