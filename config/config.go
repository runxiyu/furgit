// Package config provides configuration parsing.
package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"codeberg.org/lindenii/furgit/internal/intconv"
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

type configParser struct {
	reader         *bufio.Reader
	lineNum        int
	currentSection string
	currentSubsec  string
	peeked         byte
	hasPeeked      bool
}

func (p *configParser) parse() (*Config, error) {
	cfg := &Config{}

	if err := p.skipBOM(); err != nil {
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
			if err := p.skipToEOL(); err != nil && !errors.Is(err, io.EOF) {
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
		if isLetter(ch) {
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

func (p *configParser) nextChar() (byte, error) {
	if p.hasPeeked {
		p.hasPeeked = false
		return p.peeked, nil
	}

	ch, err := p.reader.ReadByte()
	if err != nil {
		return 0, err
	}

	if ch == '\r' {
		next, err := p.reader.ReadByte()
		if err == nil && next == '\n' {
			ch = '\n'
		} else if err == nil {
			// Weird but ok
			_ = p.reader.UnreadByte()
		}
	}

	if ch == '\n' {
		p.lineNum++
	}

	return ch, nil
}

func (p *configParser) unreadChar(ch byte) {
	p.peeked = ch
	p.hasPeeked = true
	if ch == '\n' && p.lineNum > 1 {
		p.lineNum--
	}
}

func (p *configParser) skipBOM() error {
	first, err := p.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	if first != 0xef {
		_ = p.reader.UnreadByte()
		return nil
	}
	second, err := p.reader.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			_ = p.reader.UnreadByte()
			return nil
		}
		return err
	}
	third, err := p.reader.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			_ = p.reader.UnreadByte()
			_ = p.reader.UnreadByte()
			return nil
		}
		return err
	}
	if second == 0xbb && third == 0xbf {
		return nil
	}
	_ = p.reader.UnreadByte()
	_ = p.reader.UnreadByte()
	_ = p.reader.UnreadByte()
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

		if isWhitespace(ch) {
			return p.parseExtendedSection(&name)
		}

		if !isKeyChar(ch) && ch != '.' {
			return fmt.Errorf("invalid character in section name: %q", ch)
		}

		name.WriteByte(toLower(ch))
	}
}

func (p *configParser) parseExtendedSection(sectionName *bytes.Buffer) error {
	for {
		ch, err := p.nextChar()
		if err != nil {
			return errors.New("unexpected EOF in section header")
		}
		if !isWhitespace(ch) {
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
			subsec.WriteByte(next)
		} else {
			subsec.WriteByte(ch)
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
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		if ch == '=' || ch == '\n' || isSpace(ch) {
			p.unreadChar(ch)
			break
		}

		if !isKeyChar(ch) {
			return fmt.Errorf("invalid character in key: %q", ch)
		}

		key.WriteByte(toLower(ch))
	}

	keyStr := key.String()
	if len(keyStr) == 0 {
		return errors.New("empty key name")
	}
	if !isLetter(keyStr[0]) {
		return errors.New("key must start with a letter")
	}

	for {
		ch, err := p.nextChar()
		if errors.Is(err, io.EOF) {
			cfg.entries = append(cfg.entries, ConfigEntry{
				Section:    p.currentSection,
				Subsection: p.currentSubsec,
				Key:        keyStr,
				Kind:       ValueValueless,
				Value:      "",
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
				Kind:       ValueValueless,
				Value:      "",
			})
			return nil
		}

		if ch == '#' || ch == ';' {
			if err := p.skipToEOL(); err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			cfg.entries = append(cfg.entries, ConfigEntry{
				Section:    p.currentSection,
				Subsection: p.currentSubsec,
				Key:        keyStr,
				Kind:       ValueValueless,
				Value:      "",
			})
			return nil
		}

		if ch == '=' {
			break
		}

		if !isSpace(ch) {
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
		Kind:       ValueString,
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
		if errors.Is(err, io.EOF) {
			if inQuote {
				return "", errors.New("unexpected EOF in quoted value")
			}
			if trimLen > 0 {
				return truncateAtNUL(value.String()[:trimLen]), nil
			}
			return truncateAtNUL(value.String()), nil
		}
		if err != nil {
			return "", err
		}

		if ch == '\n' {
			if inQuote {
				return "", errors.New("newline in quoted value")
			}
			if trimLen > 0 {
				return truncateAtNUL(value.String()[:trimLen]), nil
			}
			return truncateAtNUL(value.String()), nil
		}

		if inComment {
			continue
		}

		if isWhitespace(ch) && !inQuote {
			if trimLen == 0 && value.Len() > 0 {
				trimLen = value.Len()
			}
			if value.Len() > 0 {
				value.WriteByte(ch)
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
			if errors.Is(err, io.EOF) {
				return "", errors.New("unexpected EOF after backslash")
			}
			if err != nil {
				return "", err
			}

			switch next {
			case '\n':
				continue
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case 'b':
				value.WriteByte('\b')
			case '\\', '"':
				value.WriteByte(next)
			default:
				return "", fmt.Errorf("invalid escape sequence: \\%c", next)
			}
			continue
		}

		if ch == '"' {
			inQuote = !inQuote
			continue
		}

		value.WriteByte(ch)
	}
}

func isValidSection(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := range len(s) {
		ch := s[i]
		if !isLetter(ch) && !isDigit(ch) && ch != '-' && ch != '.' {
			return false
		}
	}
	return true
}

func isKeyChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-'
}

func parseBool(value string) (bool, error) {
	switch {
	case strings.EqualFold(value, "true"),
		strings.EqualFold(value, "yes"),
		strings.EqualFold(value, "on"):
		return true, nil
	case strings.EqualFold(value, "false"),
		strings.EqualFold(value, "no"),
		strings.EqualFold(value, "off"),
		value == "":
		return false, nil
	}

	n, err := parseInt32(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value %q", value)
	}
	return n != 0, nil
}

func parseInt32(value string) (int32, error) {
	n64, err := parseInt64WithMax(value, math.MaxInt32)
	if err != nil {
		return 0, err
	}
	return intconv.Int64ToInt32(n64)
}

func parseInt(value string) (int, error) {
	n64, err := parseInt64WithMax(value, int64(int(^uint(0)>>1)))
	if err != nil {
		return 0, err
	}
	return int(n64), nil
}

func parseInt64(value string) (int64, error) {
	return parseInt64WithMax(value, int64(^uint64(0)>>1))
}

func parseInt64WithMax(value string, maxValue int64) (int64, error) {
	if value == "" {
		return 0, errors.New("empty value")
	}

	trimmed := strings.TrimLeft(value, " \t\n\r\f\v")
	if trimmed == "" {
		return 0, errors.New("empty value")
	}

	numPart := trimmed
	factor := int64(1)
	if last := trimmed[len(trimmed)-1]; last == 'k' || last == 'K' || last == 'm' || last == 'M' || last == 'g' || last == 'G' {
		switch toLower(last) {
		case 'k':
			factor = 1024
		case 'm':
			factor = 1024 * 1024
		case 'g':
			factor = 1024 * 1024 * 1024
		}
		numPart = trimmed[:len(trimmed)-1]
	}
	if numPart == "" {
		return 0, errors.New("missing integer value")
	}

	n, err := strconv.ParseInt(numPart, 0, 64)
	if err != nil {
		return 0, err
	}

	intMax := maxValue
	intMin := -maxValue - 1
	if n > 0 && n > intMax/factor {
		return 0, errors.New("integer overflow")
	}
	if n < 0 && n < intMin/factor {
		return 0, errors.New("integer overflow")
	}

	n *= factor
	return n, nil
}

func truncateAtNUL(value string) string {
	for i := range len(value) {
		if value[i] == 0 {
			return value[:i]
		}
	}
	return value
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func toLower(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}
	return ch
}
