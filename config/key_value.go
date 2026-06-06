package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

func (p *configParser) parseKeyValue(cfg *Config) error {
	if p.currentSection == "" {
		return p.parseError("key-value pair before any section header")
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
			return p.parseError(fmt.Sprintf("invalid character in key: %q", ch))
		}

		key.WriteByte(toLower(ch))
	}

	keyStr := key.String()
	if len(keyStr) == 0 {
		return p.parseError("empty key name")
	}

	if !isLetter(keyStr[0]) {
		return p.parseError("key must start with a letter")
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
			return p.parseError(fmt.Sprintf("unexpected character after key: %q", ch))
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
