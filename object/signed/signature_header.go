package signed

import "lindenii.org/go/furgit/object/id"

// signatureHeaderNames maps each object format
// to the commit and tag signature header name
// that carries its signature,
// such as "gpgsig" for SHA-1
// and "gpgsig-sha256" for SHA-256.
//
//nolint:gochecknoglobals
var signatureHeaderNames = map[id.ObjectFormat]string{
	id.ObjectFormatSHA1:   "gpgsig",
	id.ObjectFormatSHA256: "gpgsig-sha256",
}

//nolint:gochecknoglobals
var objectFormatBySignatureHeaderName = map[string]id.ObjectFormat{}

func init() { //nolint:gochecknoinits
	for objectFormat, name := range signatureHeaderNames {
		objectFormatBySignatureHeaderName[name] = objectFormat
	}
}

// SignatureHeaderName returns the signature header name for objectFormat,
// such as "gpgsig" for SHA-1
// or "gpgsig-sha256" for SHA-256.
func SignatureHeaderName(objectFormat id.ObjectFormat) (string, bool) {
	name, ok := signatureHeaderNames[objectFormat]

	return name, ok
}

// ParseSignatureHeaderName parses one canonical signature header name
// such as "gpgsig" or "gpgsig-sha256"
// into its object format.
func ParseSignatureHeaderName(name string) (id.ObjectFormat, bool) {
	objectFormat, ok := objectFormatBySignatureHeaderName[name]

	return objectFormat, ok
}
