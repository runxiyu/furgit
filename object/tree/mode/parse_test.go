package mode_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/tree/mode"
)

func TestParse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		raw  string
		want mode.Mode
	}{
		{raw: "40000", want: mode.Directory},
		{raw: "100644", want: mode.Regular},
		{raw: "100755", want: mode.Executable},
		{raw: "120000", want: mode.Symlink},
		{raw: "160000", want: mode.Gitlink},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			t.Parallel()

			got, err := mode.Parse([]byte(tc.raw))
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.raw, err)
			}

			if got != tc.want {
				t.Fatalf("Parse(%q) = %o, want %o", tc.raw, got, tc.want)
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: ""},
		{name: "zero-padded-directory", raw: "040000"},
		{name: "zero-padded-regular", raw: "0100644"},
		{name: "bare-zero", raw: "0"},
		{name: "non-octal-digit", raw: "100648"},
		{name: "non-octal-letter", raw: "10064x"},
		{name: "unsupported", raw: "100640"},
		{name: "unsupported-directory-variant", raw: "40755"},
		{name: "too-long", raw: "1006440"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := mode.Parse([]byte(tc.raw))
			if !errors.Is(err, mode.ErrInvalidMode) {
				t.Fatalf("Parse(%q) error = %v, want ErrInvalidMode", tc.raw, err)
			}
		})
	}
}
