// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package adler32 implements the Adler-32 checksum.
//
// It is defined in RFC 1950:
//
//	Adler-32 is composed of two sums accumulated per byte: s1 is
//	the sum of all bytes, s2 is the sum of all s1 values. Both sums
//	are done modulo 65521. s1 is initialized to 1, s2 to zero.  The
//	Adler-32 checksum is stored as s2*65536 + s1 in most-
//	significant-byte first (network) order.
package adler32

import (
	"errors"
	"hash"

	"git.sr.ht/~runxiyu/furgit/internal/byteorder"
)

const (
	// mod is the largest prime that is less than 65536.
	mod = 65521
	// nmax is the largest n such that
	// 255 * n * (n+1) / 2 + (n+1) * (mod-1) <= 2^32-1.
	// It is mentioned in RFC 1950 (search for "5552").
	nmax = 5552
)

// The size of an Adler-32 checksum in bytes.
const Size = 4

// digest represents the partial evaluation of a checksum.
// The low 16 bits are s1, the high 16 bits are s2.
type digest uint32

func (d *digest) Reset() { *d = 1 }

// New returns a new hash.Hash32 computing the Adler-32 checksum. Its
// Sum method will lay the value out in big-endian byte order. The
// returned Hash32 also implements [encoding.BinaryMarshaler] and
// [encoding.BinaryUnmarshaler] to marshal and unmarshal the internal
// state of the hash.
func New() hash.Hash32 {
	d := new(digest)
	d.Reset()
	return d
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return 4 }

const (
	magic         = "adl\x01"
	marshaledSize = len(magic) + 4
)

func (d *digest) AppendBinary(b []byte) ([]byte, error) {
	b = append(b, magic...)
	b = byteorder.BEAppendUint32(b, uint32(*d))
	return b, nil
}

func (d *digest) MarshalBinary() ([]byte, error) {
	return d.AppendBinary(make([]byte, 0, marshaledSize))
}

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic) || string(b[:len(magic)]) != magic {
		return errors.New("hash/adler32: invalid hash state identifier")
	}
	if len(b) != marshaledSize {
		return errors.New("hash/adler32: invalid hash state size")
	}
	*d = digest(byteorder.BEUint32(b[len(magic):]))
	return nil
}

func (d *digest) Clone() (hash.Cloner, error) {
	r := *d
	return &r, nil
}

// Add p to the running checksum d.
func update(d digest, p []byte) digest {
	s1, s2 := uint32(d&0xffff), uint32(d>>16)

	for len(p) > 0 {
		var q []byte
		if len(p) > nmax {
			p, q = p[:nmax], p[nmax:]
		}

		for len(p) >= 32 {
			v := p[:32]
			p = p[32:]

			s1 += uint32(v[0])
			s2 += s1
			s1 += uint32(v[1])
			s2 += s1
			s1 += uint32(v[2])
			s2 += s1
			s1 += uint32(v[3])
			s2 += s1
			s1 += uint32(v[4])
			s2 += s1
			s1 += uint32(v[5])
			s2 += s1
			s1 += uint32(v[6])
			s2 += s1
			s1 += uint32(v[7])
			s2 += s1
			s1 += uint32(v[8])
			s2 += s1
			s1 += uint32(v[9])
			s2 += s1
			s1 += uint32(v[10])
			s2 += s1
			s1 += uint32(v[11])
			s2 += s1
			s1 += uint32(v[12])
			s2 += s1
			s1 += uint32(v[13])
			s2 += s1
			s1 += uint32(v[14])
			s2 += s1
			s1 += uint32(v[15])
			s2 += s1
			s1 += uint32(v[16])
			s2 += s1
			s1 += uint32(v[17])
			s2 += s1
			s1 += uint32(v[18])
			s2 += s1
			s1 += uint32(v[19])
			s2 += s1
			s1 += uint32(v[20])
			s2 += s1
			s1 += uint32(v[21])
			s2 += s1
			s1 += uint32(v[22])
			s2 += s1
			s1 += uint32(v[23])
			s2 += s1
			s1 += uint32(v[24])
			s2 += s1
			s1 += uint32(v[25])
			s2 += s1
			s1 += uint32(v[26])
			s2 += s1
			s1 += uint32(v[27])
			s2 += s1
			s1 += uint32(v[28])
			s2 += s1
			s1 += uint32(v[29])
			s2 += s1
			s1 += uint32(v[30])
			s2 += s1
			s1 += uint32(v[31])
			s2 += s1
		}

		for _, x := range p {
			s1 += uint32(x)
			s2 += s1
		}

		s1 %= mod
		s2 %= mod
		p = q
	}
	return digest(s2<<16 | s1)
}

func (d *digest) Write(p []byte) (nn int, err error) {
	*d = update(*d, p)
	return len(p), nil
}

func (d *digest) Sum32() uint32 { return uint32(*d) }

func (d *digest) Sum(in []byte) []byte {
	s := uint32(*d)
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}

// Checksum returns the Adler-32 checksum of data.
func Checksum(data []byte) uint32 { return uint32(update(1, data)) }
