package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

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
				return nil, fmt.Errorf("furgit: config: line %d: %w", p.lineNum, err)
			}

			continue
		}

		// Key-value pair
		if isLetter(ch) {
			p.unreadChar(ch)

			err := p.parseKeyValue(cfg)
			if err != nil {
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
			err := p.skipToEOL()
			if err != nil && !errors.Is(err, io.EOF) {
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
	var (
		value     bytes.Buffer
		inQuote   bool
		inComment bool
	)

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
