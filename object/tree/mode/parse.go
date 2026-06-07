package mode

import (
	"errors"
	"fmt"
)

// ErrInvalidMode indicates a malformed or unsupported tree entry mode.
var ErrInvalidMode = errors.New("object/tree/mode: invalid mode")

// Parse decodes a canonical octal tree entry mode.
//
// It accepts only the modes Git itself writes
// and rejects malformed, unsupported, and zero-padded encodings.
func Parse(raw []byte) (Mode, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("%w: empty mode", ErrInvalidMode)
	}

	if raw[0] == '0' {
		return 0, fmt.Errorf("%w: zero-padded mode %q", ErrInvalidMode, raw)
	}

	if len(raw) > MaxModeDigits {
		return 0, fmt.Errorf("%w: mode %q too long", ErrInvalidMode, raw)
	}

	var mode Mode

	for _, c := range raw {
		if c < '0' || c > '7' {
			return 0, fmt.Errorf("%w: non-octal byte in mode %q", ErrInvalidMode, raw)
		}

		mode = mode<<3 | Mode(c-'0')
	}

	if !mode.IsValid() {
		return 0, fmt.Errorf("%w: unsupported mode %q", ErrInvalidMode, raw)
	}

	return mode, nil
}
