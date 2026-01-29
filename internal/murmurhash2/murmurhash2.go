// Package murmurhash2 provides a non-cryptographic hash.
package murmurhash2

import "encoding/binary"

// Sum32 computes the MurmurHash2 value for key with the provided seed.
func Sum32(key []byte, seed uint32) uint32 {
	const (
		m uint32 = 0x5bd1e995
		r uint32 = 24
	)

	h := seed ^ uint32(len(key))
	i := 0
	for len(key)-i >= 4 {
		k := binary.LittleEndian.Uint32(key[i:])
		k *= m
		k ^= k >> r
		k *= m

		h *= m
		h ^= k
		i += 4
	}

	switch len(key) - i {
	case 3:
		h ^= uint32(key[i+2]) << 16
		fallthrough
	case 2:
		h ^= uint32(key[i+1]) << 8
		fallthrough
	case 1:
		h ^= uint32(key[i])
		h *= m
	}

	h ^= h >> 13
	h *= m
	h ^= h >> 15
	return h
}
