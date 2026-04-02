package signature

import "errors"

// ErrInvalidSignature indicates an attempt to parse an invalid signature.
var ErrInvalidSignature = errors.New("object: signature: invalid signature")
