package typ

import "errors"

// ErrInvalidType indicates an unknown or unsupported Git object type.
var ErrInvalidType = errors.New("object/typ: invalid type")

// Parse parses a canonical Git object type name.
func Parse(name string) (Type, error) {
	ty, ok := typeByName[name]

	if !ok {
		return 0, ErrInvalidType
	}

	return ty, nil
}
