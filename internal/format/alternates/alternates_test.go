package alternates_test

import (
	"slices"
	"testing"

	"lindenii.org/go/furgit/internal/format/alternates"
)

func TestParse(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name string
		data string
		want []string
	}{
		{
			name: "empty",
			data: "",
			want: []string{},
		},
		{
			name: "one absolute name",
			data: "/srv/repo.git/objects\n",
			want: []string{"/srv/repo.git/objects"},
		},
		{
			name: "several names",
			data: "/one\n/two\n/three\n",
			want: []string{"/one", "/two", "/three"},
		},
		{
			name: "relative name joins the base",
			data: "../../shared.git/objects\n",
			want: []string{"/repo/shared.git/objects"},
		},
		{
			name: "comments and blanks are skipped",
			data: "# a comment\n\n/one\n\n# another\n/two\n",
			want: []string{"/one", "/two"},
		},
		{
			name: "no trailing newline",
			data: "/one",
			want: []string{"/one"},
		},
		{
			name: "quoted name",
			data: "\"/has space/objects\"\n",
			want: []string{"/has space/objects"},
		},
		{
			name: "quoted name with escapes",
			data: `"/a\tb\\c\"d"` + "\n",
			want: []string{"/a\tb\\c\"d"},
		},
		{
			name: "quoted name with octal escape",
			data: `"/caf\303\251"` + "\n",
			want: []string{"/café"},
		},
		{
			name: "broken quoting is taken literally",
			data: "\"/unterminated\n",
			want: []string{`/repo/one.git/objects/"/unterminated`},
		},
		{
			name: "a name is not a comment past its start",
			data: "/has#hash\n",
			want: []string{"/has#hash"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := alternates.Parse([]byte(testCase.data), "/repo/one.git/objects")
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("Parse = %q, want %q", got, testCase.want)
			}
		})
	}
}
