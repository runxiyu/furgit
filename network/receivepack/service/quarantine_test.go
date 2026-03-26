package service //nolint:testpackage

// because we need access to quarantine internals

import (
	"os"
	"path"
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/storer/memory"
)

type quarantineFixture struct {
	svc            *Service
	objectsRoot    *os.Root
	quarantineName string
	quarantineRoot *os.Root
}

func newQuarantineFixture(tb testing.TB, opts Options) *quarantineFixture {
	tb.Helper()

	objectsRoot, err := os.OpenRoot(tb.TempDir())
	if err != nil {
		tb.Fatalf("os.OpenRoot: %v", err)
	}

	tb.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	opts.Algorithm = objectid.AlgorithmSHA1
	opts.ExistingObjects = memory.New(objectid.AlgorithmSHA1)
	opts.ObjectsRoot = objectsRoot

	svc := New(opts)

	quarantineName, quarantineRoot, err := svc.createQuarantineRoot()
	if err != nil {
		tb.Fatalf("createQuarantineRoot: %v", err)
	}

	tb.Cleanup(func() {
		_ = quarantineRoot.Close()
		_ = objectsRoot.RemoveAll(quarantineName)
	})

	return &quarantineFixture{
		svc:            svc,
		objectsRoot:    objectsRoot,
		quarantineName: quarantineName,
		quarantineRoot: quarantineRoot,
	}
}

func writeMatchingPromotedFile(
	tb testing.TB,
	quarantineRoot, objectsRoot *os.Root,
	dir, name, payload string,
) {
	tb.Helper()

	err := quarantineRoot.Mkdir(dir, 0o755)
	if err != nil {
		tb.Fatalf("Mkdir(%s): %v", dir, err)
	}

	err = objectsRoot.Mkdir(dir, 0o755)
	if err != nil {
		tb.Fatalf("Mkdir(dst %s): %v", dir, err)
	}

	rel := path.Join(dir, name)

	err = quarantineRoot.WriteFile(rel, []byte(payload), 0o644)
	if err != nil {
		tb.Fatalf("WriteFile(quarantine %s): %v", rel, err)
	}

	err = objectsRoot.WriteFile(rel, []byte(payload), 0o644)
	if err != nil {
		tb.Fatalf("WriteFile(permanent %s): %v", rel, err)
	}
}

func TestPromoteQuarantineAppliesConfiguredPermissions(t *testing.T) {
	t.Parallel()

	fx := newQuarantineFixture(t, Options{
		PromotedObjectPermissions: &PromotedObjectPermissions{
			DirMode:  0o751,
			FileMode: 0o640,
		},
	})

	err := fx.quarantineRoot.Mkdir("ab", 0o700)
	if err != nil {
		t.Fatalf("Mkdir(ab): %v", err)
	}

	err = fx.quarantineRoot.WriteFile(path.Join("ab", "cdef"), []byte("payload"), 0o600)
	if err != nil {
		t.Fatalf("WriteFile(quarantine loose): %v", err)
	}

	err = fx.svc.promoteQuarantine(fx.quarantineName, fx.quarantineRoot)
	if err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}

	dirInfo, err := fx.objectsRoot.Stat("ab")
	if err != nil {
		t.Fatalf("Stat(ab): %v", err)
	}

	if got := dirInfo.Mode().Perm(); got != 0o751 {
		t.Fatalf("dir mode = %o, want 751", got)
	}

	fileInfo, err := fx.objectsRoot.Stat(path.Join("ab", "cdef"))
	if err != nil {
		t.Fatalf("Stat(ab/cdef): %v", err)
	}

	if got := fileInfo.Mode().Perm(); got != 0o640 {
		t.Fatalf("file mode = %o, want 640", got)
	}
}

func TestPromoteQuarantineTreatsExistingLooseObjectAsSuccess(t *testing.T) {
	t.Parallel()

	fx := newQuarantineFixture(t, Options{})
	writeMatchingPromotedFile(t, fx.quarantineRoot, fx.objectsRoot, "ab", "cdef", "same object bytes")

	err := fx.svc.promoteQuarantine(fx.quarantineName, fx.quarantineRoot)
	if err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}
}

func TestPromoteQuarantineRejectsDifferentExistingPackFile(t *testing.T) {
	t.Parallel()

	fx := newQuarantineFixture(t, Options{})

	err := fx.quarantineRoot.Mkdir("pack", 0o755)
	if err != nil {
		t.Fatalf("Mkdir(pack): %v", err)
	}

	err = fx.objectsRoot.Mkdir("pack", 0o755)
	if err != nil {
		t.Fatalf("Mkdir(dst pack): %v", err)
	}

	err = fx.quarantineRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte("new bytes"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile(quarantine pack): %v", err)
	}

	err = fx.objectsRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte("old bytes"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile(permanent pack): %v", err)
	}

	err = fx.svc.promoteQuarantine(fx.quarantineName, fx.quarantineRoot)
	if err == nil {
		t.Fatal("promoteQuarantine unexpectedly succeeded")
	}
}

func TestPromoteQuarantineAcceptsMatchingExistingPackFile(t *testing.T) {
	t.Parallel()

	fx := newQuarantineFixture(t, Options{})
	writeMatchingPromotedFile(t, fx.quarantineRoot, fx.objectsRoot, "pack", "pack-a.pack", "identical pack bytes")

	err := fx.svc.promoteQuarantine(fx.quarantineName, fx.quarantineRoot)
	if err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}
}
