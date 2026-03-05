package bloom

type key struct {
	hashes []uint32
}

func keyvec(path []byte, settings *Settings) ([]key, error) {
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

	full, err := keyFill(path, settings)
	if err != nil {
		return nil, err
	}

	keys = append(keys, full)

	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			k, err := keyFill(path[:i], settings)
			if err != nil {
				return nil, err
			}

			keys = append(keys, k)
		}
	}

	return keys, nil
}

func keyFill(path []byte, settings *Settings) (key, error) {
	const (
		seed0 = 0x293ae76f
		seed1 = 0x7e646e2c
	)

	var (
		h0  uint32
		h1  uint32
		err error
	)

	if settings.HashVersion == 2 { //nolint:nestif
		h0, err = murmur3SeededV2(seed0, path)
		if err != nil {
			return key{}, err
		}

		h1, err = murmur3SeededV2(seed1, path)
		if err != nil {
			return key{}, err
		}
	} else {
		h0, err = murmur3SeededV1(seed0, path)
		if err != nil {
			return key{}, err
		}

		h1, err = murmur3SeededV1(seed1, path)
		if err != nil {
			return key{}, err
		}
	}

	hashes := make([]uint32, settings.NumHashes)
	for i := range settings.NumHashes {
		hashes[i] = h0 + i*h1
	}

	return key{hashes: hashes}, nil
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
