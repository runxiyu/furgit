package bloom

// MightContain reports whether the Bloom filter may contain the given path.
//
// Evaluated against the full path and each of its directory prefixes. A true
// result indicates a possible match; false means the path definitely did not
// change.
func (f *Filter) MightContain(path []byte, settings *Settings) (bool, error) {
	if f == nil || settings == nil {
		return false, nil
	}

	if len(f.Data) == 0 {
		return false, nil
	}

	keys, err := keyvec(path, settings)
	if err != nil {
		return false, err
	}

	for i := range keys {
		if filterContainsKey(f, &keys[i], settings) {
			return true, nil
		}
	}

	return false, nil
}
