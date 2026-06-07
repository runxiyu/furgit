package tag_test

import (
	"slices"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/signed/tag"
)

func TestParseSignedTag(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"object 04b871796dc0420f8e7561a895b52484b701d51a\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger C O Mitter <committer@example.com> 1465981006 +0000\n" +
		"gpgsig-sha256 -----BEGIN PGP SIGNATURE-----\n" +
		" Version: GnuPG v1\n" +
		" \n" +
		" header-signature\n" +
		" -----END PGP SIGNATURE-----\n" +
		"\n" +
		"signed tag\n" +
		"\n" +
		"signed tag message body\n" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"Version: GnuPG v1\n" +
		"\n" +
		"body-signature\n" +
		"-----END PGP SIGNATURE-----\n")

	parsed, err := tag.Parse(body, id.ObjectFormatSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(parsed.AppendPayload(nil))

	wantPayload := "" +
		"object 04b871796dc0420f8e7561a895b52484b701d51a\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger C O Mitter <committer@example.com> 1465981006 +0000\n" +
		"\n" +
		"signed tag\n" +
		"\n" +
		"signed tag message body\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	gotObjectFormats := parsed.ObjectFormats()

	wantObjectFormats := []id.ObjectFormat{
		id.ObjectFormatSHA1,
		id.ObjectFormatSHA256,
	}
	if !slices.Equal(gotObjectFormats, wantObjectFormats) {
		t.Fatalf("ObjectFormats() = %v, want %v", gotObjectFormats, wantObjectFormats)
	}

	gotSignature, ok := parsed.AppendSignature(nil, id.ObjectFormatSHA1)
	if !ok {
		t.Fatal("missing sha1 signature")
	}

	wantSignature := "" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"Version: GnuPG v1\n" +
		"\n" +
		"body-signature\n" +
		"-----END PGP SIGNATURE-----\n"
	if string(gotSignature) != wantSignature {
		t.Fatalf("signature mismatch:\n got: %q\nwant: %q", string(gotSignature), wantSignature)
	}

	gotHeaderSignature, ok := parsed.AppendSignature(nil, id.ObjectFormatSHA256)
	if !ok {
		t.Fatal("missing sha256 signature")
	}

	wantHeaderSignature := "" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"Version: GnuPG v1\n" +
		"\n" +
		"header-signature\n" +
		"-----END PGP SIGNATURE-----\n"
	if string(gotHeaderSignature) != wantHeaderSignature {
		t.Fatalf("header signature mismatch:\n got: %q\nwant: %q", string(gotHeaderSignature), wantHeaderSignature)
	}
}

func TestParseHeaderOnlyTagStripsHeaderAndKeepsHeaderSignature(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"gpgsig-sha256 header\n" +
		" continued\n" +
		"\n" +
		"message\n")

	parsed, err := tag.Parse(body, id.ObjectFormatSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(parsed.AppendPayload(nil))

	wantPayload := "" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"\n" +
		"message\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	gotSignature, ok := parsed.AppendSignature(nil, id.ObjectFormatSHA256)
	if !ok {
		t.Fatal("missing sha256 signature")
	}

	wantSignature := "" +
		"header\n" +
		"continued\n"
	if string(gotSignature) != wantSignature {
		t.Fatalf("signature mismatch:\n got: %q\nwant: %q", string(gotSignature), wantSignature)
	}

	if _, ok := parsed.AppendSignature(nil, id.ObjectFormatSHA1); ok {
		t.Fatal("unexpected sha1 signature")
	}
}

func TestParseKeepsUnknownHeaderSignatureTextInPayload(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"gpgsig-future header\n" +
		" continued\n" +
		"\n" +
		"message line\n" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"body-signature\n" +
		"-----END PGP SIGNATURE-----\n")

	parsed, err := tag.Parse(body, id.ObjectFormatSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(parsed.AppendPayload(nil))

	wantPayload := "" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"gpgsig-future header\n" +
		" continued\n" +
		"\n" +
		"message line\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}
}

func TestParseKeepsMessageGpgsigTextInPayload(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"\n" +
		"message line\n" +
		"gpgsig-future header\n" +
		" continued\n" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"body-signature\n" +
		"-----END PGP SIGNATURE-----\n")

	parsed, err := tag.Parse(body, id.ObjectFormatSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(parsed.AppendPayload(nil))

	wantPayload := "" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"\n" +
		"message line\n" +
		"gpgsig-future header\n" +
		" continued\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}
}

func TestParseUsesLastSignatureBeginByPrefix(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"\n" +
		"message\n" +
		"-----BEGIN PGP SIGNATURE----- stray\n" +
		"still message\n" +
		"-----BEGIN PGP SIGNATURE----- trailing\n" +
		"body-signature\n")

	parsed, err := tag.Parse(body, id.ObjectFormatSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(parsed.AppendPayload(nil))

	wantPayload := "" +
		"object deadbeef\n" +
		"type commit\n" +
		"tag signedtag\n" +
		"tagger T A Gger <tagger@example.com> 1465981006 +0000\n" +
		"\n" +
		"message\n" +
		"-----BEGIN PGP SIGNATURE----- stray\n" +
		"still message\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	gotSignature, ok := parsed.AppendSignature(nil, id.ObjectFormatSHA1)
	if !ok {
		t.Fatal("missing signature")
	}

	wantSignature := "" +
		"-----BEGIN PGP SIGNATURE----- trailing\n" +
		"body-signature\n"
	if string(gotSignature) != wantSignature {
		t.Fatalf("signature mismatch:\n got: %q\nwant: %q", string(gotSignature), wantSignature)
	}
}
