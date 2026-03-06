package iolimit_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/iolimit"
)

func TestCappedCaptureWriterWithinLimit(t *testing.T) {
	t.Parallel()

	writer := iolimit.NewCappedCaptureWriter(8)

	_, _ = writer.Write([]byte("hello"))
	_, _ = writer.Write([]byte("!"))

	if got := writer.Bytes(); !bytes.Equal(got, []byte("hello!")) {
		t.Fatalf("Bytes() = %q, want %q", got, "hello!")
	}
}

func TestCappedCaptureWriterExceededLimit(t *testing.T) {
	t.Parallel()

	writer := iolimit.NewCappedCaptureWriter(4)

	_, _ = writer.Write([]byte("abcd"))
	_, _ = writer.Write([]byte("x"))

	if got := writer.Bytes(); got != nil {
		t.Fatalf("Bytes() = %q, want nil after overflow", got)
	}
}

func TestCappedCaptureWriterZeroLimit(t *testing.T) {
	t.Parallel()

	writer := iolimit.NewCappedCaptureWriter(0)

	_, _ = writer.Write([]byte("x"))
	if got := writer.Bytes(); got != nil {
		t.Fatalf("Bytes() = %q, want nil at zero limit", got)
	}
}
