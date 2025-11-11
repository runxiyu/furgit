package furgit

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Ident models an author/committer identity together with its timestamp
// and timezone offset, mirroring the fields that appear in Git objects.
type Ident struct {
	Name          []byte
	Email         []byte
	WhenUnix      int64
	OffsetMinutes int32
}

// parseIdent parses an identity line from the canonical Git format:
// "Name <email> 123456789 +0000".
func parseIdent(line []byte) (*Ident, error) {
	lt := bytes.IndexByte(line, '<')
	if lt < 0 {
		return nil, errors.New("furgit: ident: missing opening <")
	}
	gtRel := bytes.IndexByte(line[lt+1:], '>')
	if gtRel < 0 {
		return nil, errors.New("furgit: ident: missing closing >")
	}
	gt := lt + 1 + gtRel
	nameBytes := append([]byte(nil), line[:lt]...)
	emailBytes := append([]byte(nil), line[lt+1:gt]...)

	rest := line[gt+1:]
	if len(rest) == 0 || rest[0] != ' ' {
		return nil, errors.New("furgit: ident: missing timestamp separator")
	}
	rest = rest[1:]
	sp := bytes.IndexByte(rest, ' ')
	if sp < 0 {
		return nil, errors.New("furgit: ident: missing timezone separator")
	}
	whenStr := string(rest[:sp])
	when, err := strconv.ParseInt(whenStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("furgit: ident: invalid timestamp: %w", err)
	}

	tz := rest[sp+1:]
	if len(tz) < 5 {
		return nil, errors.New("furgit: ident: invalid timezone encoding")
	}
	sign := 1
	switch tz[0] {
	case '-':
		sign = -1
	case '+':
	default:
		return nil, errors.New("furgit: ident: invalid timezone sign")
	}

	hh, err := strconv.Atoi(string(tz[1:3]))
	if err != nil {
		return nil, fmt.Errorf("furgit: ident: invalid timezone hours: %w", err)
	}
	mm, err := strconv.Atoi(string(tz[3:5]))
	if err != nil {
		return nil, fmt.Errorf("furgit: ident: invalid timezone minutes: %w", err)
	}
	if hh < 0 || hh > 23 {
		return nil, errors.New("furgit: ident: invalid timezone hours range")
	}
	if mm < 0 || mm > 59 {
		return nil, errors.New("furgit: ident: invalid timezone minutes range")
	}
	total := int64(hh)*60 + int64(mm)
	if total > math.MaxInt32 {
		return nil, errors.New("furgit: ident: timezone overflow")
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

// Serialize renders an Ident into canonical Git format.
func (id Ident) Serialize() []byte {
	var b strings.Builder
	b.Grow(len(id.Name) + len(id.Email) + 32)
	b.Write(id.Name)
	b.WriteString(" <")
	b.Write(id.Email)
	b.WriteString("> ")
	b.WriteString(strconv.FormatInt(id.WhenUnix, 10))
	b.WriteByte(' ')

	offset := id.OffsetMinutes
	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}
	hh := offset / 60
	mm := offset % 60
	fmt.Fprintf(&b, "%c%02d%02d", sign, hh, mm)
	return []byte(b.String())
}

// When returns the timestamp as time.Time using the embedded offset.
func (id Ident) When() time.Time {
	loc := time.FixedZone("git", int(id.OffsetMinutes)*60)
	return time.Unix(id.WhenUnix, 0).In(loc)
}
