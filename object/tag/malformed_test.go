package tag_test

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tag"
	refname "lindenii.org/go/furgit/ref/name"
)

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			objectID := strings.Repeat("1", objectFormat.HexLen())
			shortID := strings.Repeat("2", objectFormat.HexLen()-2)
			object := "object " + objectID + "\n"
			typ := "type commit\n"
			name := "tag v1\n"

			tagger := "tagger Test Tagger <tagger@example.org> 1234567890 +0000\n"
			for _, tc := range []struct {
				name string
				body string
			}{
				{
					name: "empty",
					body: "",
				},
				{
					name: "missing-object",
					body: typ + name + tagger + "\nmessage\n",
				},
				{
					name: "malformed-object",
					body: "object not-an-oid\n" + typ + name + tagger + "\nmessage\n",
				},
				{
					name: "short-object",
					body: "object " + shortID + "\n" + typ + name + tagger + "\nmessage\n",
				},
				{
					name: "missing-type",
					body: object + name + tagger + "\nmessage\n",
				},
				{
					name: "bad-type",
					body: object + "type widget\n" + name + tagger + "\nmessage\n",
				},
				{
					name: "missing-tag",
					body: object + typ + tagger + "\nmessage\n",
				},
				{
					name: "bad-tag-name",
					body: object + typ + "tag bad tag\n" + tagger + "\nmessage\n",
				},
				{
					name: "missing-tagger",
					body: object + typ + name + "\nmessage\n",
				},
				{
					name: "bad-tagger",
					body: object + typ + name + "tagger Test Tagger <tagger@example.org> UTC\n\nmessage\n",
				},
				{
					name: "duplicate-object",
					body: object + typ + name + tagger + object + "\nmessage\n",
				},
				{
					name: "duplicate-type",
					body: object + typ + name + tagger + typ + "\nmessage\n",
				},
				{
					name: "duplicate-tag",
					body: object + typ + name + tagger + name + "\nmessage\n",
				},
				{
					name: "duplicate-tagger",
					body: object + typ + name + tagger + tagger + "\nmessage\n",
				},
				{
					name: "extra-header-without-space",
					body: object + typ + name + tagger + "encoding\n\nmessage\n",
				},
				{
					name: "nonempty-message-without-blank-line",
					body: object + typ + name + tagger + "message\n",
				},
				{
					name: "unterminated-signature-continuation",
					body: object + typ + name + tagger + "gpgsig header\n continuation",
				},
				{
					name: "unterminated-extra-header",
					body: object + typ + name + tagger + "encoding UTF-8",
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					_, err := tag.Parse([]byte(tc.body), objectFormat)
					if !errors.Is(err, tag.ErrInvalidTag) {
						t.Fatalf("Parse error = %v, want ErrInvalidTag", err)
					}
				})
			}

			t.Run("bad-tag-name-wraps-ref-name-error", func(t *testing.T) {
				t.Parallel()

				_, err := tag.Parse([]byte(object+typ+"tag bad tag\n"+tagger+"\nmessage\n"), objectFormat)
				if !errors.Is(err, tag.ErrInvalidTag) {
					t.Fatalf("Parse error = %v, want ErrInvalidTag", err)
				}

				if !errors.Is(err, refname.ErrInvalidName) {
					t.Fatalf("Parse error = %v, want ErrInvalidName", err)
				}
			})
		})
	}
}
