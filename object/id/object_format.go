package id

// ObjectFormat identifies the Git object format,
// i.e., the hash algorithm used in git object IDs.
type ObjectFormat uint8

const (
	// ObjectFormatUnknown identifies an unknown object format.
	ObjectFormatUnknown ObjectFormat = iota

	// ObjectFormatSHA1 identifies the SHA-1 object format.
	// This is the default for all versions of Git until Git 3.0.
	ObjectFormatSHA1

	// ObjectFormatSHA256 identifies the SHA-256 object format.
	// This is the default for Git 3.0 and beyond.
	ObjectFormatSHA256
)

// SupportedObjectFormats returns all object formats supported by furgit.
//
// Labels: Mut-No.
func SupportedObjectFormats() []ObjectFormat {
	return supportedObjectFormats
}
