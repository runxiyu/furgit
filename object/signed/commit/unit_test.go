package signedcommit_test

import (
	"slices"
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
	signedcommit "codeberg.org/lindenii/furgit/object/signed/commit"
)

func TestParseUpstreamMultiplySignedCommit(t *testing.T) {
	t.Parallel()

	// t/t7510-signed-commit.sh
	body := []byte("" +
		"tree 0cfbf08886fca9a91cb753ec8734c84fcbe52c9f\n" +
		"parent 9da738312d24ef0a29be2c8c2b6fc5cf7085a293\n" +
		"author A U Thor <author@example.com> 1112912653 -0700\n" +
		"committer C O Mitter <committer@example.com> 1112912653 -0700\n" +
		"gpgsig -----BEGIN PGP SIGNATURE-----\n" +
		" \n" +
		" iHQEABECADQWIQRz11h0S+chaY7FTocTtvUezd5DDQUCX/uBDRYcY29tbWl0dGVy\n" +
		" QGV4YW1wbGUuY29tAAoJEBO29R7N3kMNd+8AoK1I8mhLHviPH+q2I5fIVgPsEtYC\n" +
		" AKCTqBh+VabJceXcGIZuF0Ry+udbBQ==\n" +
		" =tQ0N\n" +
		" -----END PGP SIGNATURE-----\n" +
		"gpgsig-sha256 -----BEGIN PGP SIGNATURE-----\n" +
		" \n" +
		" iHQEABECADQWIQRz11h0S+chaY7FTocTtvUezd5DDQUCX/uBIBYcY29tbWl0dGVy\n" +
		" QGV4YW1wbGUuY29tAAoJEBO29R7N3kMN/NEAn0XO9RYSBj2dFyozi0JKSbssYMtO\n" +
		" AJwKCQ1BQOtuwz//IjU8TiS+6S4iUw==\n" +
		" =pIwP\n" +
		" -----END PGP SIGNATURE-----\n" +
		"\n" +
		"second\n")

	commit, err := signedcommit.Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(commit.AppendPayload(nil))

	wantPayload := "" +
		"tree 0cfbf08886fca9a91cb753ec8734c84fcbe52c9f\n" +
		"parent 9da738312d24ef0a29be2c8c2b6fc5cf7085a293\n" +
		"author A U Thor <author@example.com> 1112912653 -0700\n" +
		"committer C O Mitter <committer@example.com> 1112912653 -0700\n" +
		"\n" +
		"second\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	gotSHA1, ok := commit.AppendSignature(nil, objectid.AlgorithmSHA1)
	if !ok {
		t.Fatal("missing sha1 signature")
	}

	wantSHA1 := "" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"\n" +
		"iHQEABECADQWIQRz11h0S+chaY7FTocTtvUezd5DDQUCX/uBDRYcY29tbWl0dGVy\n" +
		"QGV4YW1wbGUuY29tAAoJEBO29R7N3kMNd+8AoK1I8mhLHviPH+q2I5fIVgPsEtYC\n" +
		"AKCTqBh+VabJceXcGIZuF0Ry+udbBQ==\n" +
		"=tQ0N\n" +
		"-----END PGP SIGNATURE-----\n"
	if string(gotSHA1) != wantSHA1 {
		t.Fatalf("sha1 signature mismatch:\n got: %q\nwant: %q", string(gotSHA1), wantSHA1)
	}

	gotSHA256, ok := commit.AppendSignature(nil, objectid.AlgorithmSHA256)
	if !ok {
		t.Fatal("missing sha256 signature")
	}

	wantSHA256 := "" +
		"-----BEGIN PGP SIGNATURE-----\n" +
		"\n" +
		"iHQEABECADQWIQRz11h0S+chaY7FTocTtvUezd5DDQUCX/uBIBYcY29tbWl0dGVy\n" +
		"QGV4YW1wbGUuY29tAAoJEBO29R7N3kMN/NEAn0XO9RYSBj2dFyozi0JKSbssYMtO\n" +
		"AJwKCQ1BQOtuwz//IjU8TiS+6S4iUw==\n" +
		"=pIwP\n" +
		"-----END PGP SIGNATURE-----\n"
	if string(gotSHA256) != wantSHA256 {
		t.Fatalf("sha256 signature mismatch:\n got: %q\nwant: %q", string(gotSHA256), wantSHA256)
	}

	gotAlgorithms := commit.Algorithms()

	wantAlgorithms := []objectid.Algorithm{
		objectid.AlgorithmSHA1,
		objectid.AlgorithmSHA256,
	}
	if !slices.Equal(gotAlgorithms, wantAlgorithms) {
		t.Fatalf("Algorithms() = %v, want %v", gotAlgorithms, wantAlgorithms)
	}
}

func TestParseStripsUnknownGpgsigHeadersFromPayload(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"tree deadbeef\n" +
		"gpgsig-future header\n" +
		" continued\n" +
		"\n" +
		"message\n")

	commit, err := signedcommit.Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(commit.AppendPayload(nil))

	wantPayload := "" +
		"tree deadbeef\n" +
		"\n" +
		"message\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	if gotAlgorithms := commit.Algorithms(); len(gotAlgorithms) != 0 {
		t.Fatalf("Algorithms() = %v, want none", gotAlgorithms)
	}
}

func TestParseAllowsDuplicateSignatureHeaders(t *testing.T) {
	t.Parallel()

	body := []byte("" +
		"tree deadbeef\n" +
		"gpgsig one\n" +
		" two\n" +
		"gpgsig three\n" +
		" four\n" +
		"\n" +
		"message\n")

	commit, err := signedcommit.Parse(body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	gotPayload := string(commit.AppendPayload(nil))

	wantPayload := "" +
		"tree deadbeef\n" +
		"\n" +
		"message\n"
	if gotPayload != wantPayload {
		t.Fatalf("payload mismatch:\n got: %q\nwant: %q", gotPayload, wantPayload)
	}

	gotSignature, ok := commit.AppendSignature(nil, objectid.AlgorithmSHA1)
	if !ok {
		t.Fatal("missing sha1 signature")
	}

	wantSignature := "" +
		"one\n" +
		"two\n" +
		"three\n" +
		"four\n"
	if string(gotSignature) != wantSignature {
		t.Fatalf("signature mismatch:\n got: %q\nwant: %q", string(gotSignature), wantSignature)
	}
}
