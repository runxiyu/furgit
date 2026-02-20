package object

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Ident represents a Git identity (author/committer/tagger).
type Ident struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}

// ParseIdent parses a canonical Git identity line:
// "Name <email> 123456789 +0000".
func ParseIdent(line []byte) (*Ident, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, errors.New("object: ident: missing opening <")
	}
	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, errors.New("object: ident: missing closing >")
	}
	gt := lt + 1 + gtRel

	nameBytes := append([]byte(nil), bytes.TrimRight(line[:lt], " ")...)
	emailBytes := append([]byte(nil), line[lt+1:gt]...)

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, errors.New("object: ident: missing timestamp separator")
	}
	rest = rest[1:]
	before, after, ok := bytes.Cut(rest, []byte{' '})
	if !ok {
		return nil, errors.New("object: ident: missing timezone separator")
	}
	when, err := strconv.ParseInt(string(before), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("object: ident: invalid timestamp: %w", err)
	}

	tz := after
	if len(tz) < 5 {
		return nil, errors.New("object: ident: invalid timezone encoding")
	}
	sign := 1
	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, errors.New("object: ident: invalid timezone sign")
	}

	hh, err := strconv.Atoi(string(tz[1:3]))
	if err != nil {
		return nil, fmt.Errorf("object: ident: invalid timezone hours: %w", err)
	}
	mm, err := strconv.Atoi(string(tz[3:5]))
	if err != nil {
		return nil, fmt.Errorf("object: ident: invalid timezone minutes: %w", err)
	}
	if hh < 0 || hh > 23 {
		return nil, errors.New("object: ident: invalid timezone hours range")
	}
	if mm < 0 || mm > 59 {
		return nil, errors.New("object: ident: invalid timezone minutes range")
	}
	total := int64(hh)*60 + int64(mm)
	if total > math.MaxInt32 {
		return nil, errors.New("object: ident: timezone overflow")
	}

	offset := int32(total)
	if sign < 0 {
		offset = -offset
	}
	return &Ident{
		Name:          nameBytes,
		Email:         emailBytes,
		WhenUnix:      when,
		OffsetMinutes: offset,
	}, nil
}

// Serialize renders the identity in canonical Git format.
func (ident Ident) Serialize() ([]byte, error) {
	var b strings.Builder
	b.Grow(len(ident.Name) + len(ident.Email) + 32)
	b.Write(ident.Name)
	b.WriteString(" <")
	b.Write(ident.Email)
	b.WriteString("> ")
	b.WriteString(strconv.FormatInt(ident.WhenUnix, 10))
	b.WriteByte(' ')

	offset := ident.OffsetMinutes
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

// When returns a time.Time with the identity's timezone offset.
func (ident Ident) When() time.Time {
	loc := time.FixedZone("git", int(ident.OffsetMinutes)*60)
	return time.Unix(ident.WhenUnix, 0).In(loc)
}
