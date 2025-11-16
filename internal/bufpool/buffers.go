// Package bufpool provides a lightweight byte-buffer type with optional
// pooling.
package bufpool

import "sync"

const (
	// DefaultBufferCap is the minimum capacity a borrowed buffer will have.
	// Borrow() will allocate or retrieve a buffer with at least this capacity.
	DefaultBufferCap = 32 * 1024

	// maxPooledBuffer defines the maximum capacity of a buffer that may be
	// returned to the pool. Buffers larger than this will not be pooled to
	// avoid unbounded memory usage.
	maxPooledBuffer = 8 << 20
)

// Buffer is a growable byte container that optionally participates in a
// memory pool. A Buffer may be obtained through Borrow() or constructed
// directly from owned data via FromOwned().
//
// A Buffer's underlying slice may grow as needed. When finished with a
// pooled buffer, the caller should invoke Release() to return it to the pool.
//
// A zero-value Buffer is not valid for use.
type Buffer struct {
	buf    []byte
	pooled bool
}

var bufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, DefaultBufferCap)
		return &buf
	},
}

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
	buf := bufPool.Get().(*[]byte)
	if cap(*buf) < capHint {
		bufPool.Put(buf)
		newBuf := make([]byte, 0, capHint)
		return Buffer{buf: newBuf, pooled: false}
	}
	slice := (*buf)[:0]
	return Buffer{buf: slice, pooled: true}
}

// FromOwned constructs a Buffer from a caller-owned byte slice. The resulting
// Buffer does not participate in pooling and will never be returned to the
// internal pool when released.
func FromOwned(buf []byte) Buffer {
	return Buffer{buf: buf, pooled: false}
}

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

// Bytes returns the underlying byte slice that represents the current contents
// of the buffer. Modifying the returned slice modifies the Buffer itself.
func (buf *Buffer) Bytes() []byte {
	return buf.buf
}

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
	if buf.pooled && cap(buf.buf) <= maxPooledBuffer {
		tmp := buf.buf[:0]
		bufPool.Put(&tmp)
	}
	buf.buf = nil
	buf.pooled = false
}

// ensureCapacity grows the underlying buffer to accommodate the requested
// number of bytes. Growth doubles the capacity by default unless a larger
// expansion is needed. If the previous storage was pooled and not oversized,
// it is returned to the pool.
func (buf *Buffer) ensureCapacity(needed int) {
	if cap(buf.buf) >= needed {
		return
	}
	old := buf.buf
	wasPooled := buf.pooled
	newCap := cap(buf.buf) * 2
	if newCap < needed {
		newCap = needed
	}
	newBuf := make([]byte, len(buf.buf), newCap)
	copy(newBuf, buf.buf)
	buf.buf = newBuf
	buf.pooled = false
	if wasPooled && cap(old) <= maxPooledBuffer {
		tmp := old[:0]
		bufPool.Put(&tmp)
	}
}
