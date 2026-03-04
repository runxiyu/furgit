package iolimit_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"codeberg.org/lindenii/furgit/internal/iolimit"
)

func TestExpectLengthReaderExact(t *testing.T) {
	t.Parallel()

	r := iolimit.ExpectLengthReader(bytes.NewReader([]byte("hello")), 5)

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("ReadAll = %q, want %q", got, "hello")
	}

	buf := make([]byte, 1)

	n, err := r.Read(buf)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("post-boundary Read = (%d,%v), want (0,EOF)", n, err)
	}
}

func TestExpectLengthReaderShort(t *testing.T) {
	t.Parallel()

	r := iolimit.ExpectLengthReader(bytes.NewReader([]byte("hey")), 5)

	_, err := io.ReadAll(r)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("ReadAll error = %v, want ErrUnexpectedEOF", err)
	}
}

func TestExpectLengthReaderLongDetectedOnNextRead(t *testing.T) {
	t.Parallel()

	r := iolimit.ExpectLengthReader(bytes.NewReader([]byte("hello!")), 5)
	buf := make([]byte, 5)

	n, err := io.ReadFull(r, buf)
	if err != nil {
		t.Fatalf("ReadFull error: %v", err)
	}

	if n != 5 || !bytes.Equal(buf, []byte("hello")) {
		t.Fatalf("ReadFull = (%d,%q), want (5,hello)", n, buf)
	}

	probe := make([]byte, 1)

	n, err = r.Read(probe)
	if n != 0 || !errors.Is(err, iolimit.ErrExpectedLengthExceeded) {
		t.Fatalf("overflow Read = (%d,%v), want (0,ErrExpectedLengthExceeded)", n, err)
	}
}

func TestExpectLengthReaderEmptyExpected(t *testing.T) {
	t.Parallel()

	r := iolimit.ExpectLengthReader(bytes.NewReader(nil), 0)
	buf := make([]byte, 1)

	n, err := r.Read(buf)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("Read = (%d,%v), want (0,EOF)", n, err)
	}
}
