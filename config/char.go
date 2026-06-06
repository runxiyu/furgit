package config

import (
	"errors"
	"fmt"
	"io"
)

func (p *configParser) nextChar() (byte, error) {
	if p.hasPeeked {
		p.hasPeeked = false

		return p.peeked, nil
	}

	ch, err := p.reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("config: read character: %w", err)
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

func (p *configParser) skipToEOL() error {
	for {
		ch, err := p.nextChar()
		if err != nil {
			return fmt.Errorf("config: skip to end of line: %w", err)
		}

		if ch == '\n' {
			return nil
		}
	}
}

func (p *configParser) skipBOM() error {
	first, err := p.reader.ReadByte()
	if errors.Is(err, io.EOF) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("config: read byte order mark: %w", err)
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

		return fmt.Errorf("config: read byte order mark: %w", err)
	}

	third, err := p.reader.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			_ = p.reader.UnreadByte()
			_ = p.reader.UnreadByte()

			return nil
		}

		return fmt.Errorf("config: read byte order mark: %w", err)
	}

	if second == 0xbb && third == 0xbf {
		return nil
	}

	_ = p.reader.UnreadByte()
	_ = p.reader.UnreadByte()
	_ = p.reader.UnreadByte()

	return nil
}
