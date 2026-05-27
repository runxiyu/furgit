package files_test

import (
	"errors"
	"slices"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	refstore "lindenii.org/go/furgit/ref/store"
)

func TestFilesWorktreeRefsMatchGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})

		testRepo.Run(t, "commit", "--allow-empty", "-m", "initial")

		initialID, err := objectid.ParseHex(algo, testRepo.Run(t, "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("ParseHex(initial HEAD): %v", err)
		}

		testRepo.Run(t, "branch", "wt1", initialID.String())
		testRepo.Run(t, "branch", "wt2", initialID.String())
		testRepo.Run(t, "worktree", "add", "wt1", "wt1")
		testRepo.Run(t, "worktree", "add", "wt2", "wt2")

		testRepo.Run(t, "-C", "wt1", "commit", "--allow-empty", "-m", "wt1")
		testRepo.Run(t, "-C", "wt2", "commit", "--allow-empty", "-m", "wt2")

		wt1ID, err := objectid.ParseHex(algo, testRepo.Run(t, "-C", "wt1", "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("ParseHex(wt1 HEAD): %v", err)
		}

		wt2ID, err := objectid.ParseHex(algo, testRepo.Run(t, "-C", "wt2", "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("ParseHex(wt2 HEAD): %v", err)
		}

		testRepo.UpdateRef(t, "refs/worktree/foo", initialID)
		testRepo.Run(t, "-C", "wt1", "update-ref", "refs/worktree/foo", wt1ID.String())
		testRepo.Run(t, "-C", "wt2", "update-ref", "refs/worktree/foo", wt2ID.String())

		mainStore := openFilesStore(t, testRepo, algo)
		repoRoot := testRepo.OpenRoot(t)
		wt1Store := openFilesStoreAt(t, openGitRootUnder(t, repoRoot, "wt1"), algo)
		wt2Store := openFilesStoreAt(t, openGitRootUnder(t, repoRoot, "wt2"), algo)

		got, err := mainStore.ResolveToDetached("refs/worktree/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached(main refs/worktree/foo): %v", err)
		}

		if got.ID != initialID {
			t.Fatalf("ResolveToDetached(main refs/worktree/foo) = %s, want %s", got.ID, initialID)
		}

		got, err = wt1Store.ResolveToDetached("refs/worktree/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached(wt1 refs/worktree/foo): %v", err)
		}

		if got.ID != wt1ID {
			t.Fatalf("ResolveToDetached(wt1 refs/worktree/foo) = %s, want %s", got.ID, wt1ID)
		}

		got, err = wt2Store.ResolveToDetached("refs/worktree/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached(wt2 refs/worktree/foo): %v", err)
		}

		if got.ID != wt2ID {
			t.Fatalf("ResolveToDetached(wt2 refs/worktree/foo) = %s, want %s", got.ID, wt2ID)
		}

		got, err = wt1Store.ResolveToDetached("main-worktree/HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(wt1 main-worktree/HEAD): %v", err)
		}

		if got.ID != initialID {
			t.Fatalf("ResolveToDetached(wt1 main-worktree/HEAD) = %s, want %s", got.ID, initialID)
		}

		got, err = mainStore.ResolveToDetached("worktrees/wt1/HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(main worktrees/wt1/HEAD): %v", err)
		}

		if got.ID != wt1ID {
			t.Fatalf("ResolveToDetached(main worktrees/wt1/HEAD) = %s, want %s", got.ID, wt1ID)
		}

		got, err = wt2Store.ResolveToDetached("worktrees/wt1/HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(wt2 worktrees/wt1/HEAD): %v", err)
		}

		if got.ID != wt1ID {
			t.Fatalf("ResolveToDetached(wt2 worktrees/wt1/HEAD) = %s, want %s", got.ID, wt1ID)
		}

		assertListMatchesGitForEachRef(t, testRepo.Run(t, "for-each-ref", "--format=%(refname)"), mainStore)
		assertListMatchesGitForEachRef(t, testRepo.Run(t, "-C", "wt1", "for-each-ref", "--format=%(refname)"), wt1Store)
	})
}

func TestFilesTransactionPerWorktreeRefsMatchGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
		testRepo.Run(t, "commit", "--allow-empty", "-m", "initial")
		testRepo.Run(t, "branch", "wt1", "HEAD")
		testRepo.Run(t, "worktree", "add", "wt1", "wt1")

		mainID, err := objectid.ParseHex(algo, testRepo.Run(t, "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("ParseHex(main HEAD): %v", err)
		}

		testRepo.Run(t, "-C", "wt1", "commit", "--allow-empty", "-m", "wt1")

		wt1ID, err := objectid.ParseHex(algo, testRepo.Run(t, "-C", "wt1", "rev-parse", "HEAD"))
		if err != nil {
			t.Fatalf("ParseHex(wt1 HEAD): %v", err)
		}

		mainStore := openFilesStore(t, testRepo, algo)
		repoRoot := testRepo.OpenRoot(t)
		wt1Store := openFilesStoreAt(t, openGitRootUnder(t, repoRoot, "wt1"), algo)

		mainTx, err := mainStore.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(main): %v", err)
		}

		err = mainTx.Create("refs/bisect/main-only", mainID)
		if err != nil {
			t.Fatalf("Create(main-only) queue: %v", err)
		}

		err = mainTx.Commit()
		if err != nil {
			t.Fatalf("Commit(main-only): %v", err)
		}

		wtTx, err := wt1Store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(wt1): %v", err)
		}

		err = wtTx.Create("refs/bisect/wt-only", wt1ID)
		if err != nil {
			t.Fatalf("Create(wt-only) queue: %v", err)
		}

		err = wtTx.Commit()
		if err != nil {
			t.Fatalf("Commit(wt-only): %v", err)
		}

		got, err := mainStore.ResolveToDetached("refs/bisect/main-only")
		if err != nil {
			t.Fatalf("ResolveToDetached(main-only): %v", err)
		}

		if got.ID != mainID {
			t.Fatalf("ResolveToDetached(main-only) = %s, want %s", got.ID, mainID)
		}

		got, err = wt1Store.ResolveToDetached("refs/bisect/wt-only")
		if err != nil {
			t.Fatalf("ResolveToDetached(wt-only): %v", err)
		}

		if got.ID != wt1ID {
			t.Fatalf("ResolveToDetached(wt-only) = %s, want %s", got.ID, wt1ID)
		}

		_, err = mainStore.Resolve("refs/bisect/wt-only")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(main sees wt-only) error = %v, want ErrReferenceNotFound", err)
		}

		_, err = wt1Store.Resolve("refs/bisect/main-only")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(wt sees main-only) error = %v, want ErrReferenceNotFound", err)
		}

		mainRefs := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(refname)", "refs/bisect"))

		wtRefs := forEachRefLines(testRepo.Run(t, "-C", "wt1", "for-each-ref", "--format=%(refname)", "refs/bisect"))
		if !slices.Equal(mainRefs, []string{"refs/bisect/main-only"}) {
			t.Fatalf("main for-each-ref refs/bisect = %v", mainRefs)
		}

		if !slices.Equal(wtRefs, []string{"refs/bisect/wt-only"}) {
			t.Fatalf("wt1 for-each-ref refs/bisect = %v", wtRefs)
		}
	})
}
