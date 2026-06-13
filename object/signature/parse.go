package signature

import (
	"bytes"
	"fmt"
	"strconv"

	"lindenii.org/go/lgo/intconv"
)

// Parse parses a canonical Git signature line.
//
// The returned signature aliases line:
// its Name and Email share line's backing array,
// so the signature inherits line's lifetime
// and must not be mutated unless line may be.
// Use [Signature.Clone] for an independent copy.
//
// Labels: Life-Parent, Mut-No.
func Parse(line []byte) (*Signature, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, fmt.Errorf("%w: missing '<' before email", ErrInvalidSignature)
	}

	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, fmt.Errorf("%w: missing '>' after email", ErrInvalidSignature)
	}

	gt := lt + 1 + gtRel

	nameBytes := bytes.TrimRight(line[:lt], " ")
	emailBytes := line[lt+1:gt]

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, fmt.Errorf("%w: missing ' ' before timestamp", ErrInvalidSignature)
	}

	rest = rest[1:]

	before, after, ok := bytes.Cut(rest, []byte{' '})
	if !ok {
		return nil, fmt.Errorf("%w: missing ' ' between timestamp and timezone", ErrInvalidSignature)
	}

	when, err := strconv.ParseInt(string(before), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: timestamp %q: %w", ErrInvalidSignature, before, err)
	}

	tz := after
	if len(tz) < 5 {
		return nil, fmt.Errorf("%w: timezone %q is too short", ErrInvalidSignature, tz)
	}

	sign := 1

	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, fmt.Errorf("%w: timezone sign %q is not '+' or '-'", ErrInvalidSignature, tz[0])
	}

	hh, err := strconv.Atoi(string(tz[1:3]))
	if err != nil {
		return nil, fmt.Errorf("%w: timezone hour %q: %w", ErrInvalidSignature, tz[1:3], err)
	}

	mm, err := strconv.Atoi(string(tz[3:5]))
	if err != nil {
		return nil, fmt.Errorf("%w: timezone minute %q: %w", ErrInvalidSignature, tz[3:5], err)
	}

	if hh < 0 || hh > 23 {
		return nil, fmt.Errorf("%w: timezone hour %d out of range", ErrInvalidSignature, hh)
	}

	if mm < 0 || mm > 59 {
		return nil, fmt.Errorf("%w: timezone minute %d out of range", ErrInvalidSignature, mm)
	}

	total := int64(hh)*60 + int64(mm)

	offset, err := intconv.Int64ToInt32(total)
	if err != nil {
		return nil, fmt.Errorf("%w: timezone offset out of range", ErrInvalidSignature)
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
