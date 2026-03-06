package config

import (
	"errors"
	"fmt"
)

// LookupResult is a value returned by Lookup/LookupAll.
type LookupResult struct {
	Kind  ValueKind
	Value string
}

// String returns the explicit string value.
func (r LookupResult) String() (string, error) {
	switch r.Kind {
	case ValueMissing:
		return "", errors.New("missing config value")
	case ValueValueless:
		return "", errors.New("valueless config key")
	case ValueString:
		return r.Value, nil
	default:
		return "", fmt.Errorf("unknown value kind %d", r.Kind)
	}
}

// Bool interprets this lookup result using Git config boolean rules.
func (r LookupResult) Bool() (bool, error) {
	switch r.Kind {
	case ValueMissing:
		return false, errors.New("missing config value")
	case ValueValueless:
		return true, nil
	case ValueString:
		return parseBool(r.Value)
	default:
		return false, fmt.Errorf("unknown value kind %d", r.Kind)
	}
}

// Int interprets this lookup result as a Git integer value.
func (r LookupResult) Int() (int, error) {
	switch r.Kind {
	case ValueMissing:
		return 0, errors.New("missing config value")
	case ValueValueless:
		return 0, errors.New("valueless config key")
	case ValueString:
		return parseInt(r.Value)
	default:
		return 0, fmt.Errorf("unknown value kind %d", r.Kind)
	}
}

// Int64 interprets this lookup result as a Git int64 value.
func (r LookupResult) Int64() (int64, error) {
	switch r.Kind {
	case ValueMissing:
		return 0, errors.New("missing config value")
	case ValueValueless:
		return 0, errors.New("valueless config key")
	case ValueString:
		return parseInt64(r.Value)
	default:
		return 0, fmt.Errorf("unknown value kind %d", r.Kind)
	}
}
