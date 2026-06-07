package commit_test

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/commit"
	"lindenii.org/go/furgit/object/id"
)

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			tree := strings.Repeat("1", objectFormat.HexLen())
			parent := strings.Repeat("2", objectFormat.HexLen())
			shortTree := strings.Repeat("3", objectFormat.HexLen()-2)
			author := "author Test Author <author@example.org> 1234567890 +0000\n"
			committer := "committer Test Committer <committer@example.org> 1234567890 +0000\n"
			valid := "tree " + tree + "\n" + author + committer + "\nmessage\n"

			for _, tc := range []struct {
				name string
				body string
			}{
				{
					name: "empty",
					body: "",
				},
				{
					name: "missing-tree",
					body: author + committer + "\nmessage\n",
				},
				{
					name: "malformed-tree",
					body: "tree not-an-oid\n" + author + committer + "\nmessage\n",
				},
				{
					name: "short-tree",
					body: "tree " + shortTree + "\n" + author + committer + "\nmessage\n",
				},
				{
					name: "parent-before-tree",
					body: "parent " + parent + "\n" + valid,
				},
				{
					name: "malformed-parent",
					body: "tree " + tree + "\nparent not-an-oid\n" + author + committer + "\nmessage\n",
				},
				{
					name: "missing-author",
					body: "tree " + tree + "\n" + committer + "\nmessage\n",
				},
				{
					name: "malformed-author-missing-lt",
					body: "tree " + tree + "\nauthor Test Author author@example.org> 1234567890 +0000\n" + committer + "\nmessage\n",
				},
				{
					name: "malformed-author-missing-gt",
					body: "tree " + tree + "\nauthor Test Author <author@example.org 1234567890 +0000\n" + committer + "\nmessage\n",
				},
				{
					name: "malformed-author-missing-timestamp",
					body: "tree " + tree + "\nauthor Test Author <author@example.org> +0000\n" + committer + "\nmessage\n",
				},
				{
					name: "malformed-author-bad-timezone",
					body: "tree " + tree + "\nauthor Test Author <author@example.org> 1234567890 UTC\n" + committer + "\nmessage\n",
				},
				{
					name: "missing-committer",
					body: "tree " + tree + "\n" + author + "\nmessage\n",
				},
				{
					name: "malformed-committer-missing-lt",
					body: "tree " + tree + "\n" + author + "committer Test Committer committer@example.org> 1234567890 +0000\n\nmessage\n",
				},
				{
					name: "malformed-committer-missing-gt",
					body: "tree " + tree + "\n" + author + "committer Test Committer <committer@example.org 1234567890 +0000\n\nmessage\n",
				},
				{
					name: "malformed-committer-missing-timestamp",
					body: "tree " + tree + "\n" + author + "committer Test Committer <committer@example.org> +0000\n\nmessage\n",
				},
				{
					name: "malformed-committer-bad-timezone",
					body: "tree " + tree + "\n" + author + "committer Test Committer <committer@example.org> 1234567890 UTC\n\nmessage\n",
				},
				{
					name: "missing-blank-line",
					body: "tree " + tree + "\n" + author + committer,
				},
				{
					name: "header-without-space",
					body: "tree " + tree + "\n" + author + committer + "encoding\n\nmessage\n",
				},
				{
					name: "unknown-header-before-required-fields",
					body: "tree " + tree + "\nencoding UTF-8\n" + author + committer + "\nmessage\n",
				},
				{
					name: "duplicate-tree",
					body: "tree " + tree + "\ntree " + tree + "\n" + author + committer + "\nmessage\n",
				},
				{
					name: "duplicate-author",
					body: "tree " + tree + "\n" + author + author + committer + "\nmessage\n",
				},
				{
					name: "duplicate-committer",
					body: "tree " + tree + "\n" + author + committer + committer + "\nmessage\n",
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					_, err := commit.Parse([]byte(tc.body), objectFormat)
					if !errors.Is(err, commit.ErrInvalidCommit) {
						t.Fatalf("Parse error = %v, want ErrInvalidCommit", err)
					}
				})
			}
		})
	}
}
