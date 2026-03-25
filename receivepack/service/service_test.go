package service_test

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store/memory"
	"codeberg.org/lindenii/furgit/receivepack/service"
)

func TestExecutePackExpectedWithoutObjectsRoot(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store := memory.New(algo)
		svc := service.New(service.Options{
			Algorithm:       algo,
			ExistingObjects: store,
		})

		result, err := svc.Execute(context.Background(), &service.Request{
			Commands: []service.Command{{
				Name:  "refs/heads/main",
				OldID: objectid.Zero(algo),
				NewID: objectid.Zero(algo),
			}},
			PackExpected: true,
			Pack:         strings.NewReader("not a pack"),
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		if result.UnpackError != "objects root not configured" {
			t.Fatalf("unexpected unpack error %q", result.UnpackError)
		}
	})
}

func TestExecuteRemovesDerivedQuarantineAfterIngestFailure(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store := memory.New(algo)
		objectsDir := t.TempDir()

		objectsRoot, err := os.OpenRoot(objectsDir)
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}

		t.Cleanup(func() {
			_ = objectsRoot.Close()
		})

		svc := service.New(service.Options{
			Algorithm:       algo,
			ExistingObjects: store,
			ObjectsRoot:     objectsRoot,
		})

		result, err := svc.Execute(context.Background(), &service.Request{
			Commands: []service.Command{{
				Name:  "refs/heads/main",
				OldID: objectid.Zero(algo),
				NewID: objectid.Zero(algo),
			}},
			PackExpected: true,
			Pack:         strings.NewReader("not a pack"),
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		if result.UnpackError == "" {
			t.Fatal("Execute returned empty unpack error for invalid pack")
		}

		entries, err := fs.ReadDir(objectsRoot.FS(), ".")
		if err != nil {
			t.Fatalf("fs.ReadDir: %v", err)
		}

		if len(entries) != 0 {
			t.Fatalf("objects root still has entries after failed ingest: %d", len(entries))
		}
	})
}
