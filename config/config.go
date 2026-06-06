// Package config provides configuration parsing.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
)

// Config holds all parsed configuration entries from a Git config file.
//
// A Config preserves the ordering of entries as they appeared in the source.
//
// Lookups are matched case-insensitively for section and key names,
// and subsections must match exactly.
//
// Includes aren't supported yet; they will be supported in a later revision.
//
// Labels: MT-Safe.
type Config struct {
	entries []Entry
}

// Parse reads and parses Git configuration entries from r.
func Parse(r io.Reader) (*Config, error) {
	parser := &configParser{
		reader:         bufio.NewReader(r),
		lineNum:        1,
		currentSection: "",
		currentSubsec:  "",
		peeked:         0,
		hasPeeked:      false,
	}

	return parser.parse()
}

// Entry represents a single parsed configuration directive.
type Entry struct {
	// Section is the section name in canonical lowercase form.
	Section string

	// Subsection is the subsection name,
	// retaining the exact form parsed from the input.
	Subsection string

	// Key is the key name in canonical lowercase form.
	Key string

	// Kind reports whether this entry has no value or an explicit value.
	Kind Kind

	// Value is the interpreted value of the configuration entry,
	// including unescaped characters where appropriate.
	Value string
}

// Entries returns a copy of all parsed configuration entries
// in the order they appeared.
//
// Modifying the returned slice does not affect the Config.
func (config *Config) Entries() []Entry {
	return slices.Clone(config.entries)
}

type configParser struct {
	reader         *bufio.Reader
	lineNum        int
	currentSection string
	currentSubsec  string
	peeked         byte
	hasPeeked      bool
}

func (p *configParser) parse() (*Config, error) {
	cfg := &Config{
		entries: nil,
	}

	err := p.skipBOM()
	if err != nil {
		return nil, err
	}

	for {
		ch, err := p.nextChar()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		// Skip leading whitespace between entries.
		if isWhitespace(ch) {
			continue
		}

		// Comments
		if ch == '#' || ch == ';' {
			err := p.skipToEOL()
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}

			continue
		}

		// Section header
		if ch == '[' {
			err := p.parseSection()
			if err != nil {
				return nil, err
			}

			continue
		}

		// Key-value pair
		if isLetter(ch) {
			p.unreadChar(ch)

			err := p.parseKeyValue(cfg)
			if err != nil {
				return nil, err
			}

			continue
		}

		return nil, p.parseError(fmt.Sprintf("unexpected character %q", ch))
	}

	return cfg, nil
}

func (p *configParser) parseError(reason string) error {
	return &ParseError{
		Line:   p.lineNum,
		reason: reason,
	}
}
