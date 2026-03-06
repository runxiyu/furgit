package bloom

import "codeberg.org/lindenii/furgit/internal/intconv"

type key struct {
	hashes []uint32
}

func keyvec(path []byte, filter *Filter) ([]key, error) {
	if len(path) == 0 {
		return nil, nil
	}

	count := 1

	for _, b := range path {
		if b == '/' {
			count++
		}
	}

	keys := make([]key, 0, count)

	full, err := keyFill(path, filter)
	if err != nil {
		return nil, err
	}

	keys = append(keys, full)

	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			k, err := keyFill(path[:i], filter)
			if err != nil {
				return nil, err
			}

			keys = append(keys, k)
		}
	}

	return keys, nil
}

func keyFill(path []byte, filter *Filter) (key, error) {
	const (
		seed0 = 0x293ae76f
		seed1 = 0x7e646e2c
	)

	var (
		h0  uint32
		h1  uint32
		err error
	)

	switch filter.HashVersion {
	case 2:
		h0, err = murmur3SeededV2(seed0, path)
		if err != nil {
			return key{}, err
		}

		h1, err = murmur3SeededV2(seed1, path)
		if err != nil {
			return key{}, err
		}
	case 1:
		h0, err = murmur3SeededV1(seed0, path)
		if err != nil {
			return key{}, err
		}

		h1, err = murmur3SeededV1(seed1, path)
		if err != nil {
			return key{}, err
		}
	default:
		return key{}, ErrInvalid
	}

	hashCount, err := intconv.Uint32ToInt(filter.NumHashes)
	if err != nil {
		return key{}, ErrInvalid
	}

	hashes := make([]uint32, hashCount)
	for i := range hashCount {
		iU32, err := intconv.IntToUint32(i)
		if err != nil {
			return key{}, ErrInvalid
		}

		hashes[i] = h0 + iU32*h1
	}

	return key{hashes: hashes}, nil
}

func filterContainsKey(filter *Filter, key key) bool {
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
