package packed_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
	"codeberg.org/lindenii/furgit/refstore/packed"
)

func openPackedRefStoreFromRepo(t *testing.T, repoPath string, algo objectid.Algorithm) *packed.Store {
	t.Helper()
	file, err := os.Open(filepath.Join(repoPath, "packed-refs"))
	if err != nil {
		t.Fatalf("open packed-refs: %v", err)
	}
	defer func() { _ = file.Close() }()

	store, err := packed.New(file, algo)
	if err != nil {
		t.Fatalf("packed.New: %v", err)
	}
	return store
}

func TestPackedResolveAndPeeled(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "packed refs commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		tagID := testRepo.TagAnnotated(t, "v1.0.0", commitID, "annotated tag")
		testRepo.PackRefs(t, "--all", "--prune")

		store := openPackedRefStoreFromRepo(t, testRepo.Dir(), algo)

		resolvedMain, err := store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}
		mainDet, ok := resolvedMain.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(main) type = %T, want ref.Detached", resolvedMain)
		}
		if mainDet.ID != commitID {
			t.Fatalf("Resolve(main) id = %s, want %s", mainDet.ID, commitID)
		}

		resolvedTag, err := store.Resolve("refs/tags/v1.0.0")
		if err != nil {
			t.Fatalf("Resolve(tag): %v", err)
		}
		tagDet, ok := resolvedTag.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(tag) type = %T, want ref.Detached", resolvedTag)
		}
		if tagDet.ID != tagID {
			t.Fatalf("Resolve(tag) id = %s, want %s", tagDet.ID, tagID)
		}
		if tagDet.Peeled == nil {
			t.Fatalf("Resolve(tag) peeled = nil, want commit")
		}
		if *tagDet.Peeled != commitID {
			t.Fatalf("Resolve(tag) peeled = %s, want %s", *tagDet.Peeled, commitID)
		}

		fullTag, err := store.ResolveFully("refs/tags/v1.0.0")
		if err != nil {
			t.Fatalf("ResolveFully(tag): %v", err)
		}
		if fullTag.ID != tagDet.ID {
			t.Fatalf("ResolveFully(tag) id = %s, want %s", fullTag.ID, tagDet.ID)
		}

		if _, err := store.Resolve("refs/heads/does-not-exist"); !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(not-found) error = %v", err)
		}
	})
}

func TestPackedListAndShorten(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "packed refs list commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/tags/main", commitID)
		testRepo.UpdateRef(t, "refs/remotes/origin/main", commitID)
		testRepo.PackRefs(t, "--all", "--prune")

		store := openPackedRefStoreFromRepo(t, testRepo.Dir(), algo)

		all, err := store.List("")
		if err != nil {
			t.Fatalf("List(all): %v", err)
		}
		allNames := make([]string, 0, len(all))
		for _, entry := range all {
			allNames = append(allNames, entry.Name())
		}
		slices.Sort(allNames)
		wantAll := []string{"refs/heads/main", "refs/remotes/origin/main", "refs/tags/main"}
		if !slices.Equal(allNames, wantAll) {
			t.Fatalf("List(all) names = %v, want %v", allNames, wantAll)
		}

		filtered, err := store.List("refs/heads/*")
		if err != nil {
			t.Fatalf("List(pattern): %v", err)
		}
		if len(filtered) != 1 || filtered[0].Name() != "refs/heads/main" {
			t.Fatalf("List(refs/heads/*) = %v, want refs/heads/main only", filtered)
		}

		short, err := store.Shorten("refs/heads/main")
		if err != nil {
			t.Fatalf("Shorten(main): %v", err)
		}
		if short != "heads/main" {
			t.Fatalf("Shorten(main) = %q, want %q", short, "heads/main")
		}

		if _, err := store.Shorten("refs/heads/does-not-exist"); !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Shorten(not-found) error = %v", err)
		}
	})
}

func TestPackedParseErrors(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		cases := []struct {
			name string
			data string
		}{
			{
				name: "peeled without ref",
				data: "^" + stringsOfLen("0", algo.HexLen()) + "\n",
			},
			{
				name: "invalid entry",
				data: "not-a-valid-line\n",
			},
			{
				name: "duplicate ref",
				data: stringsOfLen("0", algo.HexLen()) + " refs/heads/main\n" +
					stringsOfLen("1", algo.HexLen()) + " refs/heads/main\n",
			},
		}

		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				if _, err := packed.New(bytes.NewBufferString(tt.data), algo); err == nil {
					t.Fatalf("packed.New expected parse error")
				}
			})
		}
	})
}

func TestPackedNewValidation(t *testing.T) {
	if _, err := packed.New(bytes.NewReader(nil), objectid.AlgorithmUnknown); !errors.Is(err, objectid.ErrInvalidAlgorithm) {
		t.Fatalf("packed.New invalid algorithm error = %v", err)
	}
	if _, err := packed.New(nil, objectid.AlgorithmSHA1); err == nil {
		t.Fatalf("packed.New nil reader expected error")
	}
}

func stringsOfLen(ch string, n int) string {
	return string(bytes.Repeat([]byte(ch), n))
}
