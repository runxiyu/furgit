package loose_test

import (
	"os"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/typ"
)

// gitOracleObject is one object created by git,
// paired with its expected content body and full serialized form.
type gitOracleObject struct {
	name string
	ty   typ.Type
	id   id.ObjectID
	body []byte
	raw  []byte
}

// openLooseStore opens a loose store over the repository's objects directory.
func openLooseStore(t *testing.T, repo *testgit.Repo) *loose.Loose {
	t.Helper()

	repoRoot := repo.Root(t)

	objectsRoot, err := repoRoot.OpenRoot(".git/objects")
	if err != nil {
		_ = repoRoot.Close()

		t.Fatalf("OpenRoot(.git/objects): %v", err)
	}

	_ = repoRoot.Close()

	t.Cleanup(func() { _ = objectsRoot.Close() })

	looseStore, err := loose.New(objectsRoot, repo.ObjectFormat(t))
	if err != nil {
		t.Fatalf("loose.New: %v", err)
	}

	return looseStore
}

// gitOracleObjects builds one object of each kind with git
// and precomputes each one's expected content body and full serialized form.
func gitOracleObjects(t *testing.T, repo *testgit.Repo) []gitOracleObject {
	t.Helper()

	blobID, err := repo.HashObject(t, typ.TypeBlob, strings.NewReader("blob body\n"))
	if err != nil {
		t.Fatalf("HashObject(blob): %v", err)
	}

	treeID, err := repo.MkTree(t, []testgit.TreeEntry{
		{Mode: "100644", Type: typ.TypeBlob, OID: blobID, Name: "file"},
	})
	if err != nil {
		t.Fatalf("MkTree: %v", err)
	}

	commitID, err := repo.CommitTree(t, treeID, testgit.CommitTreeOptions{Message: "subject\n\nbody"})
	if err != nil {
		t.Fatalf("CommitTree: %v", err)
	}

	tagID, err := repo.TagAnnotated(t, "v1", commitID, testgit.TagAnnotatedOptions{Message: "tag message"})
	if err != nil {
		t.Fatalf("TagAnnotated: %v", err)
	}

	kinds := []struct {
		name string
		ty   typ.Type
		id   id.ObjectID
	}{
		{name: "blob", ty: typ.TypeBlob, id: blobID},
		{name: "tree", ty: typ.TypeTree, id: treeID},
		{name: "commit", ty: typ.TypeCommit, id: commitID},
		{name: "tag", ty: typ.TypeTag, id: tagID},
	}

	objects := make([]gitOracleObject, 0, len(kinds))

	for _, kind := range kinds {
		body, err := repo.CatFile(t, kind.ty, kind.id)
		if err != nil {
			t.Fatalf("CatFile(%s): %v", kind.name, err)
		}

		raw := header.Append(nil, kind.ty, uint64(len(body)))
		raw = append(raw, body...)

		objects = append(objects, gitOracleObject{
			name: kind.name,
			ty:   kind.ty,
			id:   kind.id,
			body: body,
			raw:  raw,
		})
	}

	return objects
}

// corruptLooseObjectTrailer flips the final byte of a loose object file,
// damaging the zlib Adler-32 trailer.
func corruptLooseObjectTrailer(t *testing.T, repo *testgit.Repo, objectID id.ObjectID) {
	t.Helper()

	root := repo.Root(t)

	defer func() { _ = root.Close() }()

	hex := objectID.String()
	relPath := ".git/objects/" + hex[:2] + "/" + hex[2:]

	file, err := root.OpenFile(relPath, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("OpenFile(%q): %v", relPath, err)
	}

	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Stat(%q): %v", relPath, err)
	}

	if info.Size() == 0 {
		t.Fatalf("corrupt trailer on empty file %q", relPath)
	}

	last := make([]byte, 1)

	_, err = file.ReadAt(last, info.Size()-1)
	if err != nil {
		t.Fatalf("ReadAt(%q): %v", relPath, err)
	}

	last[0] ^= 0xff

	_, err = file.WriteAt(last, info.Size()-1)
	if err != nil {
		t.Fatalf("WriteAt(%q): %v", relPath, err)
	}
}
