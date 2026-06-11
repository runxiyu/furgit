package mode_test

import (
	"testing"

	"lindenii.org/go/furgit/object/tree/mode"
	"lindenii.org/go/furgit/object/typ"
)

func TestObjectType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		mode mode.Mode
		want typ.Type
	}{
		{mode: mode.Directory, want: typ.Tree},
		{mode: mode.Regular, want: typ.Blob},
		{mode: mode.Executable, want: typ.Blob},
		{mode: mode.Symlink, want: typ.Blob},
		{mode: mode.Gitlink, want: typ.Commit},
		{mode: mode.Mode(0), want: typ.Unknown},
	} {
		if got := tc.mode.ObjectType(); got != tc.want {
			t.Fatalf("Mode(%o).ObjectType() = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

func TestPredicates(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		mode          mode.Mode
		valid         bool
		isBlobLike    bool
		isRegularFile bool
	}{
		{mode: mode.Directory, valid: true, isBlobLike: false, isRegularFile: false},
		{mode: mode.Regular, valid: true, isBlobLike: true, isRegularFile: true},
		{mode: mode.Executable, valid: true, isBlobLike: true, isRegularFile: true},
		{mode: mode.Symlink, valid: true, isBlobLike: true, isRegularFile: false},
		{mode: mode.Gitlink, valid: true, isBlobLike: false, isRegularFile: false},
		{mode: mode.Mode(0), valid: false, isBlobLike: false, isRegularFile: false},
		{mode: mode.Mode(0o100640), valid: false, isBlobLike: false, isRegularFile: false},
	} {
		if got := tc.mode.IsValid(); got != tc.valid {
			t.Fatalf("Mode(%o).IsValid() = %v, want %v", tc.mode, got, tc.valid)
		}

		if got := tc.mode.IsBlobLike(); got != tc.isBlobLike {
			t.Fatalf("Mode(%o).IsBlobLike() = %v, want %v", tc.mode, got, tc.isBlobLike)
		}

		if got := tc.mode.IsRegularFile(); got != tc.isRegularFile {
			t.Fatalf("Mode(%o).IsRegularFile() = %v, want %v", tc.mode, got, tc.isRegularFile)
		}
	}
}

func TestHasSameType(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		a    mode.Mode
		b    mode.Mode
		want bool
	}{
		{a: mode.Regular, b: mode.Regular, want: true},
		{a: mode.Regular, b: mode.Executable, want: true},
		{a: mode.Executable, b: mode.Regular, want: true},
		{a: mode.Regular, b: mode.Symlink, want: false},
		{a: mode.Directory, b: mode.Directory, want: true},
		{a: mode.Directory, b: mode.Regular, want: false},
		{a: mode.Symlink, b: mode.Symlink, want: true},
		{a: mode.Gitlink, b: mode.Gitlink, want: true},
		{a: mode.Gitlink, b: mode.Directory, want: false},
	} {
		if got := tc.a.HasSameType(tc.b); got != tc.want {
			t.Fatalf("Mode(%o).HasSameType(%o) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
