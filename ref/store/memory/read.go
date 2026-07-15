package memory

import (
	"fmt"

	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
)

// Resolve resolves one reference name from the in-memory namespace.
func (memory *Memory) Resolve(name string) (ref.Ref, error) {
	memory.mu.RLock()
	defer memory.mu.RUnlock()

	return publicRef(name, memory.refs[name])
}

// ResolveToDirect resolves symbolic references
// until one direct reference is reached.
func (memory *Memory) ResolveToDirect(name string) (ref.Direct, error) {
	memory.mu.RLock()
	defer memory.mu.RUnlock()

	return memory.resolveToDirectLocked(name)
}

func (memory *Memory) resolveToDirectLocked(name string) (ref.Direct, error) {
	cur := name
	seen := make(map[string]struct{})

	for {
		if _, ok := seen[cur]; ok {
			return ref.Direct{}, fmt.Errorf("%w: at %q", store.ErrSymbolicCycle, cur)
		}

		seen[cur] = struct{}{}

		resolved, err := publicRef(cur, memory.refs[cur])
		if err != nil {
			return ref.Direct{}, err
		}

		switch resolved := resolved.(type) {
		case ref.Direct:
			return resolved, nil
		case ref.Symbolic:
			if resolved.Target == "" {
				return ref.Direct{}, fmt.Errorf(
					"%w: symbolic reference %q has empty target",
					store.ErrInvalidValue, resolved.Name(),
				)
			}

			cur = resolved.Target
		default:
			panic(fmt.Sprintf("ref/store/memory: unsupported reference type %T", resolved))
		}
	}
}

func publicRef(name string, stored storedRef) (ref.Ref, error) {
	switch stored.kind {
	case storedDirect:
		return ref.Direct{
			RefName:   name,
			ID:        stored.id,
			PeelState: stored.peelState,
			PeeledID:  stored.peeledID,
		}, nil
	case storedSymbolic:
		return ref.Symbolic{RefName: name, Target: stored.target}, nil
	case storedMissing:
		return nil, store.ErrReferenceNotFound
	default:
		panic(fmt.Sprintf("ref/store/memory: unsupported stored reference kind %d", stored.kind))
	}
}
