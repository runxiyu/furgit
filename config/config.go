// Package config provides configuration parsing.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Config holds all parsed configuration entries from a Git config file.
//
// A Config preserves the ordering of entries as they appeared in the source.
//
// Lookups are matched case-insensitively for section and key names, and
// subsections must match exactly.
//
// Includes aren't supported yet; they will be supported in a later revision.
type Config struct {
	entries []ConfigEntry
}

// ValueKind describes the presence and form of a config value.
type ValueKind uint8

const (
	// ValueMissing means the queried key does not exist.
	ValueMissing ValueKind = iota
	// ValueValueless means the key exists but has no "= <value>" part.
	ValueValueless
	// ValueString means the key exists and has an explicit value (possibly "").
	ValueString
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

// ConfigEntry represents a single parsed configuration directive.
type ConfigEntry struct {
	// The section name in canonical lowercase form.
	Section string
	// The subsection name, retaining the exact form parsed from the input.
	Subsection string
	// The key name in canonical lowercase form.
	Key string
	// Kind records whether this entry has no value or an explicit value.
	Kind ValueKind
	// The interpreted value of the configuration entry, including unescaped
	// characters where appropriate.
	Value string
}

// ParseConfig reads and parses Git configuration entries from r.
func ParseConfig(r io.Reader) (*Config, error) {
	parser := &configParser{
		reader:  bufio.NewReader(r),
		lineNum: 1,
	}

	return parser.parse()
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

	return LookupResult{Kind: ValueMissing}
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

// Entries returns a copy of all parsed configuration entries in the order they
// appeared. Modifying the returned slice does not affect the Config.
func (c *Config) Entries() []ConfigEntry {
	result := make([]ConfigEntry, len(c.entries))
	copy(result, c.entries)

	return result
}
