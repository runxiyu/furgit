package id

// ParseObjectFormat parses a canonical object format name (e.g. "sha1", "sha256").
func ParseObjectFormat(s string) (ObjectFormat, error) {
	objectFormat, ok := objectFormatByName[s]
	if !ok {
		return ObjectFormatUnknown, ErrInvalidObjectFormat
	}

	return objectFormat, nil
}
