// Package config provides configuration parsing.
package config

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
