package mode_test

import (
	"testing"

	"lindenii.org/go/furgit/object/tree/mode"
)

func TestParseAppendRoundTrip(t *testing.T) {
	t.Parallel()

	for _, m := range []mode.Mode{mode.Directory, mode.Regular, mode.Executable, mode.Symlink, mode.Gitlink} {
		raw := m.Append(nil)

		got, err := mode.Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%q): %v", raw, err)
		}

		if got != m {
			t.Fatalf("round trip %o -> %q -> %o", m, raw, got)
		}
	}
}
