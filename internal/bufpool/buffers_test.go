package bufpool

import "testing"

func TestBorrowBufferResizeAndAppend(t *testing.T) {
	b := Borrow(1)
	defer b.Release()

	if cap(b.buf) < DefaultBufferCap {
		t.Fatalf("expected capacity >= %d, got %d", DefaultBufferCap, cap(b.buf))
	}

	b.Append([]byte("alpha"))
	b.Append([]byte("beta"))
	if got := string(b.Bytes()); got != "alphabeta" {
		t.Fatalf("unexpected contents: %q", got)
	}

	b.Resize(3)
	if got := string(b.Bytes()); got != "alp" {
		t.Fatalf("resize shrink mismatch: %q", got)
	}

	b.Resize(8)
	if len(b.Bytes()) != 8 {
		t.Fatalf("expected len 8 after grow, got %d", len(b.Bytes()))
	}
	if prefix := string(b.Bytes()[:3]); prefix != "alp" {
		t.Fatalf("prefix lost after grow: %q", prefix)
	}
}

func TestBorrowBufferRelease(t *testing.T) {
	b := Borrow(DefaultBufferCap / 2)
	b.Append([]byte("data"))
	b.Release()
	if b.buf != nil {
		t.Fatal("expected buffer cleared after release")
	}
}

func TestBorrowUsesLargerPools(t *testing.T) {
	const request = DefaultBufferCap * 4

	classIdx, classCap, pooled := classFor(request)
	if !pooled {
		t.Fatalf("expected %d to map to a pooled class", request)
	}

	b := Borrow(request)
	if b.pool != poolIndex(classIdx) {
		t.Fatalf("expected pooled buffer in class %d, got %d", classIdx, b.pool)
	}
	if cap(b.buf) != classCap {
		t.Fatalf("expected capacity %d, got %d", classCap, cap(b.buf))
	}
	b.Release()

	b2 := Borrow(request)
	defer b2.Release()
	if b2.pool != poolIndex(classIdx) {
		t.Fatalf("expected pooled buffer in class %d on reuse, got %d", classIdx, b2.pool)
	}
	if cap(b2.buf) != classCap {
		t.Fatalf("expected capacity %d on reuse, got %d", classCap, cap(b2.buf))
	}
}

func TestGrowingBufferStaysPooled(t *testing.T) {
	b := Borrow(DefaultBufferCap)
	defer b.Release()

	b.Append(make([]byte, DefaultBufferCap*3))
	if b.pool == unpooled {
		t.Fatal("buffer should stay pooled after growth within limit")
	}
}
