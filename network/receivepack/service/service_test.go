package service_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/network/receivepack/service"
	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objectdual "lindenii.org/go/furgit/object/store/dual"
	objectloose "lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/store/memory"
	objectpacked "lindenii.org/go/furgit/object/store/packed"
)

func TestExecutePackExpectedWithoutObjectIngress(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store := memory.New(algo)
		svc := service.New(service.Options{
			ExistingObjects: store,
		})

		result, err := svc.Execute(context.Background(), &service.Request{
			Commands: []service.Command{{
				Name:  "refs/heads/main",
				OldID: algo.Zero(),
				NewID: algo.Zero(),
			}},
			PackExpected: true,
			Pack:         strings.NewReader("not a pack"),
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		if result.UnpackError != "object ingress not configured" {
			t.Fatalf("unexpected unpack error %q", result.UnpackError)
		}
	})
}

func TestExecuteDiscardedQuarantineAfterIngestFailure(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store := memory.New(algo)
		objectIngress := newDualIngress(t, algo)

		svc := service.New(service.Options{
			ExistingObjects: store,
			ObjectIngress:   objectIngress,
		})

		result, err := svc.Execute(context.Background(), &service.Request{
			Commands: []service.Command{{
				Name:  "refs/heads/main",
				OldID: algo.Zero(),
				NewID: algo.Zero(),
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
	})
}

func newDualIngress(tb testing.TB, algo objectid.Algorithm) objectstore.Quarantiner {
	tb.Helper()

	objectsRoot, err := os.OpenRoot(tb.TempDir())
	if err != nil {
		tb.Fatalf("os.OpenRoot: %v", err)
	}

	tb.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	err = objectsRoot.Mkdir("pack", 0o755)
	if err != nil {
		tb.Fatalf("Mkdir(pack): %v", err)
	}

	packRoot, err := objectsRoot.OpenRoot("pack")
	if err != nil {
		tb.Fatalf("OpenRoot(pack): %v", err)
	}

	tb.Cleanup(func() {
		_ = packRoot.Close()
	})

	looseStore, err := objectloose.New(objectsRoot, algo)
	if err != nil {
		tb.Fatalf("loose.New: %v", err)
	}

	tb.Cleanup(func() {
		_ = looseStore.Close()
	})

	packedStore, err := objectpacked.New(packRoot, algo, objectpacked.Options{WriteRev: true})
	if err != nil {
		tb.Fatalf("packed.New: %v", err)
	}

	tb.Cleanup(func() {
		_ = packedStore.Close()
	})

	return objectdual.New(looseStore, packedStore)
}
