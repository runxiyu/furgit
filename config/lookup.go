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
	case ValueMissing:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	case ValueValueless:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	case ValueString:
		return r.Value, nil
	default:
		return "", &LookupError{Kind: r.Kind, Operation: "string"}
	}
}

// Bool interprets this lookup result using Git config boolean rules.
func (r LookupResult) Bool() (bool, error) {
	switch r.Kind {
	case ValueMissing:
		return false, &LookupError{Kind: r.Kind, Operation: "bool"}
	case ValueValueless:
		return true, nil
	case ValueString:
		return parseBool(r.Value)
	default:
		return false, &LookupError{Kind: r.Kind, Operation: "bool"}
	}
}

// Int interprets this lookup result as a Git integer value.
func (r LookupResult) Int() (int, error) {
	switch r.Kind {
	case ValueMissing:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	case ValueValueless:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	case ValueString:
		return parseInt(r.Value)
	default:
		return 0, &LookupError{Kind: r.Kind, Operation: "int"}
	}
}

// Int64 interprets this lookup result as a Git int64 value.
func (r LookupResult) Int64() (int64, error) {
	switch r.Kind {
	case ValueMissing:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	case ValueValueless:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	case ValueString:
		return parseInt64(r.Value)
	default:
		return 0, &LookupError{Kind: r.Kind, Operation: "int64"}
	}
}

// Lookup retrieves the first value for a given section, optional subsection,
// and key.
func (c *Config) Lookup(section, subsection, key string) LookupResult {
	section = strings.ToLower(section)

	key = strings.ToLower(key)
	for _, entry := range c.entries {
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
		Kind:  ValueMissing,
		Value: "",
	}
}

// LookupAll retrieves all values for a given section, optional subsection,
// and key.
func (c *Config) LookupAll(section, subsection, key string) []LookupResult {
	section = strings.ToLower(section)
	key = strings.ToLower(key)

	var values []LookupResult

	for _, entry := range c.entries {
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
