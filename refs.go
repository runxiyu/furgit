package furgit

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"
)

func (repo *Repository) resolveLooseRef(refname string) (Ref, error) {
	data, err := os.ReadFile(repo.repoPath(refname))
	if err != nil {
		if os.IsNotExist(err) {
			return Ref{}, ErrNotFound
		}
		return Ref{}, err
	}
	line := strings.TrimSpace(string(data))

	if strings.HasPrefix(line, "ref: ") {
		target := strings.TrimSpace(line[5:])
		if target == "" {
			return Ref{Kind: RefKindInvalid}, ErrInvalidRef
		}
		return Ref{
			Kind: RefKindSymbolic,
			Ref:  target,
		}, nil
	}

	id, err := repo.ParseHash(line)
	if err != nil {
		return Ref{Kind: RefKindInvalid}, err
	}
	return Ref{
		Kind: RefKindDetached,
		Hash: id,
	}, nil
}

func (repo *Repository) resolvePackedRef(refname string) (Ref, error) {
	// According to git-pack-refs(1), symbolic refs are never
	// stored in packed-refs, so we only need to look for detached
	// refs here.

	path := repo.repoPath("packed-refs")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Ref{}, ErrNotFound
		}
		return Ref{}, err
	}
	defer func() { _ = f.Close() }()

	want := []byte(refname)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()

		if len(line) == 0 || line[0] == '#' || line[0] == '^' {
			continue
		}

		sp := bytes.IndexByte(line, ' ')
		if sp != repo.hashSize*2 {
			continue
		}

		name := line[sp+1:]

		if !bytes.Equal(name, want) {
			continue
		}

		hex := string(line[:sp])
		id, err := repo.ParseHash(hex)
		if err != nil {
			return Ref{Kind: RefKindInvalid}, err
		}

		ref := Ref{
			Kind: RefKindDetached,
			Hash: id,
		}

		if scanner.Scan() {
			next := scanner.Bytes()
			if len(next) > 0 && next[0] == '^' {
				peeledHex := strings.TrimPrefix(string(next), "^")
				peeledHex = strings.TrimSpace(peeledHex)

				peeledID, err := repo.ParseHash(peeledHex)
				if err != nil {
					return Ref{Kind: RefKindInvalid}, err
				}
				ref.Peeled = peeledID
			}
		}

		if scanErr := scanner.Err(); scanErr != nil {
			return Ref{Kind: RefKindInvalid}, scanErr
		}

		return ref, nil
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return Ref{Kind: RefKindInvalid}, scanErr
	}
	return Ref{}, ErrNotFound
}

// RefKind represents the kind of HEAD reference.
type RefKind int

const (
	// The HEAD reference is invalid.
	RefKindInvalid RefKind = iota
	// The HEAD reference points to a detached commit hash.
	RefKindDetached
	// The HEAD reference points to a symbolic ref.
	RefKindSymbolic
)

// Ref represents a reference.
type Ref struct {
	// Kind is the kind of the reference.
	Kind RefKind
	// When Kind is RefKindSymbolic, Ref is the fully qualified ref name.
	// Otherwise the value is undefined.
	Ref string
	// When Kind is RefKindDetached, Hash is the commit hash.
	// Otherwise the value is undefined.
	Hash Hash
	// When Kind is RefKindDetached, and the ref supposedly points to an
	// annotated tag, Peeled is the peeled hash, i.e., the hash of the
	// object that the tag points to.
	Peeled Hash
}

// ResolveRef reads the given fully qualified ref (such as "HEAD" or "refs/heads/main")
// and interprets its contents as either a symbolic ref ("ref: refs/..."), a detached
// hash, or invalid.
// If path is empty, it defaults to "HEAD".
// (While typically only HEAD may be a symbolic reference, others may be as well.)
func (repo *Repository) ResolveRef(path string) (Ref, error) {
	if path == "" {
		path = "HEAD"
	}

	if !strings.HasPrefix(path, "refs/") && !slices.Contains([]string{
		"HEAD", "ORIG_HEAD", "FETCH_HEAD", "MERGE_HEAD",
		"CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD", "BISECT_HEAD",
	}, path) {
		id, err := repo.ParseHash(path)
		if err == nil {
			return Ref{
				Kind: RefKindDetached,
				Hash: id,
			}, nil
		}

		// For now let's keep this to prevent e.g., random users from
		// specifying something crazy like objects/... or ./config.
		// There may be other legal pseudo-refs in the future,
		// but it's probably the best to stay cautious for now.
		return Ref{Kind: RefKindInvalid}, ErrInvalidRef
	}

	loose, err := repo.resolveLooseRef(path)
	if err == nil {
		return loose, nil
	}
	if err != ErrNotFound {
		return Ref{Kind: RefKindInvalid}, err
	}

	packed, err := repo.resolvePackedRef(path)
	if err == nil {
		return packed, nil
	}
	if err != ErrNotFound {
		return Ref{Kind: RefKindInvalid}, err
	}

	return Ref{Kind: RefKindInvalid}, ErrNotFound
}

// ResolveRefFully resolves a ref by recursively following
// symbolic references until it reaches a detached ref.
// Symbolic cycles are detected and reported.
// Tags are not peeled automatically.
func (repo *Repository) ResolveRefFully(path string) (Hash, error) {
	seen := make(map[string]struct{})
	return repo.resolveRefFully(path, seen)
}

func (repo *Repository) resolveRefFully(path string, seen map[string]struct{}) (Hash, error) {
	if _, found := seen[path]; found {
		return Hash{}, fmt.Errorf("symbolic ref cycle involving %q", path)
	}
	seen[path] = struct{}{}

	ref, err := repo.ResolveRef(path)
	if err != nil {
		return Hash{}, err
	}

	switch ref.Kind {
	case RefKindDetached:
		return ref.Hash, nil

	case RefKindSymbolic:
		if ref.Ref == "" {
			return Hash{}, ErrInvalidRef
		}
		return repo.resolveRefFully(ref.Ref, seen)

	default:
		return Hash{}, ErrInvalidRef
	}
}
