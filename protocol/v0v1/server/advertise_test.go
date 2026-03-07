package server_test

import (
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	server "codeberg.org/lindenii/furgit/protocol/v0v1/server"
)

func TestAdvertiseRefsWritesVersionOneHeadCapsAndPeeledTag(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		headID := mustHexID(t, algo, "1")
		tagID := mustHexID(t, algo, "2")
		peeledID := mustHexID(t, algo, "3")
		mainID := mustHexID(t, algo, "4")

		var out bufferWriteFlusher

		session := server.NewSession(
			strings.NewReader(""),
			&out,
			server.Options{
				Version:   server.Version1,
				Algorithm: algo,
			},
		)

		err := session.AdvertiseRefs(server.Advertisement{
			Refs: []server.AdvertisedRef{
				{Name: "refs/tags/v1", ID: tagID, Peeled: &peeledID},
				{Name: "HEAD", ID: headID},
				{Name: "refs/heads/main", ID: mainID},
			},
		}, []string{
			"report-status",
			"delete-refs",
			"object-format=" + algo.String(),
			"agent=furgit-test/1",
		})
		if err != nil {
			t.Fatalf("AdvertiseRefs: %v", err)
		}

		got := out.String()
		wantParts := []string{
			"000eversion 1\n",
			headID.String() + " HEAD\x00report-status delete-refs object-format=" + algo.String() + " agent=furgit-test/1\n",
			mainID.String() + " refs/heads/main\n",
			tagID.String() + " refs/tags/v1\n",
			peeledID.String() + " refs/tags/v1^{}\n",
			"0000",
		}

		for _, part := range wantParts {
			if !strings.Contains(got, part) {
				t.Fatalf("advertisement missing %q in %q", part, got)
			}
		}
	})
}

func TestAdvertiseRefsWritesNoRefsCapabilitiesLine(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		var out bufferWriteFlusher

		session := server.NewSession(
			strings.NewReader(""),
			&out,
			server.Options{
				Algorithm: algo,
			},
		)

		err := session.AdvertiseRefs(server.Advertisement{}, []string{
			"report-status",
			"object-format=" + algo.String(),
		})
		if err != nil {
			t.Fatalf("AdvertiseRefs: %v", err)
		}

		got := out.String()

		want := objectid.Zero(algo).String() + " capabilities^{}\x00report-status object-format=" + algo.String() + "\n"
		if !strings.Contains(got, want) {
			t.Fatalf("unexpected no-refs advertisement %q", got)
		}
	})
}
