//go:build amd64 && !purego

package adler32

import (
	"encoding/binary"
	"errors"
	"hash"
	"hash/adler32"

	"codeberg.org/lindenii/furgit/internal/cpu"
)

// digest represents the partial evaluation of a checksum.
// The low 16 bits are s1, the high 16 bits are s2.
type digest uint32

func (d *digest) Reset() { *d = 1 }

// New returns a new hash.Hash32 computing the Adler-32 checksum.
func New() hash.Hash32 {
	if !cpu.X86.HasAVX2 {
		return adler32.New()
	}

	d := new(digest)
	d.Reset()

	return d
}

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, marshaledSize)
	b = append(b, magic...)
	b = binary.BigEndian.AppendUint32(b, uint32(*d))

	return b, nil
}

var (
	ErrInvalidHashStateIdentifier = errors.New("internal/adler32: invalid hash state identifier")
	ErrInvalidHashStateSize       = errors.New("internal/adler32: invalid hash state size")
)

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic) || string(b[:len(magic)]) != magic {
		return ErrInvalidHashStateIdentifier
	}

	if len(b) != marshaledSize {
		return ErrInvalidHashStateSize
	}

	*d = digest(binary.BigEndian.Uint32(b[len(magic):]))

	return nil
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return 4 }

func (d *digest) Write(data []byte) (nn int, err error) {
	if cpu.X86.HasAVX2 && len(data) >= 64 {
		h := adler32_avx2(uint32(*d), data)
		*d = digest(h)
	} else {
		h := update(uint32(*d), data)
		*d = digest(h)
	}

	return len(data), nil
}

func (d *digest) Sum32() uint32 { return uint32(*d) }

func (d *digest) Sum(in []byte) []byte {
	return binary.BigEndian.AppendUint32(in, uint32(*d))
}

// Checksum returns the Adler-32 checksum of data.
func Checksum(data []byte) uint32 {
	if cpu.X86.HasAVX2 && len(data) >= 64 {
		return adler32_avx2(1, data)
	}

	return adler32.Checksum(data)
}
