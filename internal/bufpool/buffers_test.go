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
