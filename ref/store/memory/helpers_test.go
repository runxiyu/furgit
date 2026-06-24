package memory_test

import (
	"testing"

	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/memory"
)

// resolveDirect resolves name and asserts that it is a direct reference.
//
// Unlike Memory.ResolveToDirect, it does not follow symbolic references.
func resolveDirect(t *testing.T, memory *memory.Memory, name string) ref.Direct {
	t.Helper()

	resolved, err := memory.Resolve(name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}

	direct, ok := resolved.(ref.Direct)
	if !ok {
		t.Fatalf("Resolve(%q) = %T, want ref.Direct", name, resolved)
	}

	return direct
}

// seed runs fn against a fresh transaction and commits it,
// failing the test on any error.
func seed(t *testing.T, memory *memory.Memory, fn func(tx store.Transaction)) {
	t.Helper()

	tx, err := memory.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction: %v", err)
	}

	fn(tx)

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
}
