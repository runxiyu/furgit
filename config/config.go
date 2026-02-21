// Package config provides configuration parsing.
package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
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

// ConfigEntry represents a single parsed configuration directive.
type ConfigEntry struct {
	// The section name in canonical lowercase form.
	Section string
	// The subsection name, retaining the exact form parsed from the input.
	Subsection string
	// The key name in canonical lowercase form.
	Key string
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

// Get retrieves the first value for a given section, optional subsection, and key.
// Returns an empty string if not found.
func (c *Config) Get(section, subsection, key string) string {
	section = strings.ToLower(section)
	key = strings.ToLower(key)
	for _, entry := range c.entries {
		if strings.EqualFold(entry.Section, section) &&
			entry.Subsection == subsection &&
			strings.EqualFold(entry.Key, key) {
			return entry.Value
		}
	}
	return ""
}

// GetAll retrieves all values for a given section, optional subsection, and key.
func (c *Config) GetAll(section, subsection, key string) []string {
	section = strings.ToLower(section)
	key = strings.ToLower(key)
	var values []string
	for _, entry := range c.entries {
		if strings.EqualFold(entry.Section, section) &&
			entry.Subsection == subsection &&
			strings.EqualFold(entry.Key, key) {
			values = append(values, entry.Value)
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

type configParser struct {
	reader         *bufio.Reader
	lineNum        int
	currentSection string
	currentSubsec  string
	peeked         rune
	hasPeeked      bool
}

func (p *configParser) parse() (*Config, error) {
	cfg := &Config{}

	if err := p.skipBOM(); err != nil {
		return nil, err
	}

	for {
		ch, err := p.nextChar()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Skip whitespace and newlines
		if ch == '\n' || unicode.IsSpace(ch) {
			continue
		}

		// Comments
		if ch == '#' || ch == ';' {
			if err := p.skipToEOL(); err != nil && err != io.EOF {
				return nil, err
			}
			continue
		}

		// Section header
		if ch == '[' {
			if err := p.parseSection(); err != nil {
				return nil, fmt.Errorf("furgit: config: line %d: %w", p.lineNum, err)
			}
			continue
		}

		// Key-value pair
		if unicode.IsLetter(ch) {
			p.unreadChar(ch)
			if err := p.parseKeyValue(cfg); err != nil {
				return nil, fmt.Errorf("furgit: config: line %d: %w", p.lineNum, err)
			}
			continue
		}

		return nil, fmt.Errorf("furgit: config: line %d: unexpected character %q", p.lineNum, ch)
	}

	return cfg, nil
}

func (p *configParser) nextChar() (rune, error) {
	if p.hasPeeked {
		p.hasPeeked = false
		return p.peeked, nil
	}

	ch, _, err := p.reader.ReadRune()
	if err != nil {
		return 0, err
	}

	if ch == '\r' {
		next, _, err := p.reader.ReadRune()
		if err == nil && next == '\n' {
			ch = '\n'
		} else if err == nil {
			// Weird but ok
			_ = p.reader.UnreadRune()
		}
	}

	if ch == '\n' {
		p.lineNum++
	}

	return ch, nil
}

func (p *configParser) unreadChar(ch rune) {
	p.peeked = ch
	p.hasPeeked = true
	if ch == '\n' && p.lineNum > 1 {
		p.lineNum--
	}
}

func (p *configParser) skipBOM() error {
	first, _, err := p.reader.ReadRune()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if first != '\uFEFF' {
		_ = p.reader.UnreadRune()
	}
	return nil
}

func (p *configParser) skipToEOL() error {
	for {
		ch, err := p.nextChar()
		if err != nil {
			return err
		}
		if ch == '\n' {
			return nil
		}
	}
}

func (p *configParser) parseSection() error {
	var name bytes.Buffer

	for {
		ch, err := p.nextChar()
		if err != nil {
			return errors.New("unexpected EOF in section header")
		}

		if ch == ']' {
			section := name.String()
			if !isValidSection(section) {
				return fmt.Errorf("invalid section name: %q", section)
			}
			p.currentSection = strings.ToLower(section)
			p.currentSubsec = ""
			return nil
		}

		if unicode.IsSpace(ch) {
			return p.parseExtendedSection(&name)
		}

		if !isKeyChar(ch) && ch != '.' {
			return fmt.Errorf("invalid character in section name: %q", ch)
		}

		name.WriteRune(unicode.ToLower(ch))
	}
}

func (p *configParser) parseExtendedSection(sectionName *bytes.Buffer) error {
	for {
		ch, err := p.nextChar()
		if err != nil {
			return errors.New("unexpected EOF in section header")
		}
		if !unicode.IsSpace(ch) {
			if ch != '"' {
				return errors.New("expected quote after section name")
			}
			break
		}
	}

	var subsec bytes.Buffer
	for {
		ch, err := p.nextChar()
		if err != nil {
			return errors.New("unexpected EOF in subsection")
		}

		if ch == '\n' {
			return errors.New("newline in subsection")
		}

		if ch == '"' {
			break
		}

		if ch == '\\' {
			next, err := p.nextChar()
			if err != nil {
				return errors.New("unexpected EOF after backslash in subsection")
			}
			if next == '\n' {
				return errors.New("newline after backslash in subsection")
			}
			subsec.WriteRune(next)
		} else {
			subsec.WriteRune(ch)
		}
	}

	ch, err := p.nextChar()
	if err != nil {
		return errors.New("unexpected EOF after subsection")
	}
	if ch != ']' {
		return fmt.Errorf("expected ']' after subsection, got %q", ch)
	}

	section := sectionName.String()
	if !isValidSection(section) {
		return fmt.Errorf("invalid section name: %q", section)
	}

	p.currentSection = strings.ToLower(section)
	p.currentSubsec = subsec.String()
	return nil
}

func (p *configParser) parseKeyValue(cfg *Config) error {
	if p.currentSection == "" {
		return errors.New("key-value pair before any section header")
	}

	var key bytes.Buffer
	for {
		ch, err := p.nextChar()
		if err != nil {
			return errors.New("unexpected EOF reading key")
		}

		if ch == '=' || ch == '\n' || unicode.IsSpace(ch) {
			p.unreadChar(ch)
			break
		}

		if !isKeyChar(ch) {
			return fmt.Errorf("invalid character in key: %q", ch)
		}

		key.WriteRune(unicode.ToLower(ch))
	}

	keyStr := key.String()
	if len(keyStr) == 0 {
		return errors.New("empty key name")
	}
	if !unicode.IsLetter(rune(keyStr[0])) {
		return errors.New("key must start with a letter")
	}

	for {
		ch, err := p.nextChar()
		if err == io.EOF {
			cfg.entries = append(cfg.entries, ConfigEntry{
				Section:    p.currentSection,
				Subsection: p.currentSubsec,
				Key:        keyStr,
				Value:      "true",
			})
			return nil
		}
		if err != nil {
			return err
		}

		if ch == '\n' {
			cfg.entries = append(cfg.entries, ConfigEntry{
				Section:    p.currentSection,
				Subsection: p.currentSubsec,
				Key:        keyStr,
				Value:      "true",
			})
			return nil
		}

		if ch == '#' || ch == ';' {
			if err := p.skipToEOL(); err != nil && err != io.EOF {
				return err
			}
			cfg.entries = append(cfg.entries, ConfigEntry{
				Section:    p.currentSection,
				Subsection: p.currentSubsec,
				Key:        keyStr,
				Value:      "true",
			})
			return nil
		}

		if ch == '=' {
			break
		}

		if !unicode.IsSpace(ch) {
			return fmt.Errorf("unexpected character after key: %q", ch)
		}
	}

	value, err := p.parseValue()
	if err != nil {
		return err
	}

	cfg.entries = append(cfg.entries, ConfigEntry{
		Section:    p.currentSection,
		Subsection: p.currentSubsec,
		Key:        keyStr,
		Value:      value,
	})

	return nil
}

func (p *configParser) parseValue() (string, error) {
	var value bytes.Buffer
	var inQuote bool
	var inComment bool
	trimLen := 0

	for {
		ch, err := p.nextChar()
		if err == io.EOF {
			if inQuote {
				return "", errors.New("unexpected EOF in quoted value")
			}
			if trimLen > 0 {
				return value.String()[:trimLen], nil
			}
			return value.String(), nil
		}
		if err != nil {
			return "", err
		}

		if ch == '\n' {
			if inQuote {
				return "", errors.New("newline in quoted value")
			}
			if trimLen > 0 {
				return value.String()[:trimLen], nil
			}
			return value.String(), nil
		}

		if inComment {
			continue
		}

		if unicode.IsSpace(ch) && !inQuote {
			if trimLen == 0 && value.Len() > 0 {
				trimLen = value.Len()
			}
			if value.Len() > 0 {
				value.WriteRune(ch)
			}
			continue
		}

		if !inQuote && (ch == '#' || ch == ';') {
			inComment = true
			continue
		}

		if trimLen > 0 {
			trimLen = 0
		}

		if ch == '\\' {
			next, err := p.nextChar()
			if err == io.EOF {
				return "", errors.New("unexpected EOF after backslash")
			}
			if err != nil {
				return "", err
			}

			switch next {
			case '\n':
				continue
			case 'n':
				value.WriteRune('\n')
			case 't':
				value.WriteRune('\t')
			case 'b':
				value.WriteRune('\b')
			case '\\', '"':
				value.WriteRune(next)
			default:
				return "", fmt.Errorf("invalid escape sequence: \\%c", next)
			}
			continue
		}

		if ch == '"' {
			inQuote = !inQuote
			continue
		}

		value.WriteRune(ch)
	}
}

func isValidSection(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '-' && ch != '.' {
			return false
		}
	}
	return true
}

func isKeyChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-'
}
