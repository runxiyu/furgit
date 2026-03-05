// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib

import (
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"sync"

	"codeberg.org/lindenii/furgit/internal/compress/flate"
)

// These constants are copied from the [flate] package, so that code that imports
// [compress/zlib] does not also have to import [compress/flate].
const (
	NoCompression      = flate.NoCompression
	BestSpeed          = flate.BestSpeed
	BestCompression    = flate.BestCompression
	DefaultCompression = flate.DefaultCompression
	HuffmanOnly        = flate.HuffmanOnly
)

// A Writer takes data written to it and writes the compressed
// form of that data to an underlying writer (see [NewWriter]).
type Writer struct {
	w           io.Writer
	level       int
	dict        []byte
	compressor  *flate.Writer
	digest      hash.Hash32
	err         error
	scratch     [4]byte
	wroteHeader bool
}

var writerPool = sync.Pool{
	New: func() any {
		return new(Writer)
	},
}

// NewWriter creates a new [Writer].
// Writes to the returned Writer are compressed and written to w.
//
// It is the caller's responsibility to call Close on the Writer when done.
// Writes may be buffered and not flushed until Close.
func NewWriter(w io.Writer) *Writer {
	z, _ := NewWriterLevelDict(w, DefaultCompression, nil)

	return z
}

// NewWriterLevel is like [NewWriter] but specifies the compression level instead
// of assuming [DefaultCompression].
//
// The compression level can be [DefaultCompression], [NoCompression], [HuffmanOnly]
// or any integer value between [BestSpeed] and [BestCompression] inclusive.
// The error returned will be nil if the level is valid.
func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	return NewWriterLevelDict(w, level, nil)
}

// NewWriterLevelDict is like [NewWriterLevel] but specifies a dictionary to
// compress with.
//
// The dictionary may be nil. If not, its contents should not be modified until
// the Writer is closed.
func NewWriterLevelDict(w io.Writer, level int, dict []byte) (*Writer, error) {
	if level < HuffmanOnly || level > BestCompression {
		return nil, fmt.Errorf("zlib: invalid compression level: %d", level)
	}

	v := writerPool.Get()

	z, ok := v.(*Writer)
	if !ok {
		panic("zlib: pool returned unexpected type")
	}

	// flate.Writer can only be Reset with the same level/dictionary mode.
	// Reuse it only when the configuration is unchanged and dictionary-free.
	reuseCompressor := z.compressor != nil && z.level == level && z.dict == nil && dict == nil
	if !reuseCompressor {
		z.compressor = nil
	}

	if z.digest != nil {
		z.digest.Reset()
	}

	*z = Writer{
		w:          w,
		level:      level,
		dict:       dict,
		compressor: z.compressor,
		digest:     z.digest,
	}
	if z.compressor != nil {
		z.compressor.Reset(w)
	}

	return z, nil
}

// Reset clears the state of the [Writer] z such that it is equivalent to its
// initial state from [NewWriterLevel] or [NewWriterLevelDict], but instead writing
// to w.
func (z *Writer) Reset(w io.Writer) {
	z.w = w
	// z.level and z.dict left unchanged.
	if z.compressor != nil {
		z.compressor.Reset(w)
	}

	if z.digest != nil {
		z.digest.Reset()
	}

	z.err = nil
	z.scratch = [4]byte{}
	z.wroteHeader = false
}

// Write writes a compressed form of p to the underlying [io.Writer]. The
// compressed bytes are not necessarily flushed until the [Writer] is closed or
// explicitly flushed.
func (z *Writer) Write(p []byte) (n int, err error) {
	if !z.wroteHeader {
		z.err = z.writeHeader()
	}

	if z.err != nil {
		return 0, z.err
	}

	if len(p) == 0 {
		return 0, nil
	}

	n, err = z.compressor.Write(p)
	if err != nil {
		z.err = err

		return n, err
	}

	_, err = z.digest.Write(p)
	if err != nil {
		z.err = err

		return 0, z.err
	}

	return n, err
}

// Flush flushes the Writer to its underlying [io.Writer].
func (z *Writer) Flush() error {
	if !z.wroteHeader {
		z.err = z.writeHeader()
	}

	if z.err != nil {
		return z.err
	}

	z.err = z.compressor.Flush()

	return z.err
}

// Close closes the Writer, flushing any unwritten data to the underlying
// [io.Writer], but does not close the underlying io.Writer.
func (z *Writer) Close() error {
	if !z.wroteHeader {
		z.err = z.writeHeader()
	}

	if z.err != nil {
		return z.err
	}

	z.err = z.compressor.Close()
	if z.err != nil {
		return z.err
	}

	checksum := z.digest.Sum32()
	// ZLIB (RFC 1950) is big-endian, unlike GZIP (RFC 1952).
	binary.BigEndian.PutUint32(z.scratch[:], checksum)

	_, z.err = z.w.Write(z.scratch[0:4])
	if z.err != nil {
		return z.err
	}

	writerPool.Put(z)

	return nil
}
