// Package reftable provides read access to Git reftable reference storage.
// This store is experimental, has many issues, and should not be used in any serious capacity for now.
package reftable

import (
	"errors"
	"os"
	"sort"
	"strings"
	"sync"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads references from a reftable stack rooted at $GIT_DIR/reftable.
//
// Store does not own root. Callers are responsible for closing root.
type Store struct {
	// root is the reftable directory capability.
	root *os.Root
	// algo is the repository object ID algorithm.
	algo objectid.Algorithm

	// loadOnce ensures tables.list and table handles are loaded once.
	loadOnce sync.Once
	// loadErr stores loading failure from loadOnce.
	loadErr error

	// stateMu guards tables publication and close transitions.
	stateMu sync.RWMutex
	// tables are loaded in oldest-to-newest order from tables.list.
	tables []*tableFile
	// closed reports whether Close has been called.
	closed bool
}

var _ refstore.Store = (*Store)(nil)

// New creates a read-only reftable store rooted at $GIT_DIR/reftable.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}
	return &Store{root: root, algo: algo}, nil
}

// Close releases mapped table resources associated with this store.
func (store *Store) Close() error {
	store.stateMu.Lock()
	if store.closed {
		store.stateMu.Unlock()
		return nil
	}
	store.closed = true
	tables := store.tables
	store.tables = nil
	store.stateMu.Unlock()

	var closeErr error
	for _, table := range tables {
		if table == nil {
			continue
		}
		if err := table.close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

// Resolve resolves a reference name to either a symbolic or detached ref.
func (store *Store) Resolve(name string) (ref.Ref, error) {
	tables, err := store.ensureTables()
	if err != nil {
		return nil, err
	}
	for i := len(tables) - 1; i >= 0; i-- {
		rec, found, err := tables[i].resolveRecord(name)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		if rec.deleted {
			return nil, refstore.ErrReferenceNotFound
		}
		resolved, err := rec.toRef(name)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	}
	return nil, refstore.ErrReferenceNotFound
}

// ResolveFully resolves symbolic references until it reaches a detached value.
//
// ResolveFully resolves symbolic references only. It does not imply peeling
// annotated tag objects.
func (store *Store) ResolveFully(name string) (ref.Detached, error) {
	seen := map[string]struct{}{}
	cur := name
	for {
		if _, exists := seen[cur]; exists {
			return ref.Detached{}, errors.New("refstore/reftable: symbolic reference cycle")
		}
		seen[cur] = struct{}{}
		resolved, err := store.Resolve(cur)
		if err != nil {
			return ref.Detached{}, err
		}
		switch resolved := resolved.(type) {
		case ref.Detached:
			return resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Detached{}, errors.New("refstore/reftable: symbolic reference has empty target")
			}
			cur = resolved.Target
		default:
			return ref.Detached{}, errors.New("refstore/reftable: unsupported reference type")
		}
	}
}

// List returns references matching pattern.
//
// Pattern uses path.Match syntax against full reference names.
// Empty pattern matches all references.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	tables, err := store.ensureTables()
	if err != nil {
		return nil, err
	}
	visible := make(map[string]ref.Ref)
	masked := make(map[string]struct{})

	for i := len(tables) - 1; i >= 0; i-- {
		if err := tables[i].forEachRecord(func(name string, rec recordValue) error {
			if _, done := masked[name]; done {
				return nil
			}
			masked[name] = struct{}{}
			if rec.deleted {
				return nil
			}
			resolved, err := rec.toRef(name)
			if err != nil {
				return err
			}
			visible[name] = resolved
			return nil
		}); err != nil {
			return nil, err
		}
	}

	matchAll := pattern == ""
	if !matchAll {
		if _, err := pathMatch(pattern, "refs/heads/main"); err != nil {
			return nil, err
		}
	}

	names := make([]string, 0, len(visible))
	for name := range visible {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]ref.Ref, 0, len(names))
	for _, name := range names {
		if !matchAll {
			ok, err := pathMatch(pattern, name)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		out = append(out, visible[name])
	}
	return out, nil
}

// Shorten returns the shortest unambiguous shorthand for a full reference name.
func (store *Store) Shorten(name string) (string, error) {
	refs, err := store.List("")
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(refs))
	found := false
	for _, entry := range refs {
		if entry == nil {
			continue
		}
		full := entry.Name()
		names = append(names, full)
		if full == name {
			found = true
		}
	}
	if !found {
		return "", refstore.ErrReferenceNotFound
	}
	return refstore.ShortenName(name, names), nil
}

// ensureTables loads and validates tables listed by tables.list once.
func (store *Store) ensureTables() ([]*tableFile, error) {
	store.loadOnce.Do(func() {
		tables, err := store.loadTables()
		store.stateMu.Lock()
		store.tables = tables
		store.loadErr = err
		store.stateMu.Unlock()
	})

	store.stateMu.RLock()
	defer store.stateMu.RUnlock()
	if store.closed {
		return nil, errors.New("refstore/reftable: store is closed")
	}
	return store.tables, store.loadErr
}

// loadTables reads tables.list and opens all listed tables.
func (store *Store) loadTables() ([]*tableFile, error) {
	listRaw, err := store.root.ReadFile("tables.list")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(listRaw), "\n")
	names := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		if strings.Contains(line, "/") {
			return nil, errors.New("refstore/reftable: invalid table name")
		}
		names = append(names, line)
	}

	out := make([]*tableFile, 0, len(names))
	for _, name := range names {
		table, err := openTableFile(store.root, name, store.algo)
		if err != nil {
			for _, opened := range out {
				_ = opened.close()
			}
			return nil, err
		}
		out = append(out, table)
	}
	return out, nil
}
