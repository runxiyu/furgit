package config

import (
	"slices"
	"strings"
)

// LookupResult is a value returned by Lookup/LookupAll.
type LookupResult struct {
	Kind  Kind
	Value string
}

// String returns the explicit string value.
func (result LookupResult) String() (string, error) {
	switch result.Kind {
	case KindString:
		return result.Value, nil
	case KindValueless:
		return "", ErrValueless
	case KindMissing:
		return "", ErrMissing
	default:
		return "", ErrMissing
	}
}

// Bool interprets this lookup result using Git config boolean rules.
func (result LookupResult) Bool() (bool, error) {
	switch result.Kind {
	case KindString:
		return parseBool(result.Value)
	case KindValueless:
		return true, nil
	case KindMissing:
		return false, ErrMissing
	default:
		return false, ErrMissing
	}
}

// Int interprets this lookup result as a Git integer value.
func (result LookupResult) Int() (int, error) {
	switch result.Kind {
	case KindString:
		return parseInt(result.Value)
	case KindValueless:
		return 0, ErrValueless
	case KindMissing:
		return 0, ErrMissing
	default:
		return 0, ErrMissing
	}
}

// Int64 interprets this lookup result as a Git int64 value.
func (result LookupResult) Int64() (int64, error) {
	switch result.Kind {
	case KindString:
		return parseInt64(result.Value)
	case KindValueless:
		return 0, ErrValueless
	case KindMissing:
		return 0, ErrMissing
	default:
		return 0, ErrMissing
	}
}

// Lookup retrieves the value for a given section,
// optional subsection, and key.
//
// A key declared more than once takes its last declaration,
// which is how a single-valued key is resolved.
// Use [Config.LookupAll] for keys that are meant to hold several values.
func (config *Config) Lookup(section, subsection, key string) LookupResult {
	section = strings.ToLower(section)

	key = strings.ToLower(key)
	for _, entry := range slices.Backward(config.entries) {
		if strings.EqualFold(entry.Section, section) &&
			entry.Subsection == subsection &&
			strings.EqualFold(entry.Key, key) {
			return LookupResult{
				Kind:  entry.Kind,
				Value: entry.Value,
			}
		}
	}

	return LookupResult{
		Kind:  KindMissing,
		Value: "",
	}
}

// LookupAll retrieves all values for a given section, optional subsection,
// and key.
func (config *Config) LookupAll(section, subsection, key string) []LookupResult {
	section = strings.ToLower(section)
	key = strings.ToLower(key)

	var values []LookupResult

	for _, entry := range config.entries {
		if strings.EqualFold(entry.Section, section) &&
			entry.Subsection == subsection &&
			strings.EqualFold(entry.Key, key) {
			values = append(values, LookupResult{
				Kind:  entry.Kind,
				Value: entry.Value,
			})
		}
	}

	return values
}
