package loose_test

import (
	"os"
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

// gitOracleObjects seeds the repository with history
// and precomputes every seeded object's
// expected content body and full serialized form.
func gitOracleObjects(t *testing.T, repo *testgit.Repo) []gitOracleObject {
	t.Helper()

	seeded, err := repo.SeedHistory(t)
	if err != nil {
		t.Fatalf("SeedHistory: %v", err)
	}

	groups := []struct {
		name string
		ty   typ.Type
		oids []id.ObjectID
	}{
		{name: "blob", ty: typ.Blob, oids: seeded.Blobs},
		{name: "tree", ty: typ.Tree, oids: seeded.Trees},
		{name: "commit", ty: typ.Commit, oids: seeded.Commits},
		{name: "tag", ty: typ.Tag, oids: seeded.Tags},
	}

	objects := make([]gitOracleObject, 0, len(seeded.All()))

	for _, group := range groups {
		for _, oid := range group.oids {
			body, err := repo.CatFile(t, group.ty, oid)
			if err != nil {
				t.Fatalf("CatFile(%s %s): %v", group.name, oid, err)
			}

			raw := header.Append(nil, group.ty, len(body))
			raw = append(raw, body...)

			objects = append(objects, gitOracleObject{
				name: group.name + " " + oid.String(),
				ty:   group.ty,
				id:   oid,
				body: body,
				raw:  raw,
			})
		}
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
