package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

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
				return "", p.parseError("unexpected EOF in quoted value")
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
				return "", p.parseError("newline in quoted value")
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
				return "", p.parseError("unexpected EOF after backslash")
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
				return "", p.parseError(fmt.Sprintf("invalid escape sequence: \\%c", next))
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

func truncateAtNUL(value string) string {
	for i := range len(value) {
		if value[i] == 0 {
			return value[:i]
		}
	}

	return value
}
