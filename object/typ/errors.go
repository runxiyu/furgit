package typ

import "errors"

// ErrInvalidType indicates an unknown or unsupported Git object type.
var ErrInvalidType = errors.New("object/typ: invalid type")
