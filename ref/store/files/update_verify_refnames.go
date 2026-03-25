package files

import (
	"strings"

	"codeberg.org/lindenii/furgit/ref/store"
)

func verifyRefnameAvailable(name string, existing map[string]struct{}, writes []string, deleted map[string]struct{}) error {
	for existingName := range existing {
		if existingName == name {
			continue
		}

		if _, skip := deleted[existingName]; skip {
			continue
		}

		if refnamesConflict(name, existingName) {
			return wrapUpdateError(name, &refstore.NameConflictError{Other: existingName})
		}
	}

	for _, other := range writes {
		if other == name {
			continue
		}

		if refnamesConflict(name, other) {
			return wrapUpdateError(name, &refstore.NameConflictError{Other: other})
		}
	}

	return nil
}

func refnamesConflict(left, right string) bool {
	return left == right ||
		strings.HasPrefix(left, right+"/") ||
		strings.HasPrefix(right, left+"/")
}
