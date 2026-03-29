package packfile

// Signature is the 4-byte "PACK" magic at the start of pack files.
const Signature = 0x5041434b

// SupportedVersion reports whether one pack version is supported.
func SupportedVersion(version uint32) bool {
	return version == 2 || version == 3
}
