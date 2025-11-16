package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"unsafe"
)

// HashType is a constraint that enumerates the supported hash sizes.
type HashType interface {
	[sha1.Size]byte | [sha256.Size]byte
}

type (
	// SHA1Hash represents a SHA-1 hash.
	SHA1Hash = [sha1.Size]byte

	// SHA256Hash represents a SHA-256 hash.
	SHA256Hash = [sha256.Size]byte
)

// Hash represents a Git object identifier with type-level hash size.
type Hash[T HashType] struct {
	v T
}

// hashLen returns the hash size for a given hash type.
func hashLen[T HashType]() int {
	var zero T
	return len(zero)
}

// String returns the hash as a hex string.
func (h Hash[T]) String() string {
	return hex.EncodeToString(unsafe.Slice((*byte)(unsafe.Pointer(&h.v)), hashLen[T]()))
}

// Bytes returns a mutable copy of the underlying bytes.
func (h Hash[T]) Bytes() []byte {
	s := unsafe.Slice((*byte)(unsafe.Pointer(&h.v)), hashLen[T]())
	return append([]byte(nil), s...)
}

// Slice returns a read-only slice view of the underlying bytes.
func (h *Hash[T]) Slice() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&h.v)), hashLen[T]())
}

// ParseHash converts a hex string into a Hash for the given hash type.
func ParseHash[T HashType](s string) (Hash[T], error) {
	var out Hash[T]
	wantHex := hashLen[T]() * 2

	if len(s) != wantHex {
		return out, fmt.Errorf("furgit: invalid hash length %d, want %d", len(s), wantHex)
	}
	raw, err := hex.DecodeString(s)
	if err != nil {
		return out, fmt.Errorf("furgit: decode hash: %w", err)
	}
	slice := unsafe.Slice((*byte)(unsafe.Pointer(&out.v)), hashLen[T]())
	copy(slice, raw)
	return out, nil
}

// computeRawHash computes a hash from data.
func computeRawHash[T HashType](data []byte) Hash[T] {
	var out Hash[T]
	slice := unsafe.Slice((*byte)(unsafe.Pointer(&out.v)), hashLen[T]())

	switch hashLen[T]() {
	case sha1.Size:
		sum := sha1.Sum(data)
		copy(slice, sum[:])
	case sha256.Size:
		sum := sha256.Sum256(data)
		copy(slice, sum[:])
	default:
		panic("furgit: unsupported hash length")
	}
	return out
}
