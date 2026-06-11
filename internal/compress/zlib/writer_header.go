// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib

import (
	"encoding/binary"

	"lindenii.org/go/furgit/internal/adler32"
	"lindenii.org/go/furgit/internal/compress/flate"
)

// writeHeader writes the ZLIB header.
func (z *Writer) writeHeader() (err error) {
	z.wroteHeader = true
	// ZLIB has a two-byte header (as documented in RFC 1950).
	// The first four bits is the CINFO (compression info), which is 7 for the default deflate window size.
	// The next four bits is the CM (compression method), which is 8 for deflate.
	z.scratch[0] = 0x78
	// The next two bits is the FLEVEL (compression level). The four values are:
	// 0=fastest, 1=fast, 2=default, 3=best.
	// The next bit, FDICT, is set if a dictionary is given.
	// The final five FCHECK bits form a mod-31 checksum.
	switch z.level {
	case -2, 0, 1:
		z.scratch[1] = 0 << 6
	case 2, 3, 4, 5:
		z.scratch[1] = 1 << 6
	case 6, -1:
		z.scratch[1] = 2 << 6
	case 7, 8, 9:
		z.scratch[1] = 3 << 6
	default:
		panic("unreachable")
	}

	if z.dict != nil {
		z.scratch[1] |= 1 << 5
	}

	z.scratch[1] += uint8(31 - binary.BigEndian.Uint16(z.scratch[:2])%31) //#nosec G115

	_, err = z.w.Write(z.scratch[0:2])
	if err != nil {
		return err //nolint:wrapcheck
	}

	if z.dict != nil {
		// The next four bytes are the Adler-32 checksum of the dictionary.
		binary.BigEndian.PutUint32(z.scratch[:], adler32.Checksum(z.dict))

		_, err = z.w.Write(z.scratch[0:4])
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	if z.compressor == nil {
		// Initialize deflater unless the Writer is being reused
		// after a Reset call.
		z.compressor, err = flate.NewWriterDict(z.w, z.level, z.dict)
		if err != nil {
			return err //nolint:wrapcheck
		}

		z.digest = adler32.New()
	}

	return nil
}
