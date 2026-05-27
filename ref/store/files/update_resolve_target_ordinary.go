package files

import (
	"fmt"
	"strings"

	refstore "lindenii.org/go/furgit/ref/store"
)

func (executor *refUpdateExecutor) resolveOrdinaryTarget(name string, allowMissing bool) (resolvedUpdateTarget, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return resolvedUpdateTarget{}, fmt.Errorf("refstore/files: symbolic reference cycle at %q", cur)
		}

		seen[cur] = struct{}{}

		refState, err := executor.directRead(cur)
		if err != nil {
			return resolvedUpdateTarget{}, err
		}

		switch refState.kind {
		case directMissing:
			if !allowMissing {
				return resolvedUpdateTarget{}, wrapUpdateError(name, refstore.ErrReferenceNotFound)
			}

			return resolvedUpdateTarget{name: cur, loc: executor.store.loosePath(cur), ref: refState}, nil
		case directDetached:
			return resolvedUpdateTarget{name: cur, loc: executor.store.loosePath(cur), ref: refState}, nil
		case directSymbolic:
			target := strings.TrimSpace(refState.target)
			if target == "" {
				return resolvedUpdateTarget{}, wrapUpdateError(name, &refstore.InvalidValueError{
					Err: fmt.Errorf("symbolic reference has empty target"),
				})
			}

			cur = target
		default:
			return resolvedUpdateTarget{}, fmt.Errorf("refstore/files: unsupported direct reference state %d", refState.kind)
		}
	}
}
