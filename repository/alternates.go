package repository

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"lindenii.org/go/furgit/internal/format/alternates"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/loose"
	"lindenii.org/go/furgit/object/store/packed"
)

// ErrAlternates indicates that a repository names alternate object directories
// while [Open] was not allowed to read them.
var ErrAlternates = errors.New("repository: alternate object directories not allowed")

// maxAlternateDepth bounds how deeply alternates may name further alternates.
const maxAlternateDepth = 5

// alternate is one alternate object directory, opened for reading.
//
// An alternate need not hold packs,
// in which case packRoot and packed are nil.
type alternate struct {
	root     *os.Root
	packRoot *os.Root
	loose    *loose.Loose
	packed   *packed.Packed
}

// openAlternates opens the alternates that an objects directory names,
// which [Open] must have been allowed to read.
func openAlternates(
	objectsRoot *os.Root,
	objectFormat id.ObjectFormat,
	options Options,
) ([]*alternate, error) {
	named, err := namesAlternates(objectsRoot)
	if err != nil {
		return nil, err
	}

	if !named {
		return nil, nil
	}

	if !options.AllowAlternates {
		return nil, ErrAlternates
	}

	paths := resolveAlternates(objectsRoot.Name())
	opened := make([]*alternate, 0, len(paths))

	for _, path := range paths {
		alt, err := openAlternate(path, objectFormat)
		if err != nil {
			for _, previous := range opened {
				_ = previous.close()
			}

			return nil, err
		}

		opened = append(opened, alt)
	}

	return opened, nil
}

// namesAlternates reports whether an objects directory names alternates.
func namesAlternates(objectsRoot *os.Root) (bool, error) {
	_, err := objectsRoot.Stat("info/alternates")
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("repository: stat objects/info/alternates: %w", err)
	}

	return true, nil
}

// resolveAlternates returns the alternate object directories
// reachable from objectsPath,
// in the order they are to be consulted.
//
// A name that no longer resolves to a directory is skipped,
// since an alternate may disappear,
// as are repeats and the objects directory itself.
func resolveAlternates(objectsPath string) []string {
	primary, err := filepath.EvalSymlinks(objectsPath)
	if err != nil {
		primary = filepath.Clean(objectsPath)
	}

	seen := map[string]struct{}{primary: {}}
	paths := []string{}

	collectAlternates(primary, 0, seen, &paths)

	return paths
}

// collectAlternates appends the alternates named by objectsPath,
// and then those they name in turn.
func collectAlternates(objectsPath string, depth int, seen map[string]struct{}, paths *[]string) {
	data, err := os.ReadFile(filepath.Join(objectsPath, "info", "alternates")) //#nosec G304 G703
	if err != nil {
		return
	}

	for _, named := range alternates.Parse(data, objectsPath) {
		path, err := filepath.EvalSymlinks(named)
		if err != nil {
			continue
		}

		_, repeated := seen[path]
		if repeated {
			continue
		}

		info, err := os.Stat(path) //#nosec G703
		if err != nil || !info.IsDir() {
			continue
		}

		seen[path] = struct{}{}
		*paths = append(*paths, path)

		if depth+1 > maxAlternateDepth {
			continue
		}

		collectAlternates(path, depth+1, seen, paths)
	}
}

// openAlternate opens one alternate object directory for reading.
func openAlternate(path string, objectFormat id.ObjectFormat) (*alternate, error) {
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, fmt.Errorf("repository: open alternate: %w", err)
	}

	looseStore, err := loose.New(root, objectFormat)
	if err != nil {
		_ = root.Close()

		return nil, fmt.Errorf("repository: open alternate loose objects: %w", err)
	}

	// An alternate is not ours to write to,
	// so a missing pack directory is left alone rather than created.
	packRoot, err := root.OpenRoot("pack")
	if errors.Is(err, fs.ErrNotExist) {
		return &alternate{root: root, packRoot: nil, loose: looseStore, packed: nil}, nil
	}

	if err != nil {
		_ = looseStore.Close()
		_ = root.Close()

		return nil, fmt.Errorf("repository: open alternate pack: %w", err)
	}

	packedStore, err := packed.New(packRoot, objectFormat)
	if err != nil {
		_ = packRoot.Close()
		_ = looseStore.Close()
		_ = root.Close()

		return nil, fmt.Errorf("repository: open alternate packed objects: %w", err)
	}

	return &alternate{
		root:     root,
		packRoot: packRoot,
		loose:    looseStore,
		packed:   packedStore,
	}, nil
}

// readers returns the object readers of one alternate,
// in the order they are to be consulted.
func (alt *alternate) readers() []store.ObjectReader {
	if alt.packed == nil {
		return []store.ObjectReader{alt.loose}
	}

	return []store.ObjectReader{alt.loose, alt.packed}
}

// close releases the stores and roots of one alternate.
func (alt *alternate) close() error {
	errs := []error{alt.loose.Close()}

	if alt.packed != nil {
		errs = append(errs, alt.packed.Close())
	}

	if alt.packRoot != nil {
		errs = append(errs, alt.packRoot.Close())
	}

	errs = append(errs, alt.root.Close())

	return errors.Join(errs...)
}
