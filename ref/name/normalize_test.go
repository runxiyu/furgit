package name_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/ref/name"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	type testCase struct {
		name string
		opts name.Options //exhaustruct:optional
	}

	tests := []testCase{
		{name: "/"},
		{name: "/", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "///refs///heads//main"},
		{name: "refs////tags///v1"},
		{name: "refs///heads///"},
		{name: "HEAD", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "refs/heads/*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "refs///heads/foo"},
		{name: "/heads/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "///heads/foo"},
		{name: "heads/foo/../bar"},
		{name: "heads/./foo"},
		{name: "heads\\foo"},
		{name: "heads/foo.lock"},
		{name: "heads///foo.lock"},
		{name: "foo.lock/bar"},
		{name: "foo.lock///bar"},
		{name: "foo"},
		{name: "/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.opts.String(), func(t *testing.T) {
			t.Parallel()

			got, err := name.Normalize(tt.name, tt.opts)
			want, gitErr := repo.NormalizeRefFormat(t, tt.name, testgit.RefFormatOptions(tt.opts))

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("Normalize(%q, %+v) err=%v, git err=%v", tt.name, tt.opts, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("Normalize(%q, %+v) = %q, want %q", tt.name, tt.opts, got, want)
			}
		})
	}
}
