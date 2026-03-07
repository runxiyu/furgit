package service

import (
	"os"
	"path"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore/memory"
)

func TestPromoteQuarantineAppliesConfiguredPermissions(t *testing.T) {
	t.Parallel()

	objectsDir := t.TempDir()
	objectsRoot, err := os.OpenRoot(objectsDir)
	if err != nil {
		t.Fatalf("os.OpenRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	svc := New(Options{
		Algorithm:       objectid.AlgorithmSHA1,
		ExistingObjects: memory.New(objectid.AlgorithmSHA1),
		ObjectsRoot:     objectsRoot,
		PromotedObjectPermissions: &PromotedObjectPermissions{
			DirMode:  0o751,
			FileMode: 0o640,
		},
	})

	quarantineName, quarantineRoot, err := svc.createQuarantineRoot()
	if err != nil {
		t.Fatalf("createQuarantineRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = quarantineRoot.Close()
		_ = objectsRoot.RemoveAll(quarantineName)
	})

	if err := quarantineRoot.Mkdir("ab", 0o700); err != nil {
		t.Fatalf("Mkdir(ab): %v", err)
	}

	if err := quarantineRoot.WriteFile(path.Join("ab", "cdef"), []byte("payload"), 0o600); err != nil {
		t.Fatalf("WriteFile(quarantine loose): %v", err)
	}

	if err := svc.promoteQuarantine(quarantineName, quarantineRoot); err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}

	dirInfo, err := objectsRoot.Stat("ab")
	if err != nil {
		t.Fatalf("Stat(ab): %v", err)
	}

	if got := dirInfo.Mode().Perm(); got != 0o751 {
		t.Fatalf("dir mode = %o, want 751", got)
	}

	fileInfo, err := objectsRoot.Stat(path.Join("ab", "cdef"))
	if err != nil {
		t.Fatalf("Stat(ab/cdef): %v", err)
	}

	if got := fileInfo.Mode().Perm(); got != 0o640 {
		t.Fatalf("file mode = %o, want 640", got)
	}
}

func TestPromoteQuarantineTreatsExistingLooseObjectAsSuccess(t *testing.T) {
	t.Parallel()

	objectsDir := t.TempDir()
	objectsRoot, err := os.OpenRoot(objectsDir)
	if err != nil {
		t.Fatalf("os.OpenRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	svc := New(Options{
		Algorithm:       objectid.AlgorithmSHA1,
		ExistingObjects: memory.New(objectid.AlgorithmSHA1),
		ObjectsRoot:     objectsRoot,
	})

	quarantineName, quarantineRoot, err := svc.createQuarantineRoot()
	if err != nil {
		t.Fatalf("createQuarantineRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = quarantineRoot.Close()
		_ = objectsRoot.RemoveAll(quarantineName)
	})

	if err := quarantineRoot.Mkdir("ab", 0o755); err != nil {
		t.Fatalf("Mkdir(ab): %v", err)
	}

	if err := objectsRoot.Mkdir("ab", 0o755); err != nil {
		t.Fatalf("Mkdir(dst ab): %v", err)
	}

	const payload = "same object bytes"
	if err := quarantineRoot.WriteFile(path.Join("ab", "cdef"), []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile(quarantine loose): %v", err)
	}

	if err := objectsRoot.WriteFile(path.Join("ab", "cdef"), []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile(permanent loose): %v", err)
	}

	if err := svc.promoteQuarantine(quarantineName, quarantineRoot); err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}
}

func TestPromoteQuarantineRejectsDifferentExistingPackFile(t *testing.T) {
	t.Parallel()

	objectsDir := t.TempDir()
	objectsRoot, err := os.OpenRoot(objectsDir)
	if err != nil {
		t.Fatalf("os.OpenRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	svc := New(Options{
		Algorithm:       objectid.AlgorithmSHA1,
		ExistingObjects: memory.New(objectid.AlgorithmSHA1),
		ObjectsRoot:     objectsRoot,
	})

	quarantineName, quarantineRoot, err := svc.createQuarantineRoot()
	if err != nil {
		t.Fatalf("createQuarantineRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = quarantineRoot.Close()
		_ = objectsRoot.RemoveAll(quarantineName)
	})

	if err := quarantineRoot.Mkdir("pack", 0o755); err != nil {
		t.Fatalf("Mkdir(pack): %v", err)
	}

	if err := objectsRoot.Mkdir("pack", 0o755); err != nil {
		t.Fatalf("Mkdir(dst pack): %v", err)
	}

	if err := quarantineRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte("new bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile(quarantine pack): %v", err)
	}

	if err := objectsRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte("old bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile(permanent pack): %v", err)
	}

	err = svc.promoteQuarantine(quarantineName, quarantineRoot)
	if err == nil {
		t.Fatal("promoteQuarantine unexpectedly succeeded")
	}
}

func TestPromoteQuarantineAcceptsMatchingExistingPackFile(t *testing.T) {
	t.Parallel()

	objectsDir := t.TempDir()
	objectsRoot, err := os.OpenRoot(objectsDir)
	if err != nil {
		t.Fatalf("os.OpenRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	svc := New(Options{
		Algorithm:       objectid.AlgorithmSHA1,
		ExistingObjects: memory.New(objectid.AlgorithmSHA1),
		ObjectsRoot:     objectsRoot,
	})

	quarantineName, quarantineRoot, err := svc.createQuarantineRoot()
	if err != nil {
		t.Fatalf("createQuarantineRoot: %v", err)
	}

	t.Cleanup(func() {
		_ = quarantineRoot.Close()
		_ = objectsRoot.RemoveAll(quarantineName)
	})

	if err := quarantineRoot.Mkdir("pack", 0o755); err != nil {
		t.Fatalf("Mkdir(pack): %v", err)
	}

	if err := objectsRoot.Mkdir("pack", 0o755); err != nil {
		t.Fatalf("Mkdir(dst pack): %v", err)
	}

	const payload = "identical pack bytes"
	if err := quarantineRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile(quarantine pack): %v", err)
	}

	if err := objectsRoot.WriteFile(path.Join("pack", "pack-a.pack"), []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile(permanent pack): %v", err)
	}

	if err := svc.promoteQuarantine(quarantineName, quarantineRoot); err != nil {
		t.Fatalf("promoteQuarantine: %v", err)
	}
}
