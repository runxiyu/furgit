package config

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

// Entries returns a copy of all parsed configuration entries in the order they
// appeared. Modifying the returned slice does not affect the Config.
func (c *Config) Entries() []ConfigEntry {
	result := make([]ConfigEntry, len(c.entries))
	copy(result, c.entries)

	return result
}
