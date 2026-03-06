package config

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

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
