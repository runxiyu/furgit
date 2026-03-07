package files

import (
	"fmt"
	"strings"

	"codeberg.org/lindenii/furgit/refstore"
)

func (tx *Transaction) resolveOrdinaryTarget(name string, allowMissing bool) (resolvedWriteTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedWriteTarget{}, fmt.Errorf("refstore/files: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		refState, err := tx.directRead(cur)
		if err != nil {
			return resolvedWriteTarget{}, err
		}

		switch refState.kind {
		case directMissing:
			if !allowMissing {
				return resolvedWriteTarget{}, refstore.ErrReferenceNotFound
			}

			return resolvedWriteTarget{name: cur, loc: tx.store.loosePath(cur), ref: refState}, nil
		case directDetached:
			return resolvedWriteTarget{name: cur, loc: tx.store.loosePath(cur), ref: refState}, nil
		case directSymbolic:
			target := strings.TrimSpace(refState.target)
			if target == "" {
				return resolvedWriteTarget{}, fmt.Errorf("refstore/files: symbolic reference %q has empty target", cur)
			}

			cur = target
		default:
			return resolvedWriteTarget{}, fmt.Errorf("refstore/files: unsupported direct reference state %d", refState.kind)
		}
	}
}
