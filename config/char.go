package config

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
