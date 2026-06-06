package name_test

import (
	"testing"

	"lindenii.org/go/furgit/ref/name"
)

func TestIsSafe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "", want: false},
		{name: "HEAD", want: true},
		{name: "MERGE_HEAD", want: true},
		{name: "Head", want: false},
		{name: "refs/heads/main", want: true},
		{name: "refs/", want: false},
		{name: "refs//heads/main", want: false},
		{name: "refs/heads/main/", want: false},
		{name: "refs/foo/../bar", want: false},
		{name: "refs/foo/../../bar", want: false},
		{name: "refs/heads/main.lock", want: true},
	}

	for _, tt := range tests {
		if got := name.IsSafe(tt.name); got != tt.want {
			t.Fatalf("IsSafe(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
