package config

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

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
