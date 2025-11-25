// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package zlibx implements reading of zlib format compressed data,
as specified in RFC 1950.

This package differs from the standard library's compress/zlib package
in that it pools readers to reduce allocations. Writing is unsupported.

THis package will likely be refactorered much more for our specific
use case of only doing full decompressions to byte slices.

Note that closing the reader causes it to be returned to a pool for
reuse. Therefore, the caller must not retain references to the
reader after closing it; in the standard library's compress/zlib package,
it is legal to Reset a closed reader and continue using it; that is
not allowed here, so there is simply no Resetter interface.

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
package zlibx

import (
	"errors"
)

const (
	zlibDeflate   = 8
	zlibMaxWindow = 7
)

var (
	// ErrChecksum is returned when reading ZLIB data that has an invalid checksum.
	ErrChecksum = errors.New("zlib: invalid checksum")
	// ErrHeader is returned when reading ZLIB data that has an invalid header.
	ErrHeader = errors.New("zlib: invalid header")
)
