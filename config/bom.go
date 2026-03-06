package config

import (
	"errors"
	"io"
)

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
