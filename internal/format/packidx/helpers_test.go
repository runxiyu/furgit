package packidx_test

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// makeGitPack builds a repository with some blobs,
// packs them with git pack-objects,
// and returns the repository, the artifact path prefix,
// and the packed object IDs.
func makeGitPack(t *testing.T, objectFormat id.ObjectFormat) (*testgit.Repo, string, []id.ObjectID) {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	oids := make([]id.ObjectID, 0, 40)

	for i := range 40 {
		content := fmt.Sprintf("blob %d\n%s", i, strings.Repeat("filler\n", i))

		oid, err := repo.HashObject(t, typ.Blob, strings.NewReader(content))
		if err != nil {
			t.Fatalf("HashObject: %v", err)
		}

		oids = append(oids, oid)
	}

	prefix, err := repo.PackObjects(t, slices.Values(oids), testgit.PackObjectsOptions{})
	if err != nil {
		t.Fatalf("PackObjects: %v", err)
	}

	return repo, prefix, oids
}

// parseGitIdxFile reads and parses one .idx file produced by git.
func parseGitIdxFile(t *testing.T, prefix string, objectFormat id.ObjectFormat) ([]byte, packidx.Packidx) {
	t.Helper()

	data, err := os.ReadFile(prefix + ".idx") //nolint:gosec
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	idx, err := packidx.Parse(data, objectFormat.Size())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	return data, idx
}

// gitPackOffsets extracts the object-to-offset mapping
// from git verify-pack -v output.
func gitPackOffsets(t *testing.T, repo *testgit.Repo, idxPath string, objectFormat id.ObjectFormat) map[string]uint64 {
	t.Helper()

	out, err := repo.VerifyPack(t, idxPath)
	if err != nil {
		t.Fatalf("VerifyPack: %v", err)
	}

	hexLen := objectFormat.HexLen()
	offsets := make(map[string]uint64)

	for line := range strings.Lines(string(out)) {
		fields := strings.Fields(line)
		if len(fields) < 5 || len(fields[0]) != hexLen {
			continue
		}

		offset, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			continue
		}

		offsets[fields[0]] = offset
	}

	return offsets
}
