package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"lindenii.org/go/lgo/intconv"
)

// Kind describes the presence and form of a config value.
type Kind uint8

const (
	// KindMissing means the queried key does not exist.
	KindMissing Kind = iota

	// KindValueless means the key exists but has no "= <value>" part.
	KindValueless

	// KindString means the key exists and has an explicit value (possibly "").
	KindString
)

func isValidSection(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := range len(s) {
		ch := s[i]
		if !isLetter(ch) && !isDigit(ch) && ch != '-' && ch != '.' {
			return false
		}
	}

	return true
}

func isKeyChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-'
}

func parseBool(value string) (bool, error) {
	switch {
	case strings.EqualFold(value, "true"),
		strings.EqualFold(value, "yes"),
		strings.EqualFold(value, "on"):
		return true, nil
	case strings.EqualFold(value, "false"),
		strings.EqualFold(value, "no"),
		strings.EqualFold(value, "off"),
		value == "":
		return false, nil
	}

	n, err := parseInt32(value)
	if err != nil {
		return false, err
	}

	return n != 0, nil
}

func parseInt32(value string) (int32, error) {
	n64, err := parseInt64WithMax(value, math.MaxInt32)
	if err != nil {
		return 0, err
	}

	n32, err := intconv.Int64ToInt32(n64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrValueRange, value)
	}

	return n32, nil
}

func parseInt(value string) (int, error) {
	n64, err := parseInt64WithMax(value, int64(int(^uint(0)>>1)))
	if err != nil {
		return 0, err
	}

	return int(n64), nil
}

func parseInt64(value string) (int64, error) {
	return parseInt64WithMax(value, int64(^uint64(0)>>1))
}

func parseInt64WithMax(value string, maxValue int64) (int64, error) {
	if value == "" {
		return 0, ErrValueEmpty
	}

	trimmed := strings.TrimLeft(value, " \t\n\r\f\v")
	if trimmed == "" {
		return 0, fmt.Errorf("%w: %q", ErrValueEmpty, value)
	}

	numPart := trimmed
	factor := int64(1)

	if last := trimmed[len(trimmed)-1]; last == 'k' || last == 'K' || last == 'm' || last == 'M' || last == 'g' || last == 'G' {
		switch toLower(last) {
		case 'k':
			factor = 1024
		case 'm':
			factor = 1024 * 1024
		case 'g':
			factor = 1024 * 1024 * 1024
		}

		numPart = trimmed[:len(trimmed)-1]
	}

	if numPart == "" {
		return 0, fmt.Errorf("%w: %q", ErrValueSyntax, value)
	}

	n, err := strconv.ParseInt(numPart, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q: %w", ErrValueSyntax, value, err)
	}

	intMax := maxValue
	intMin := -maxValue - 1

	if n > 0 && n > intMax/factor {
		return 0, fmt.Errorf("%w: %q", ErrValueRange, value)
	}

	if n < 0 && n < intMin/factor {
		return 0, fmt.Errorf("%w: %q", ErrValueRange, value)
	}

	n *= factor

	return n, nil
}

func truncateAtNUL(value string) string {
	for i := range len(value) {
		if value[i] == 0 {
			return value[:i]
		}
	}

	return value
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\v' || ch == '\f'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func toLower(ch byte) byte {
	if ch >= 'A' && ch <= 'Z' {
		return ch + ('a' - 'A')
	}

	return ch
}
