package signature

import "errors"

// ErrInvalidSignature indicates a malformed signature.
var ErrInvalidSignature = errors.New("object/signature: invalid signature")
