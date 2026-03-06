package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// ParseConfig reads and parses Git configuration entries from r.
func ParseConfig(r io.Reader) (*Config, error) {
	parser := &configParser{
		reader:  bufio.NewReader(r),
		lineNum: 1,
	}

	return parser.parse()
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
