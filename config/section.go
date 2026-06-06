package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (p *configParser) parseSection() error {
	var name bytes.Buffer

	for {
		ch, err := p.nextChar()
		if errors.Is(err, io.EOF) {
			return p.parseError("unexpected EOF in section header")
		}

		if err != nil {
			return err
		}

		if ch == ']' {
			section := name.String()
			if !isValidSection(section) {
				return p.parseError(fmt.Sprintf("invalid section name: %q", section))
			}

			p.currentSection = strings.ToLower(section)
			p.currentSubsec = ""

			return nil
		}

		if isWhitespace(ch) {
			return p.parseExtendedSection(&name)
		}

		if !isKeyChar(ch) && ch != '.' {
			return p.parseError(fmt.Sprintf("invalid character in section name: %q", ch))
		}

		name.WriteByte(toLower(ch))
	}
}
