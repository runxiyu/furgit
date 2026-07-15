package files //nolint:testpackage

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
)

func TestParsePackedRefsPeelTraits(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			tagID := objectFormat.Sum([]byte("tag"))
			headID := objectFormat.Sum([]byte("head"))
			peeledID := objectFormat.Sum([]byte("peeled"))

			entries := tagID.String() + " refs/tags/v1\n" +
				headID.String() + " refs/heads/main\n"

			t.Run("no header", func(t *testing.T) {
				t.Parallel()

				packed, err := parsePackedRefs(strings.NewReader(entries), objectFormat)
				if err != nil {
					t.Fatalf("parsePackedRefs: %v", err)
				}

				if got := packed.byName["refs/tags/v1"].PeelState; got != ref.PeelUnknown {
					t.Fatalf("tag PeelState = %d, want PeelUnknown", got)
				}

				if got := packed.byName["refs/heads/main"].PeelState; got != ref.PeelUnknown {
					t.Fatalf("head PeelState = %d, want PeelUnknown", got)
				}
			})

			t.Run("fully-peeled", func(t *testing.T) {
				t.Parallel()

				content := "# pack-refs with: peeled fully-peeled sorted \n" + entries

				packed, err := parsePackedRefs(strings.NewReader(content), objectFormat)
				if err != nil {
					t.Fatalf("parsePackedRefs: %v", err)
				}

				if got := packed.byName["refs/tags/v1"].PeelState; got != ref.PeelNone {
					t.Fatalf("tag PeelState = %d, want PeelNone", got)
				}

				if got := packed.byName["refs/heads/main"].PeelState; got != ref.PeelNone {
					t.Fatalf("head PeelState = %d, want PeelNone", got)
				}

				if !packed.traits.peeled || !packed.traits.fullyPeeled {
					t.Fatalf("traits = %+v, want peeled and fully-peeled", packed.traits)
				}
			})

			t.Run("peeled only", func(t *testing.T) {
				t.Parallel()

				content := "# pack-refs with: peeled \n" + entries

				packed, err := parsePackedRefs(strings.NewReader(content), objectFormat)
				if err != nil {
					t.Fatalf("parsePackedRefs: %v", err)
				}

				if got := packed.byName["refs/tags/v1"].PeelState; got != ref.PeelNone {
					t.Fatalf("tag PeelState = %d, want PeelNone", got)
				}

				if got := packed.byName["refs/heads/main"].PeelState; got != ref.PeelUnknown {
					t.Fatalf("head PeelState = %d, want PeelUnknown", got)
				}
			})

			t.Run("explicit peel line", func(t *testing.T) {
				t.Parallel()

				content := tagID.String() + " refs/tags/v1\n" +
					"^" + peeledID.String() + "\n"

				packed, err := parsePackedRefs(strings.NewReader(content), objectFormat)
				if err != nil {
					t.Fatalf("parsePackedRefs: %v", err)
				}

				entry := packed.byName["refs/tags/v1"]
				if entry.PeelState != ref.PeelTo {
					t.Fatalf("PeelState = %d, want PeelTo", entry.PeelState)
				}

				if entry.PeeledID != peeledID {
					t.Fatalf("PeeledID = %v, want %v", entry.PeeledID, peeledID)
				}
			})
		})
	}
}

func TestParsePackedRefsMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			oid := objectFormat.Sum([]byte("x")).String()
			entry := oid + " refs/heads/a\n"

			cases := []struct {
				name    string
				content string
			}{
				{name: "peel line first", content: "^" + oid + "\n"},
				{name: "duplicate reference", content: entry + entry},
				{name: "unterminated line", content: oid + " refs/heads/a"},
				{name: "comment after first line", content: entry + "# comment\n"},
				{name: "malformed header", content: "# hello\n"},
				{name: "short entry", content: "abc\n"},
				{name: "blank line", content: entry + "\n"},
				{name: "missing separator", content: oid + "refs/heads/a\n"},
				{
					name:    "invalid object id",
					content: strings.Repeat("z", objectFormat.HexLen()) + " refs/heads/a\n",
				},
				{name: "short peel line", content: entry + "^abc\n"},
				{
					name:    "duplicate peel line",
					content: entry + "^" + oid + "\n" + "^" + oid + "\n",
				},
			}

			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()

					_, err := parsePackedRefs(strings.NewReader(testCase.content), objectFormat)
					if !errors.Is(err, errInvalidPackedRefs) {
						t.Fatalf("parsePackedRefs err = %v, want errInvalidPackedRefs", err)
					}
				})
			}
		})
	}
}
