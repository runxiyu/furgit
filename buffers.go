package furgit

import "sync"

const (
	defaultBodyCap = 32 * 1024
	maxPooledBody  = 8 << 20
)

type borrowedBody struct {
	buf    []byte
	pooled bool
}

var bodyPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, defaultBodyCap)
		return &buf
	},
}

func borrowBody(capHint int) borrowedBody {
	if capHint < defaultBodyCap {
		capHint = defaultBodyCap
	}
	buf := bodyPool.Get().(*[]byte)
	if cap(*buf) < capHint {
		bodyPool.Put(buf)
		newBuf := make([]byte, 0, capHint)
		return borrowedBody{buf: newBuf, pooled: false}
	}
	slice := (*buf)[:0]
	return borrowedBody{buf: slice, pooled: true}
}

func borrowedFromOwned(buf []byte) borrowedBody {
	return borrowedBody{buf: buf}
}

func (b *borrowedBody) Resize(n int) {
	if n < 0 {
		n = 0
	}
	b.ensureCapacity(n)
	b.buf = b.buf[:n]
}

func (b *borrowedBody) Append(src []byte) {
	if len(src) == 0 {
		return
	}
	start := len(b.buf)
	b.ensureCapacity(start + len(src))
	b.buf = b.buf[:start+len(src)]
	copy(b.buf[start:], src)
}

func (b *borrowedBody) Bytes() []byte {
	return b.buf
}

func (b *borrowedBody) Release() {
	if b.buf == nil {
		return
	}
	if b.pooled && cap(b.buf) <= maxPooledBody {
		tmp := b.buf[:0]
		bodyPool.Put(&tmp)
	}
	b.buf = nil
	b.pooled = false
}

func (b *borrowedBody) ensureCapacity(needed int) {
	if cap(b.buf) >= needed {
		return
	}
	old := b.buf
	wasPooled := b.pooled
	newCap := cap(b.buf) * 2
	if newCap < needed {
		newCap = needed
	}
	newBuf := make([]byte, len(b.buf), newCap)
	copy(newBuf, b.buf)
	b.buf = newBuf
	b.pooled = false
	if wasPooled && cap(old) <= maxPooledBody {
		tmp := old[:0]
		bodyPool.Put(&tmp)
	}
}
