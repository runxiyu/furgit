package repository

import (
	"codeberg.org/lindenii/furgit/ref"
)

// ResolveRef resolves one reference name to symbolic or detached form.
func (repo *Repository) ResolveRef(name string) (ref.Ref, error) {
	return repo.refs.Resolve(name)
}

// ResolveRefFully resolves one reference name to detached form.
func (repo *Repository) ResolveRefFully(name string) (ref.Detached, error) {
	return repo.refs.ResolveFully(name)
}

// ListRefs lists references matching pattern.
func (repo *Repository) ListRefs(pattern string) ([]ref.Ref, error) {
	return repo.refs.List(pattern)
}

// ShortenRef returns the shortest unambiguous shorthand for a full reference name.
func (repo *Repository) ShortenRef(name string) (string, error) {
	return repo.refs.Shorten(name)
}
