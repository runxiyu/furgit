package object

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// Signature represents a Git signature (author/committer/tagger).
type Signature struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}

// ParseSignature parses a canonical Git signature line:
// "Name <email> 123456789 +0000".
func ParseSignature(line []byte) (*Signature, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, errors.New("object: signature: missing opening <")
	}

	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, errors.New("object: signature: missing closing >")
	}

	gt := lt + 1 + gtRel

	nameBytes := append([]byte(nil), bytes.TrimRight(line[:lt], " ")...)
	emailBytes := append([]byte(nil), line[lt+1:gt]...)

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, errors.New("object: signature: missing timestamp separator")
	}

	rest = rest[1:]

	before, after, ok := bytes.Cut(rest, []byte{' '})
	if !ok {
		return nil, errors.New("object: signature: missing timezone separator")
	}

	when, err := strconv.ParseInt(string(before), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("object: signature: invalid timestamp: %w", err)
	}

	tz := after
	if len(tz) < 5 {
		return nil, errors.New("object: signature: invalid timezone encoding")
	}

	sign := 1

	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, errors.New("object: signature: invalid timezone sign")
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
		return nil, errors.New("object: signature: invalid timezone hours range")
	}

	if mm < 0 || mm > 59 {
		return nil, errors.New("object: signature: invalid timezone minutes range")
	}

	total := int64(hh)*60 + int64(mm)

	offset, err := intconv.Int64ToInt32(total)
	if err != nil {
		return nil, errors.New("object: signature: timezone overflow")
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

// Serialize renders the signature in canonical Git format.
func (signature Signature) Serialize() ([]byte, error) {
	var b strings.Builder
	b.Grow(len(signature.Name) + len(signature.Email) + 32)
	b.Write(signature.Name)
	b.WriteString(" <")
	b.Write(signature.Email)
	b.WriteString("> ")
	b.WriteString(strconv.FormatInt(signature.WhenUnix, 10))
	b.WriteByte(' ')

	offset := signature.OffsetMinutes

	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hh := offset / 60
	mm := offset % 60
	fmt.Fprintf(&b, "%c%02d%02d", sign, hh, mm)

	return []byte(b.String()), nil
}

// When returns a time.Time with the signature's timezone offset.
func (signature Signature) When() time.Time {
	loc := time.FixedZone("git", int(signature.OffsetMinutes)*60)

	return time.Unix(signature.WhenUnix, 0).In(loc)
}
