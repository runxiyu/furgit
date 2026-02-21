// Package iolimit provides small internal reader wrappers for length-constrained
// stream I/O.
package iolimit

import (
	"errors"
	"io"
)

// ErrExpectedLengthExceeded reports that a stream produced bytes beyond the
// expected length.
var ErrExpectedLengthExceeded = errors.New("iolimit: stream exceeded expected length")

// ExpectLengthReader wraps src and enforces an expected byte length.
//
// It returns io.ErrUnexpectedEOF if src ends before expected bytes are read.
// It returns ErrExpectedLengthExceeded if reads continue beyond the expected
// boundary and src still produces bytes.
//
// This reader does not drain src on close or at the expected boundary. As a
// result, overlength streams are detected only when a caller reads at or past
// the boundary.
func ExpectLengthReader(src io.Reader, expected int64) io.Reader {
	return &expectLengthReader{
		src:       src,
		remaining: expected,
	}
}

type expectLengthReader struct {
	src       io.Reader
	remaining int64
}

func (reader *expectLengthReader) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}

	if reader.remaining == 0 {
		var probe [1]byte
		n, err := reader.src.Read(probe[:])
		if n > 0 {
			return 0, ErrExpectedLengthExceeded
		}
		if err == nil {
			return 0, nil
		}
		return 0, err
	}

	if reader.remaining < 0 {
		return 0, ErrExpectedLengthExceeded
	}

	if int64(len(dst)) > reader.remaining {
		dst = dst[:reader.remaining]
	}

	n, err := reader.src.Read(dst)
	if n > 0 {
		reader.remaining -= int64(n)
	}

	if err == io.EOF {
		if reader.remaining > 0 {
			return n, io.ErrUnexpectedEOF
		}
		if n > 0 {
			return n, nil
		}
		return 0, io.EOF
	}

	return n, err
}
