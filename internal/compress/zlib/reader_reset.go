// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"

	"codeberg.org/lindenii/furgit/internal/adler32"
	"codeberg.org/lindenii/furgit/internal/compress/flate"
)

// reset resets receiver to read a new zlib stream.
func (z *Reader) reset(r io.Reader, dict []byte) error {
	*z = Reader{decompressor: z.decompressor}

	var input flate.Reader
	if fr, ok := r.(flate.Reader); ok {
		input = fr
	} else {
		input = bufio.NewReader(r)
	}

	z.r = input

	// Read the header (RFC 1950 section 2.2.).
	readN, err := io.ReadFull(z.r, z.scratch[0:2])
	z.headerRead += uint64(readN)
	z.err = err
	if z.err != nil {
		if errors.Is(z.err, io.EOF) {
			z.err = io.ErrUnexpectedEOF
		}

		return z.err
	}

	h := binary.BigEndian.Uint16(z.scratch[:2])
	if (z.scratch[0]&0x0f != zlibDeflate) || (z.scratch[0]>>4 > zlibMaxWindow) || (h%31 != 0) {
		z.err = ErrHeader

		return z.err
	}

	haveDict := z.scratch[1]&0x20 != 0
	if haveDict {
		readN, z.err = io.ReadFull(z.r, z.scratch[0:4])
		z.headerRead += uint64(readN)
		if z.err != nil {
			if errors.Is(z.err, io.EOF) {
				z.err = io.ErrUnexpectedEOF
			}

			return z.err
		}

		checksum := binary.BigEndian.Uint32(z.scratch[:4])
		if checksum != adler32.Checksum(dict) {
			z.err = ErrDictionary

			return z.err
		}
	}

	if z.decompressor != nil {
		resetter, ok := z.decompressor.(flate.Resetter)
		if !ok {
			panic("zlib: pooled decompressor does not implement flate.Resetter")
		}

		z.err = resetter.Reset(z.r, dict)
		if z.err != nil {
			return z.err
		}
		progress, ok := z.decompressor.(flate.InputProgress)
		if !ok {
			panic("zlib: pooled decompressor does not implement flate.InputProgress")
		}
		z.progress = progress

		z.digest = adler32.New()

		return nil
	}

	if haveDict {
		z.decompressor = flate.NewReaderDict(z.r, dict)
	} else {
		z.decompressor = flate.NewReader(z.r)
	}
	progress, ok := z.decompressor.(flate.InputProgress)
	if !ok {
		panic("zlib: decompressor does not implement flate.InputProgress")
	}
	z.progress = progress

	z.digest = adler32.New()

	return nil
}
