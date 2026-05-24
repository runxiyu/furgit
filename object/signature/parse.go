package signature

import (
	"bytes"
	"errors"
	"strconv"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// ErrInvalidSignature indicates an attempt to parse an invalid signature.
var ErrInvalidSignature = errors.New("object/signature: invalid signature")

// Parse parses a canonical Git signature line.
//
// Labels: Life-Independent.
func Parse(line []byte) (*Signature, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, ErrInvalidSignature
	}

	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, ErrInvalidSignature
	}

	gt := lt + 1 + gtRel

	nameBytes := append([]byte(nil), bytes.TrimRight(line[:lt], " ")...)
	emailBytes := append([]byte(nil), line[lt+1:gt]...)

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, ErrInvalidSignature
	}

	rest = rest[1:]

	before, after, ok := bytes.Cut(rest, []byte{' '})
	if !ok {
		return nil, ErrInvalidSignature
	}

	when, err := strconv.ParseInt(string(before), 10, 64)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	tz := after
	if len(tz) < 5 {
		return nil, ErrInvalidSignature
	}

	sign := 1

	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, ErrInvalidSignature
	}

	hh, err := strconv.Atoi(string(tz[1:3]))
	if err != nil {
		return nil, ErrInvalidSignature
	}

	mm, err := strconv.Atoi(string(tz[3:5]))
	if err != nil {
		return nil, ErrInvalidSignature
	}

	if hh < 0 || hh > 23 {
		return nil, ErrInvalidSignature
	}

	if mm < 0 || mm > 59 {
		return nil, ErrInvalidSignature
	}

	total := int64(hh)*60 + int64(mm)

	offset, err := intconv.Int64ToInt32(total)
	if err != nil {
		return nil, ErrInvalidSignature
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
