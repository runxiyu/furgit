package signature

import "bytes"

// Clone returns a deep copy of the signature
// whose Name and Email are independent of any memory the original may alias.
//
// Labels: Life-Independent.
func (signature Signature) Clone() Signature {
	return Signature{
		Name:          bytes.Clone(signature.Name),
		Email:         bytes.Clone(signature.Email),
		WhenUnix:      signature.WhenUnix,
		OffsetMinutes: signature.OffsetMinutes,
	}
}
