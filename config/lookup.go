package config

import "strings"

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
