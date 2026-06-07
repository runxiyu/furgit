package tag_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signature"
	"lindenii.org/go/furgit/object/tag"
	"lindenii.org/go/furgit/object/typ"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("roundtrip\n"))
			if err != nil {
				t.Fatalf("HashObject(blob): %v", err)
			}

			want := &tag.Tag{
				TargetID:   blobID,
				TargetType: typ.TypeBlob,
				Name:       []byte("roundtrip-tag"),
				Tagger: signature.Signature{
					Name:          []byte("Round Trip Tagger"),
					Email:         []byte("roundtrip-tagger@example.org"),
					WhenUnix:      1234567999,
					OffsetMinutes: 330,
				},
				Message: []byte("roundtrip subject\n\nroundtrip body\n\n"),
				ExtraHeaders: []tag.ExtraHeader{
					{Key: "encoding", Value: []byte("UTF-8")},
					{Key: "x-test-header", Value: []byte("value")},
				},
			}

			rawBody, err := want.AppendWithoutHeader(nil)
			if err != nil {
				t.Fatalf("AppendWithoutHeader: %v", err)
			}

			roundTripID, err := repo.HashObject(t, typ.TypeTag, bytes.NewReader(rawBody))
			if err != nil {
				t.Fatalf("HashObject(tag): %v", err)
			}

			err = repo.Fsck(t, testgit.FsckOptions{
				Strict:     true,
				NoDangling: true,
			}, roundTripID)
			if err != nil {
				t.Fatalf("Fsck: %v", err)
			}

			gitBody, err := repo.CatFile(t, typ.TypeTag, roundTripID)
			if err != nil {
				t.Fatalf("CatFile: %v", err)
			}

			got, err := tag.Parse(gitBody, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			assertTagEqual(t, got, want)
		})
	}
}

func assertTagEqual(t *testing.T, got *tag.Tag, want *tag.Tag) {
	t.Helper()

	if got.TargetID != want.TargetID {
		t.Fatalf("target id = %s, want %s", got.TargetID, want.TargetID)
	}

	if got.TargetType != want.TargetType {
		t.Fatalf("target type = %v, want %v", got.TargetType, want.TargetType)
	}

	if !bytes.Equal(got.Name, want.Name) {
		t.Fatalf("name = %q, want %q", got.Name, want.Name)
	}

	assertSignatureEqual(t, "tagger", got.Tagger, want.Tagger)

	if !bytes.Equal(got.Message, want.Message) {
		t.Fatalf("message = %q, want %q", got.Message, want.Message)
	}

	if !slices.EqualFunc(got.ExtraHeaders, want.ExtraHeaders, func(gotHeader tag.ExtraHeader, wantHeader tag.ExtraHeader) bool {
		return gotHeader.Key == wantHeader.Key && bytes.Equal(gotHeader.Value, wantHeader.Value)
	}) {
		t.Fatalf("extra headers = %+v, want %+v", got.ExtraHeaders, want.ExtraHeaders)
	}
}

func assertSignatureEqual(t *testing.T, name string, got signature.Signature, want signature.Signature) {
	t.Helper()

	if !bytes.Equal(got.Name, want.Name) {
		t.Fatalf("%s name = %q, want %q", name, got.Name, want.Name)
	}

	if !bytes.Equal(got.Email, want.Email) {
		t.Fatalf("%s email = %q, want %q", name, got.Email, want.Email)
	}

	if got.WhenUnix != want.WhenUnix {
		t.Fatalf("%s time = %d, want %d", name, got.WhenUnix, want.WhenUnix)
	}

	if got.OffsetMinutes != want.OffsetMinutes {
		t.Fatalf("%s offset = %d, want %d", name, got.OffsetMinutes, want.OffsetMinutes)
	}
}
