// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package zlib implements reading and writing of zlib format compressed data,
as specified in RFC 1950.

This package differs from the standard library's compress/zlib package
in that it pools readers and writers to reduce allocations.

Note that closing a reader or writer causes it to be returned to a pool
for reuse. Therefore, the caller must not retain references to a
reader or writer after closing it; in the standard library's
compress/zlib package, it is legal to Reset a closed reader or writer
and continue using it; that is not allowed here, so there is simply no
Resetter interface.

The implementation provides filters that uncompress during reading
and compress during writing.  For example, to write compressed data
to a buffer:

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte("hello, world\n"))
	w.Close()

and to read that data back:

	r, err := zlib.NewReader(&b)
	io.Copy(os.Stdout, r)
	r.Close()
*/
package zlib

import (
	"encoding/binary"
	"errors"
	"hash"
	"io"
	"sync"

	"github.com/klauspost/compress/flate"
)

const (
	zlibDeflate   = 8
	zlibMaxWindow = 7
)

var (
	// ErrChecksum is returned when reading ZLIB data that has an invalid checksum.
	ErrChecksum = errors.New("zlib: invalid checksum")
	// ErrDictionary is returned when reading ZLIB data that has an invalid dictionary.
	ErrDictionary = errors.New("zlib: invalid dictionary")
	// ErrHeader is returned when reading ZLIB data that has an invalid header.
	ErrHeader = errors.New("zlib: invalid header")
)

var readerPool = sync.Pool{
	New: func() any {
		r := new(Reader)

		return r
	},
}

// Reader reads and verifies one zlib stream.
//
// Reader implements io.ReadCloser.
type Reader struct {
	r            flate.Reader
	decompressor io.ReadCloser
	digest       hash.Hash32
	counter      *countingFlateReader
	err          error
	scratch      [4]byte
}

// countingFlateReader wraps flate input and tracks consumed bytes.
type countingFlateReader struct {
	inner flate.Reader
	read  uint64
}

// Read implements io.Reader.
func (reader *countingFlateReader) Read(dst []byte) (int, error) {
	n, err := reader.inner.Read(dst)
	reader.read += uint64(n)

	return n, err
}

// ReadByte implements io.ByteReader.
func (reader *countingFlateReader) ReadByte() (byte, error) {
	b, err := reader.inner.ReadByte()
	if err == nil {
		reader.read++
	}

	return b, err
}

// NewReader creates a new ReadCloser.
// Reads from the returned ReadCloser read and decompress data from r.
// If r does not implement [io.ByteReader], the decompressor may read more
// data than necessary from r.
// It is the caller's responsibility to call Close on the ReadCloser when done.
func NewReader(r io.Reader) (*Reader, error) {
	return NewReaderDict(r, nil)
}

// NewReaderDict is like [NewReader] but uses a preset dictionary.
// NewReaderDict ignores the dictionary if the compressed data does not refer to it.
// If the compressed data refers to a different dictionary, NewReaderDict returns [ErrDictionary].
func NewReaderDict(r io.Reader, dict []byte) (*Reader, error) {
	v := readerPool.Get()

	z, ok := v.(*Reader)
	if !ok {
		panic("zlib: pool returned unexpected type")
	}

	err := z.Reset(r, dict)
	if err != nil {
		return nil, err
	}

	return z, nil
}

// Read decompresses bytes from receiver into p.
func (z *Reader) Read(p []byte) (int, error) {
	if z.err != nil {
		return 0, z.err
	}

	var n int

	n, z.err = z.decompressor.Read(p)

	_, err := z.digest.Write(p[0:n])
	if err != nil {
		z.err = err

		return n, z.err
	}

	if !errors.Is(z.err, io.EOF) {
		// In the normal case we return here.
		return n, z.err
	}

	// Finished file; check checksum.
	_, err = io.ReadFull(z.r, z.scratch[0:4])
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}

		z.err = err

		return n, z.err
	}
	// ZLIB (RFC 1950) is big-endian, unlike GZIP (RFC 1952).
	checksum := binary.BigEndian.Uint32(z.scratch[:4])
	if checksum != z.digest.Sum32() {
		z.err = ErrChecksum

		return n, z.err
	}

	return n, io.EOF
}

// InputConsumed returns compressed bytes consumed from stream input.
//
// This count includes the zlib header, deflate payload, and zlib checksum
// trailer bytes read by the reader.
func (z *Reader) InputConsumed() uint64 {
	if z.counter == nil {
		return 0
	}

	return z.counter.read
}

// Close does not close the wrapped [io.Reader] originally passed to [NewReader].
// In order for the ZLIB checksum to be verified, the reader must be
// fully consumed until the [io.EOF].
func (z *Reader) Close() error {
	if z.err != nil && !errors.Is(z.err, io.EOF) {
		return z.err
	}

	z.err = z.decompressor.Close()
	if z.err != nil {
		return z.err
	}

	readerPool.Put(z)

	return nil
}
