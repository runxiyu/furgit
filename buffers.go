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

func (body *borrowedBody) Resize(n int) {
	if n < 0 {
		n = 0
	}
	body.ensureCapacity(n)
	body.buf = body.buf[:n]
}

func (body *borrowedBody) Append(src []byte) {
	if len(src) == 0 {
		return
	}
	start := len(body.buf)
	body.ensureCapacity(start + len(src))
	body.buf = body.buf[:start+len(src)]
	copy(body.buf[start:], src)
}

func (body *borrowedBody) Bytes() []byte {
	return body.buf
}

func (body *borrowedBody) Release() {
	if body.buf == nil {
		return
	}
	if body.pooled && cap(body.buf) <= maxPooledBody {
		tmp := body.buf[:0]
		bodyPool.Put(&tmp)
	}
	body.buf = nil
	body.pooled = false
}

func (body *borrowedBody) ensureCapacity(needed int) {
	if cap(body.buf) >= needed {
		return
	}
	old := body.buf
	wasPooled := body.pooled
	newCap := cap(body.buf) * 2
	if newCap < needed {
		newCap = needed
	}
	newBuf := make([]byte, len(body.buf), newCap)
	copy(newBuf, body.buf)
	body.buf = newBuf
	body.pooled = false
	if wasPooled && cap(old) <= maxPooledBody {
		tmp := old[:0]
		bodyPool.Put(&tmp)
	}
}
