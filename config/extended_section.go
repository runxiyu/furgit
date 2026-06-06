package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (p *configParser) parseExtendedSection(sectionName *bytes.Buffer) error {
	for {
		ch, err := p.nextChar()
		if errors.Is(err, io.EOF) {
			return p.parseError("unexpected EOF in section header")
		}

		if err != nil {
			return err
		}

		if !isWhitespace(ch) {
			if ch != '"' {
				return p.parseError("expected quote after section name")
			}

			break
		}
	}

	var subsec bytes.Buffer

	for {
		ch, err := p.nextChar()
		if errors.Is(err, io.EOF) {
			return p.parseError("unexpected EOF in subsection")
		}

		if err != nil {
			return err
		}

		if ch == '\n' {
			return p.parseError("newline in subsection")
		}

		if ch == '"' {
			break
		}

		if ch == '\\' {
			next, err := p.nextChar()
			if errors.Is(err, io.EOF) {
				return p.parseError("unexpected EOF after backslash in subsection")
			}

			if err != nil {
				return err
			}

			if next == '\n' {
				return p.parseError("newline after backslash in subsection")
			}

			subsec.WriteByte(next)
		} else {
			subsec.WriteByte(ch)
		}
	}

	ch, err := p.nextChar()
	if errors.Is(err, io.EOF) {
		return p.parseError("unexpected EOF after subsection")
	}

	if err != nil {
		return err
	}

	if ch != ']' {
		return p.parseError(fmt.Sprintf("expected ']' after subsection, got %q", ch))
	}

	section := sectionName.String()
	if !isValidSection(section) {
		return p.parseError(fmt.Sprintf("invalid section name: %q", section))
	}

	p.currentSection = strings.ToLower(section)
	p.currentSubsec = subsec.String()

	return nil
}
