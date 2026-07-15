package repository_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/furgit/repository"
)

func addAlternate(t *testing.T, repo, other *testgit.Repo) {
	t.Helper()

	writeAlternates(t, repo, objectsPath(t, other)+"\n")
}

// makeObjectsDir builds one bare objects directory naming the given alternates.
func makeObjectsDir(t *testing.T, dir string, names ...string) string {
	t.Helper()

	err := os.MkdirAll(filepath.Join(dir, "info"), 0o750)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if len(names) > 0 {
		err = os.WriteFile(
			filepath.Join(dir, "info", "alternates"),
			[]byte(strings.Join(names, "\n")+"\n"), 0o600,
		)
		if err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}

	return resolved
}

func TestResolveAlternatesSelfIsExcluded(t *testing.T) {
	t.Parallel()

	dir := makeObjectsDir(t, filepath.Join(t.TempDir(), "objects"))

	makeObjectsDir(t, dir, dir)

	got := repository.ResolveAlternates(dir)
	if len(got) != 0 {
		t.Fatalf("ResolveAlternates = %q, want none", got)
	}
}

func TestResolveAlternatesRepeatsAreDropped(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	other := makeObjectsDir(t, filepath.Join(base, "other"))
	dir := makeObjectsDir(t, filepath.Join(base, "objects"), other, other, other)

	got := repository.ResolveAlternates(dir)
	if !slices.Equal(got, []string{other}) {
		t.Fatalf("ResolveAlternates = %q, want %q", got, []string{other})
	}
}

func TestResolveAlternatesCycleTerminates(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	first := filepath.Join(base, "objects")
	second := makeObjectsDir(t, filepath.Join(base, "second"), first)

	makeObjectsDir(t, first, second)

	got := repository.ResolveAlternates(first)
	if !slices.Equal(got, []string{second}) {
		t.Fatalf("ResolveAlternates = %q, want %q", got, []string{second})
	}
}

func TestResolveAlternatesNestingIsBounded(t *testing.T) {
	t.Parallel()

	base := t.TempDir()

	const deep = 8

	dirs := make([]string, deep)
	deeper := []string{}

	for i := deep - 1; i >= 0; i-- {
		dirs[i] = makeObjectsDir(t, filepath.Join(base, "objects"+strconv.Itoa(i)), deeper...)
		deeper = []string{dirs[i]}
	}

	got := repository.ResolveAlternates(dirs[0])

	want := dirs[1:7]
	if !slices.Equal(got, want) {
		t.Fatalf("ResolveAlternates = %q, want %q", got, want)
	}
}

func TestResolveAlternatesNonDirectoryIsSkipped(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	file := filepath.Join(base, "regular")

	err := os.WriteFile(file, []byte("not an object directory"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	dir := makeObjectsDir(t, filepath.Join(base, "objects"), file)

	got := repository.ResolveAlternates(dir)
	if len(got) != 0 {
		t.Fatalf("ResolveAlternates = %q, want none", got)
	}
}

func TestResolveAlternatesStaleIsSkipped(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dir := makeObjectsDir(t, filepath.Join(base, "objects"), filepath.Join(base, "gone"))

	got := repository.ResolveAlternates(dir)
	if len(got) != 0 {
		t.Fatalf("ResolveAlternates = %q, want none", got)
	}
}

func TestAlternatesReadObjects(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			alternate := newRepo(t, objectFormat)

			commitID := makeCommit(t, alternate, "elsewhere")

			addAlternate(t, repo, alternate)

			opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

			objectType, _, err := opened.Objects().ReadBytesContent(commitID)
			if err != nil {
				t.Fatalf("ReadBytesContent: %v", err)
			}

			if objectType != typ.Commit {
				t.Fatalf("type = %v, want %v", objectType, typ.Commit)
			}
		})
	}
}

func TestAlternatesRefusedWhenNotAllowed(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	alternate := newRepo(t, id.ObjectFormatSHA256)

	addAlternate(t, repo, alternate)

	gitDir := openGitDir(t, repo)

	_, err := repository.Open(gitDir, gitDir, repository.Options{})
	if !errors.Is(err, repository.ErrAlternates) {
		t.Fatalf("Open = %v, want ErrAlternates", err)
	}
}

func TestAlternatesSharedClone(t *testing.T) {
	t.Parallel()

	origin := newRepo(t, id.ObjectFormatSHA256)
	commitID := makeCommit(t, origin, "shared")

	err := origin.UpdateRef(t, "refs/heads/main", commitID)
	if err != nil {
		t.Fatalf("UpdateRef: %v", err)
	}

	err = origin.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
	if err != nil {
		t.Fatalf("SymbolicRefUpdate: %v", err)
	}

	clone, err := origin.CloneShared(t, filepath.Join(t.TempDir(), "clone"))
	if err != nil {
		t.Fatalf("CloneShared: %v", err)
	}

	opened := openRepository(t, clone, repository.Options{AllowAlternates: true})

	fetched, err := opened.Fetcher().ExactCommit(commitID)
	if err != nil {
		t.Fatalf("ExactCommit: %v", err)
	}

	if string(fetched.Object().Message) != "shared\n" {
		t.Fatalf("Message = %q, want %q", fetched.Object().Message, "shared\n")
	}
}

func TestAlternatesPacked(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	alternate := newRepo(t, id.ObjectFormatSHA256)

	commitID := makeCommit(t, alternate, "elsewhere")

	err := alternate.UpdateRef(t, "refs/heads/main", commitID)
	if err != nil {
		t.Fatalf("UpdateRef: %v", err)
	}

	err = alternate.Repack(t)
	if err != nil {
		t.Fatalf("Repack: %v", err)
	}

	addAlternate(t, repo, alternate)

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err = opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesRelativeName(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	alternate := newRepo(t, id.ObjectFormatSHA256)

	commitID := makeCommit(t, alternate, "elsewhere")

	relative, err := filepath.Rel(objectsPath(t, repo), objectsPath(t, alternate))
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}

	writeAlternates(t, repo, relative+"\n")

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err = opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesNested(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	middle := newRepo(t, id.ObjectFormatSHA256)
	far := newRepo(t, id.ObjectFormatSHA256)

	commitID := makeCommit(t, far, "far")

	addAlternate(t, middle, far)
	addAlternate(t, repo, middle)

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err := opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesCycle(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	alternate := newRepo(t, id.ObjectFormatSHA256)

	commitID := makeCommit(t, alternate, "elsewhere")

	addAlternate(t, repo, alternate)
	addAlternate(t, alternate, repo)

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err := opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesStale(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	commitID := makeCommit(t, repo, "local")

	writeAlternates(t, repo, filepath.Join(t.TempDir(), "gone")+"\n")

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err := opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesSelf(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	commitID := makeCommit(t, repo, "local")

	addAlternate(t, repo, repo)

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	_, _, err := opened.Objects().ReadBytesContent(commitID)
	if err != nil {
		t.Fatalf("ReadBytesContent: %v", err)
	}
}

func TestAlternatesWritesStayLocal(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	alternate := newRepo(t, id.ObjectFormatSHA256)

	addAlternate(t, repo, alternate)

	opened := openRepository(t, repo, repository.Options{AllowAlternates: true})

	written, err := opened.Objects().WriteBytesContent(typ.Blob, []byte("written"))
	if err != nil {
		t.Fatalf("WriteBytesContent: %v", err)
	}

	openedAlternate := openRepository(t, alternate, repository.Options{})

	_, _, err = openedAlternate.Objects().ReadBytesContent(written)
	if err == nil {
		t.Fatal("the alternate gained an object written to the repository")
	}
}
