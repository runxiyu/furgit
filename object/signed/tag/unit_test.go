package signedtag_test

import (
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
	signedtag "lindenii.org/go/furgit/object/signed/tag"
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

	tag, err := signedtag.Parse(body, objectid.AlgorithmSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(tag.AppendPayload(nil))

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

	gotAlgorithms := tag.Algorithms()
	if len(gotAlgorithms) != 2 || gotAlgorithms[0] != objectid.AlgorithmSHA1 || gotAlgorithms[1] != objectid.AlgorithmSHA256 {
		t.Fatalf("algorithms mismatch: got %v", gotAlgorithms)
	}

	gotSignature, ok := tag.AppendSignature(nil, objectid.AlgorithmSHA1)
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

	gotHeaderSignature, ok := tag.AppendSignature(nil, objectid.AlgorithmSHA256)
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

	tag, err := signedtag.Parse(body, objectid.AlgorithmSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(tag.AppendPayload(nil))

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

	gotSignature, ok := tag.AppendSignature(nil, objectid.AlgorithmSHA256)
	if !ok {
		t.Fatal("missing sha256 signature")
	}

	wantSignature := "" +
		"header\n" +
		"continued\n"
	if string(gotSignature) != wantSignature {
		t.Fatalf("signature mismatch:\n got: %q\nwant: %q", string(gotSignature), wantSignature)
	}

	if _, ok := tag.AppendSignature(nil, objectid.AlgorithmSHA1); ok {
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

	tag, err := signedtag.Parse(body, objectid.AlgorithmSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(tag.AppendPayload(nil))

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

	tag, err := signedtag.Parse(body, objectid.AlgorithmSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(tag.AppendPayload(nil))

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

	tag, err := signedtag.Parse(body, objectid.AlgorithmSHA1)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(tag.AppendPayload(nil))

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

	gotSignature, ok := tag.AppendSignature(nil, objectid.AlgorithmSHA1)
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
