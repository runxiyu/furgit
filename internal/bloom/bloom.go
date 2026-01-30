// Package bloom provides a bloom filter implementation used for changed-path
// filters in Git commit graphs.
package bloom

import "encoding/binary"

const (
	// DataHeaderSize is the size of the BDAT header in commit-graph files.
	DataHeaderSize = 3 * 4
	// DefaultMaxChange matches Git's default max-changed-paths behavior.
	DefaultMaxChange = 512
)

// Settings describe the changed-paths Bloom filter parameters stored in
// commit-graph BDAT chunks.
//
// Obviously, they must match the repository's commit-graph settings to
// interpret filters correctly.
type Settings struct {
	HashVersion    uint32
	NumHashes      uint32
	BitsPerEntry   uint32
	MaxChangePaths uint32
}

// Filter represents a changed-paths Bloom filter associated with a commit.
//
// The filter encodes which paths changed between a commit and its first
// parent. Paths are expected to be in Git's slash-separated form and
// are queried using a path and its prefixes (e.g. "a/b/c", "a/b", "a").
type Filter struct {
	Data    []byte
	Version uint32
}

// ParseSettings reads Bloom filter settings from a BDAT chunk header.
func ParseSettings(bdat []byte) (*Settings, error) {
	if len(bdat) < DataHeaderSize {
		return nil, ErrInvalid
	}
	settings := &Settings{
		HashVersion:    binary.BigEndian.Uint32(bdat[0:4]),
		NumHashes:      binary.BigEndian.Uint32(bdat[4:8]),
		BitsPerEntry:   binary.BigEndian.Uint32(bdat[8:12]),
		MaxChangePaths: DefaultMaxChange,
	}
	return settings, nil
}

// MightContain reports whether the Bloom filter may contain the given path.
//
// Evaluated against the full path and each of its directory prefixes. A true
// result indicates a possible match; false means the path definitely did not
// change.
func (f *Filter) MightContain(path []byte, settings *Settings) bool {
	if f == nil || settings == nil {
		return false
	}
	if len(f.Data) == 0 {
		return false
	}
	keys := keyvec(path, settings)
	for i := range keys {
		if filterContainsKey(f, &keys[i], settings) {
			return true
		}
	}
	return false
}

type key struct {
	hashes []uint32
}

func keyvec(path []byte, settings *Settings) []key {
	if len(path) == 0 {
		return nil
	}
	count := 1
	for _, b := range path {
		if b == '/' {
			count++
		}
	}
	keys := make([]key, 0, count)
	keys = append(keys, keyFill(path, settings))
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			keys = append(keys, keyFill(path[:i], settings))
		}
	}
	return keys
}

func keyFill(path []byte, settings *Settings) key {
	const seed0 = 0x293ae76f
	const seed1 = 0x7e646e2c
	var h0, h1 uint32
	if settings.HashVersion == 2 {
		h0 = murmur3SeededV2(seed0, path)
		h1 = murmur3SeededV2(seed1, path)
	} else {
		h0 = murmur3SeededV1(seed0, path)
		h1 = murmur3SeededV1(seed1, path)
	}
	hashes := make([]uint32, settings.NumHashes)
	for i := uint32(0); i < settings.NumHashes; i++ {
		hashes[i] = h0 + i*h1
	}
	return key{hashes: hashes}
}

func filterContainsKey(filter *Filter, key *key, settings *Settings) bool {
	if filter == nil || key == nil || settings == nil {
		return false
	}
	if len(filter.Data) == 0 {
		return false
	}
	mod := uint64(len(filter.Data)) * 8
	for _, h := range key.hashes {
		idx := uint64(h) % mod
		bytePos := idx / 8
		bit := byte(1 << (idx & 7))
		if filter.Data[bytePos]&bit == 0 {
			return false
		}
	}
	return true
}

func murmur3SeededV2(seed uint32, data []byte) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)
	h := seed
	nblocks := len(data) / 4
	for i := 0; i < nblocks; i++ {
		k := uint32(data[4*i]) |
			(uint32(data[4*i+1]) << 8) |
			(uint32(data[4*i+2]) << 16) |
			(uint32(data[4*i+3]) << 24)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	var k1 uint32
	tail := data[nblocks*4:]
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		h ^= k1
	}

	h ^= uint32(len(data))
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func murmur3SeededV1(seed uint32, data []byte) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)
	h := seed
	nblocks := len(data) / 4
	for i := 0; i < nblocks; i++ {
		b0 := int8(data[4*i])
		b1 := int8(data[4*i+1])
		b2 := int8(data[4*i+2])
		b3 := int8(data[4*i+3])
		k := uint32(b0) |
			(uint32(b1) << 8) |
			(uint32(b2) << 16) |
			(uint32(b3) << 24)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	var k1 uint32
	tail := data[nblocks*4:]
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(int8(tail[2])) << 16
		fallthrough
	case 2:
		k1 ^= uint32(int8(tail[1])) << 8
		fallthrough
	case 1:
		k1 ^= uint32(int8(tail[0]))
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		h ^= k1
	}

	h ^= uint32(len(data))
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}
