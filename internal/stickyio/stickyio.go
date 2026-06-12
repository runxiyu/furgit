package stickyio

import (
	"encoding/binary"
	"io"
)

// Writer forwards writes to an underlying writer,
// retains the first write error,
// and turns subsequent writes into no-ops.
//
// Call sites that cannot act on individual write errors
// should use [Writer.Put] and its sibling methods
// and check [Writer.Err] once after emission;
// the retained error is deferred there, not discarded.
//
// Labels: MT-Unsafe.
type Writer struct {
	w   io.Writer
	err error
}

var _ io.Writer = (*Writer)(nil)

// New creates a sticky writer over w.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(w io.Writer) *Writer {
	return &Writer{
		w:   w,
		err: nil,
	}
}

// Write implements [io.Writer],
// forwarding p to the underlying writer
// and reporting its result truthfully.
// After a write error,
// Write does nothing and returns the retained error.
func (writer *Writer) Write(p []byte) (int, error) {
	if writer.err != nil {
		return 0, writer.err
	}

	n, err := writer.w.Write(p)
	if err != nil {
		writer.err = err
	}

	return n, err
}

// Put writes p without per-call error reporting;
// failures are retained for [Writer.Err].
func (writer *Writer) Put(p []byte) {
	_, _ = writer.Write(p)
}

// PutUint32 writes one big-endian uint32
// without per-call error reporting.
func (writer *Writer) PutUint32(v uint32) {
	var buf [4]byte

	binary.BigEndian.PutUint32(buf[:], v)
	writer.Put(buf[:])
}

// PutUint64 writes one big-endian uint64
// without per-call error reporting.
func (writer *Writer) PutUint64(v uint64) {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[:], v)
	writer.Put(buf[:])
}

// Err returns the first write error, if any.
func (writer *Writer) Err() error {
	return writer.err
}
