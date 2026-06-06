package name_test

import (
	"testing"

	"lindenii.org/go/furgit/ref/name"
)

func TestIsPseudo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "FETCH_HEAD", want: true},
		{name: "MERGE_HEAD", want: true},
		{name: "HEAD", want: false},
		{name: "AUTO_MERGE", want: false},
	}

	for _, tt := range tests {
		if got := name.IsPseudo(tt.name); got != tt.want {
			t.Fatalf("IsPseudo(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsRootSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "", want: true},
		{name: "HEAD", want: true},
		{name: "AUTO_MERGE", want: true},
		{name: "BISECT-EXPECTED_REV", want: true},
		{name: "refs/heads/main", want: false},
		{name: "Head", want: false},
		{name: "HEAD1", want: false},
	}

	for _, tt := range tests {
		if got := name.IsRootSyntax(tt.name); got != tt.want {
			t.Fatalf("IsRootSyntax(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "HEAD", want: true},
		{name: "ORIG_HEAD", want: true},
		{name: "BOGUS_HEAD", want: true},
		{name: "CHERRY_PICK_HEAD", want: true},
		{name: "REVERT_HEAD", want: true},
		{name: "AUTO_MERGE", want: true},
		{name: "BISECT_EXPECTED_REV", want: true},
		{name: "NOTES_MERGE_PARTIAL", want: true},
		{name: "NOTES_MERGE_REF", want: true},
		{name: "MERGE_AUTOSTASH", want: true},
		{name: "FETCH_HEAD", want: false},
		{name: "MERGE_HEAD", want: false},
		{name: "Head", want: false},
		{name: "refs/heads/main", want: false},
	}

	for _, tt := range tests {
		if got := name.IsRoot(tt.name); got != tt.want {
			t.Fatalf("IsRoot(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
