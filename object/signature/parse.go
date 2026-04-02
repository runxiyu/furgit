package signature

import (
	"bytes"
	"fmt"
	"strconv"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// Parse parses a canonical Git signature line.
func Parse(line []byte) (*Signature, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, fmt.Errorf("%w: missing opening <", ErrInvalidSignature)
	}

	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, fmt.Errorf("%w: missing closing >", ErrInvalidSignature)
	}

	gt := lt + 1 + gtRel

	nameBytes := append([]byte(nil), bytes.TrimRight(line[:lt], " ")...)
	emailBytes := append([]byte(nil), line[lt+1:gt]...)

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, fmt.Errorf("%w: missing timestamp separator", ErrInvalidSignature)
	}

	rest = rest[1:]

	before, after, ok := bytes.Cut(rest, []byte{' '})
	if !ok {
		return nil, fmt.Errorf("%w: missing timezone separator", ErrInvalidSignature)
	}

	when, err := strconv.ParseInt(string(before), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("object: signature: invalid timestamp: %w", err)
	}

	tz := after
	if len(tz) < 5 {
		return nil, fmt.Errorf("%w: invalid timezone encoding", ErrInvalidSignature)
	}

	sign := 1

	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, fmt.Errorf("%w: invalid timezone sign", ErrInvalidSignature)
	}

	hh, err := strconv.Atoi(string(tz[1:3]))
	if err != nil {
		return nil, fmt.Errorf("object: signature: invalid timezone hours: %w", err)
	}

	mm, err := strconv.Atoi(string(tz[3:5]))
	if err != nil {
		return nil, fmt.Errorf("object: signature: invalid timezone minutes: %w", err)
	}

	if hh < 0 || hh > 23 {
		return nil, fmt.Errorf("%w: invalid timezone hours range", ErrInvalidSignature)
	}

	if mm < 0 || mm > 59 {
		return nil, fmt.Errorf("%w: invalid timezone minutes range", ErrInvalidSignature)
	}

	total := int64(hh)*60 + int64(mm)

	offset, err := intconv.Int64ToInt32(total)
	if err != nil {
		return nil, fmt.Errorf("%w: timezone overflow", ErrInvalidSignature)
	}

	if sign < 0 {
		offset = -offset
	}

	return &Signature{
		Name:          nameBytes,
		Email:         emailBytes,
		WhenUnix:      when,
		OffsetMinutes: offset,
	}, nil
}
