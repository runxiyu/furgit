package id

// ParseObjectFormat parses a canonical object format name (e.g. "sha1", "sha256").
func ParseObjectFormat(s string) (ObjectFormat, bool) {
	objectFormat, ok := objectFormatByName[s]

	return objectFormat, ok
}
