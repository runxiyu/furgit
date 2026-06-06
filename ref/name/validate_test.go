package name_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/ref/name"
)

func TestValidate(t *testing.T) {
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
		{name: ""},
		{name: "/"},
		{name: "/", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "foo/bar/baz"},
		{name: "refs/heads/main"},
		{name: "refs/tags/v1.0.0"},
		{name: "refs///heads/foo"},
		{name: "heads/foo/"},
		{name: "/heads/foo"},
		{name: "///heads/foo"},
		{name: "./foo"},
		{name: "./foo/bar"},
		{name: "foo/./bar"},
		{name: "foo/bar/."},
		{name: ".refs/foo"},
		{name: "refs/heads/foo."},
		{name: "HEAD"},
		{name: "HEAD", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "refs/heads/.main"},
		{name: "heads/foo..bar"},
		{name: "refs/heads/main.lock"},
		{name: "heads///foo.lock"},
		{name: "refs/heads/foo..bar"},
		{name: "refs/heads/foo bar"},
		{name: "refs/heads/foo@{bar"},
		{name: "heads/foo?bar"},
		{name: "foo./bar"},
		{name: "foo.lock/bar"},
		{name: "foo.lock///bar"},
		{name: "heads/foo@bar"},
		{name: "heads/foo\\bar"},
		{name: "heads/foo\tbar"},
		{name: "heads/foo\x7fbar"},
		{name: "heads/fu\xC3\x9F"},
		{name: "heads/*foo/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "heads/foo*/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "heads/f*o/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "heads/f*o*/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "heads/foo*/bar*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "refs/heads/foo/bar."},
		{name: "refs//heads///main"},
		{name: "foo"},
		{name: "foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "foo", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/bar"},
		{name: "foo/bar", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "foo/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo/bar", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "refs/heads/*"},
		{name: "refs/heads/*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "refs/heads/feature*branch", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "refs/heads/foo*bar*baz", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo/*"},
		{name: "foo/*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo/*", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "foo/*", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/foo"},
		{name: "*/foo", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "*/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "*/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/*/bar"},
		{name: "foo/*/bar", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo/*/bar", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "foo/*/bar", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*"},
		{name: "*", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "*", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/*/*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "foo/*/*", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/foo/*", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "*/foo/*", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/*/foo", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "*/*/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "/foo"},
		{name: "/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: false}},
		{name: "/foo", opts: name.Options{AllowOneLevel: false, RefspecPattern: true}},
		{name: "/foo", opts: name.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "@"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.opts.String(), func(t *testing.T) {
			t.Parallel()

			err := name.Validate(tt.name, tt.opts)
			gitErr := repo.CheckRefFormat(t, tt.name, testgit.RefFormatOptions(tt.opts))

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("Validate(%q, %+v) err=%v, git err=%v", tt.name, tt.opts, err, gitErr)
			}
		})
	}
}

func TestSanitizeComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		component string
		want      string
	}{
		{component: ".", want: "-"},
		{component: "..", want: "-"},
		{component: "foo..bar", want: "foo.bar"},
		{component: "foo.lock", want: "foo"},
		{component: "foo.lock.lock", want: "foo"},
		{component: "foo bar", want: "foo-bar"},
		{component: "@", want: "-/@"},
		{component: "a@{b", want: "a@-b"},
		{component: "a*b", want: "a-b"},
	}

	for _, tt := range tests {
		if got := name.SanitizeComponent(tt.component); got != tt.want {
			t.Fatalf("SanitizeComponent(%q) = %q, want %q", tt.component, got, tt.want)
		}
	}
}
