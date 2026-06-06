package config

import (
	"strings"
)

// LookupResult is a value returned by Lookup/LookupAll.
type LookupResult struct {
	Kind  Kind
	Value string
}

// String returns the explicit string value.
func (r LookupResult) String() (string, error) {
	switch r.Kind {
	case KindMissing:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	case KindValueless:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	case KindString:
		return r.Value, nil
	default:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	}
}

// Bool interprets this lookup result using Git config boolean rules.
func (r LookupResult) Bool() (bool, error) {
	switch r.Kind {
	case KindMissing:
		return false, &LookupError{Kind: r.Kind, Operation: "bool"}
	case KindValueless:
		return true, nil
	case KindString:
		return parseBool(r.Value)
	default:
		return false, &LookupError{Kind: r.Kind, Operation: "bool"}
	}
}

// Int interprets this lookup result as a Git integer value.
func (r LookupResult) Int() (int, error) {
	switch r.Kind {
	case KindMissing:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	case KindValueless:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	case KindString:
		return parseInt(r.Value)
	default:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	}
}

// Int64 interprets this lookup result as a Git int64 value.
func (r LookupResult) Int64() (int64, error) {
	switch r.Kind {
	case KindMissing:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	case KindValueless:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	case KindString:
		return parseInt64(r.Value)
	default:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	}
}

// Lookup retrieves the first value for a given section, optional subsection,
// and key.
func (config *Config) Lookup(section, subsection, key string) LookupResult {
	section = strings.ToLower(section)

	key = strings.ToLower(key)
	for _, entry := range config.entries {
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
