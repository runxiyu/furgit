package mode_test

import (
	"testing"

	"lindenii.org/go/furgit/object/tree/mode"
)

func TestAppend(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		mode mode.Mode
		want string
	}{
		{mode: mode.Directory, want: "40000"},
		{mode: mode.Regular, want: "100644"},
		{mode: mode.Executable, want: "100755"},
		{mode: mode.Symlink, want: "120000"},
		{mode: mode.Gitlink, want: "160000"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			got := string(tc.mode.Append(nil))
			if got != tc.want {
				t.Fatalf("Append() = %q, want %q", got, tc.want)
			}
		})
	}
}
