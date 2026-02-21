// Package packed provides read access to packed Git references.
package packed

import (
	"io"
	"path"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
)

// Store reads references from a parsed packed-refs snapshot.
type Store struct {
	byName  map[string]ref.Detached
	ordered []ref.Detached
}

var _ refstore.Store = (*Store)(nil)

// New parses packed-refs content from r using the given object ID algorithm.
func New(r io.Reader, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}
	if r == nil {
		return nil, io.ErrUnexpectedEOF
	}
	byName, ordered, err := parsePackedRefs(r, algo)
	if err != nil {
		return nil, err
	}
	return &Store{
		byName:  byName,
		ordered: ordered,
	}, nil
}

// Resolve resolves a packed reference name to a detached ref.
func (store *Store) Resolve(name string) (ref.Ref, error) {
	detached, ok := store.byName[name]
	if !ok {
		return nil, refstore.ErrReferenceNotFound
	}
	return detached, nil
}

// ResolveFully resolves a packed reference name to a detached ref.
//
// Packed refs are detached-only, so ResolveFully is equivalent to Resolve.
func (store *Store) ResolveFully(name string) (ref.Detached, error) {
	detached, ok := store.byName[name]
	if !ok {
		return ref.Detached{}, refstore.ErrReferenceNotFound
	}
	return detached, nil
}

// List lists packed references matching pattern.
//
// Pattern uses path.Match syntax against full reference names.
// Empty pattern matches all references.
func (store *Store) List(pattern string) ([]ref.Ref, error) {
	matchAll := pattern == ""
	if !matchAll {
		if _, err := path.Match(pattern, "refs/heads/main"); err != nil {
			return nil, err
		}
	}

	refs := make([]ref.Ref, 0, len(store.ordered))
	for _, entry := range store.ordered {
		if !matchAll {
			matched, err := path.Match(pattern, entry.Name())
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}
		refs = append(refs, entry)
	}
	return refs, nil
}

// Shorten returns the shortest unambiguous shorthand for a packed ref name.
func (store *Store) Shorten(name string) (string, error) {
	_, ok := store.byName[name]
	if !ok {
		return "", refstore.ErrReferenceNotFound
	}

	names := make([]string, 0, len(store.ordered))
	for _, entry := range store.ordered {
		names = append(names, entry.Name())
	}
	return refstore.ShortenName(name, names), nil
}

// Close releases resources associated with the backend.
func (store *Store) Close() error {
	return nil
}
