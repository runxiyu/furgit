package testgit

import (
	"os"
	"path/filepath"
	"testing"
)

// OpenFile opens one file relative to the repository root.
func (testRepo *TestRepo) OpenFile(tb testing.TB, name string) *os.File {
	tb.Helper()

	root := testRepo.OpenRoot(tb)

	file, err := root.Open(name)
	if err != nil {
		tb.Fatalf("Open(%q): %v", name, err)
	}

	return file
}

// ReadFile reads one file relative to the repository root.
func (testRepo *TestRepo) ReadFile(tb testing.TB, name string) []byte {
	tb.Helper()

	root := testRepo.OpenRoot(tb)

	data, err := root.ReadFile(name)
	if err != nil {
		tb.Fatalf("ReadFile(%q): %v", name, err)
	}

	return data
}

// WriteFile writes one file relative to the repository root.
func (testRepo *TestRepo) WriteFile(tb testing.TB, name string, data []byte, perm os.FileMode) {
	tb.Helper()

	root := testRepo.OpenRoot(tb)

	err := root.WriteFile(name, data, perm)
	if err != nil {
		tb.Fatalf("WriteFile(%q): %v", name, err)
	}
}

// WriteFileAll writes one file relative to the repository root, creating any
// missing parent directories first.
func (testRepo *TestRepo) WriteFileAll(
	tb testing.TB,
	name string,
	data []byte,
	dirPerm os.FileMode,
	filePerm os.FileMode,
) {
	tb.Helper()

	root := testRepo.OpenRoot(tb)

	dir := filepath.Dir(name)
	if dir != "." {
		err := root.MkdirAll(dir, dirPerm)
		if err != nil {
			tb.Fatalf("MkdirAll(%q): %v", dir, err)
		}
	}

	err := root.WriteFile(name, data, filePerm)
	if err != nil {
		tb.Fatalf("WriteFile(%q): %v", name, err)
	}
}

// Remove removes one path relative to the repository root.
func (testRepo *TestRepo) Remove(tb testing.TB, name string) {
	tb.Helper()

	root := testRepo.OpenRoot(tb)

	err := root.Remove(name)
	if err != nil {
		tb.Fatalf("Remove(%q): %v", name, err)
	}
}
