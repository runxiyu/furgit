package reftable

import "path"

// pathMatch applies path.Match to full ref names.
func pathMatch(pattern, name string) (bool, error) {
	return path.Match(pattern, name)
}
