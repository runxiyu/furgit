package packfile

// Signature is the 4-byte "PACK" magic at the start of pack files.
const Signature = 0x5041434b

// VersionSupported reports whether one pack version is supported.
func VersionSupported(version uint32) bool {
	return version == 2 || version == 3
}
