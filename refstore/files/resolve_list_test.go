package files_test

import (
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
)

func TestFilesResolveAndListOverlay(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, packedID := testRepo.MakeCommit(t, "packed base")
		_, _, looseID := testRepo.MakeCommit(t, "loose override")
		testRepo.UpdateRef(t, "refs/heads/main", packedID)
		testRepo.UpdateRef(t, "refs/tags/v1", packedID)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.UpdateRef(t, "refs/heads/main", looseID)
		testRepo.UpdateRef(t, "refs/heads/dev", looseID)

		store := openFilesStore(t, testRepo, algo)

		resolvedMain, err := store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		mainDet, ok := resolvedMain.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(main) type = %T, want ref.Detached", resolvedMain)
		}

		if mainDet.ID != looseID {
			t.Fatalf("Resolve(main) id = %s, want %s", mainDet.ID, looseID)
		}

		resolvedHead, err := store.Resolve("HEAD")
		if err != nil {
			t.Fatalf("Resolve(HEAD): %v", err)
		}

		headSym, ok := resolvedHead.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(HEAD) type = %T, want ref.Symbolic", resolvedHead)
		}

		if headSym.Target != "refs/heads/main" {
			t.Fatalf("Resolve(HEAD) target = %q, want %q", headSym.Target, "refs/heads/main")
		}

		fullHead, err := store.ResolveToDetached("HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(HEAD): %v", err)
		}

		if fullHead.ID != looseID {
			t.Fatalf("ResolveToDetached(HEAD) = %s, want %s", fullHead.ID, looseID)
		}

		allRefs, err := store.List("")
		if err != nil {
			t.Fatalf("List(\"\"): %v", err)
		}

		names := make([]string, 0, len(allRefs))
		for _, entry := range allRefs {
			names = append(names, entry.Name())
		}

		slices.Sort(names)

		want := []string{"HEAD", "refs/heads/dev", "refs/heads/main", "refs/tags/v1"}
		if !slices.Equal(names, want) {
			t.Fatalf("List(\"\") names = %v, want %v", names, want)
		}
	})
}

func TestFilesLooseRefParsingMatchesGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
		oid := testRepo.HashObject(t, "blob", []byte("payload\n"))

		testRepo.WriteFileAll(t, ".git/refs/heads/no-lf", []byte(oid.String()), 0o755, 0o644)
		testRepo.WriteFileAll(t, ".git/refs/heads/trailing-ws", []byte(oid.String()+"   "), 0o755, 0o644)
		testRepo.WriteFileAll(t, ".git/refs/heads/leading-ws", []byte(" "+oid.String()+"\n"), 0o755, 0o644)
		testRepo.WriteFileAll(t, ".git/refs/heads/sym-trailing", []byte("ref: refs/heads/main   "), 0o755, 0o644)
		testRepo.WriteFileAll(t, ".git/refs/heads/sym-leading", []byte(" ref: refs/heads/main\n"), 0o755, 0o644)

		store := openFilesStore(t, testRepo, algo)

		got, err := store.ResolveToDetached("refs/heads/no-lf")
		if err != nil {
			t.Fatalf("ResolveToDetached(no-lf): %v", err)
		}

		if got.ID != oid {
			t.Fatalf("ResolveToDetached(no-lf) = %s, want %s", got.ID, oid)
		}

		got, err = store.ResolveToDetached("refs/heads/trailing-ws")
		if err != nil {
			t.Fatalf("ResolveToDetached(trailing-ws): %v", err)
		}

		if got.ID != oid {
			t.Fatalf("ResolveToDetached(trailing-ws) = %s, want %s", got.ID, oid)
		}

		_, err = store.Resolve("refs/heads/leading-ws")
		if err == nil {
			t.Fatal("Resolve(leading-ws) unexpectedly succeeded")
		}

		resolved, err := store.Resolve("refs/heads/sym-trailing")
		if err != nil {
			t.Fatalf("Resolve(sym-trailing): %v", err)
		}

		sym, ok := resolved.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(sym-trailing) type = %T, want ref.Symbolic", resolved)
		}

		if sym.Target != "refs/heads/main" {
			t.Fatalf("Resolve(sym-trailing) target = %q, want %q", sym.Target, "refs/heads/main")
		}

		_, err = store.Resolve("refs/heads/sym-leading")
		if err == nil {
			t.Fatal("Resolve(sym-leading) unexpectedly succeeded")
		}
	})
}

func TestFilesRejectMalformedPackedRefs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
		_, _, commitID := testRepo.MakeCommit(t, "packed")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.PackRefs(t, "--all", "--prune")

		hex := commitID.String()

		cases := []struct {
			name    string
			content string
		}{
			{
				name:    "unterminated line",
				content: "# pack-refs with: peeled fully-peeled sorted\n" + hex + " refs/heads/main",
			},
			{
				name:    "junk line",
				content: "# pack-refs with: peeled fully-peeled sorted\nbogus content\n",
			},
			{
				name:    "short oid",
				content: "# pack-refs with: peeled fully-peeled sorted\n" + hex[:7] + " refs/heads/main\n",
			},
			{
				name:    "trailing garbage after oid",
				content: "# pack-refs with: peeled fully-peeled sorted\n" + hex + "xrefs/heads/main\n",
			},
			{
				name:    "malformed peeled line",
				content: "# pack-refs with: peeled fully-peeled sorted\n" + hex + " refs/tags/v1\n^" + hex + " garbage\n",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				testRepo.WriteFile(t, "packed-refs", []byte(tc.content), 0o644)
				store := openFilesStore(t, testRepo, algo)

				_, err := store.List("")
				if err == nil {
					t.Fatal("List unexpectedly succeeded")
				}
			})
		}
	})
}

func TestFilesPackedRefsReadSemanticsMatchGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("stale packed entry is still readable", func(t *testing.T) {
			t.Parallel()

			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
			testRepo.Run(t, "commit", "--allow-empty", "-m", "one")

			oneID, err := objectid.ParseHex(algo, testRepo.Run(t, "rev-parse", "HEAD"))
			if err != nil {
				t.Fatalf("ParseHex(one): %v", err)
			}

			testRepo.Run(t, "tag", "-a", "v1.0", "-m", "v1.0", "HEAD")
			testRepo.PackRefs(t, "--all", "--prune")
			testRepo.Run(t, "checkout", "--orphan", "another")
			testRepo.Run(t, "commit", "--allow-empty", "-m", "two")
			testRepo.Run(t, "checkout", "-B", "main")
			testRepo.Run(t, "branch", "-D", "another")
			testRepo.Run(t, "reflog", "expire", "--expire=now", "--all")
			testRepo.Run(t, "prune")

			store := openFilesStore(t, testRepo, algo)

			got, err := store.ResolveToDetached("refs/heads/main")
			if err != nil {
				t.Fatalf("ResolveToDetached(main): %v", err)
			}

			if got.ID == oneID {
				t.Fatalf("ResolveToDetached(main) unexpectedly returned stale packed id %s", oneID)
			}

			tagRef, err := store.Resolve("refs/tags/v1.0")
			if err != nil {
				t.Fatalf("Resolve(tag): %v", err)
			}

			tagDet, ok := tagRef.(ref.Detached)
			if !ok {
				t.Fatalf("Resolve(tag) type = %T, want ref.Detached", tagRef)
			}

			if tagDet.ID.Size() == 0 {
				t.Fatal("Resolve(tag) returned zero object id")
			}
		})

		t.Run("exact unicode packed ref remains enumerable", func(t *testing.T) {
			t.Parallel()

			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
			_, _, commitID := testRepo.MakeCommit(t, "unicode")
			testRepo.UpdateRef(t, "refs/heads/\ue43f", commitID)
			testRepo.UpdateRef(t, "refs/heads/z", commitID)
			testRepo.PackRefs(t, "--all", "--prune")

			store := openFilesStore(t, testRepo, algo)

			listed, err := store.List("refs/heads/z")
			if err != nil {
				t.Fatalf("List(refs/heads/z): %v", err)
			}

			if len(listed) != 1 {
				t.Fatalf("List(refs/heads/z) len = %d, want 1", len(listed))
			}

			if listed[0].Name() != "refs/heads/z" {
				t.Fatalf("List(refs/heads/z)[0] = %q, want %q", listed[0].Name(), "refs/heads/z")
			}
		})
	})
}
