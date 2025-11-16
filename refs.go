package furgit

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
)

// ResolveRef resolves a fully qualified ref name to its object ID.
func (repo *Repository) ResolveRef(refname string) (Hash, error) {
	id, err := repo.resolveLooseRef(refname)
	if err == nil {
		return id, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Hash{}, err
	}

	return repo.resolvePackedRef(refname)
}

func (repo *Repository) resolveLooseRef(refname string) (Hash, error) {
	data, err := os.ReadFile(repo.repoPath(refname))
	if err != nil {
		if os.IsNotExist(err) {
			return Hash{}, ErrNotFound
		}
		return Hash{}, err
	}
	line := strings.TrimSpace(string(data))
	id, err := repo.ParseHash(line)
	if err != nil {
		return Hash{}, err
	}
	return id, nil
}

func (repo *Repository) resolvePackedRef(refname string) (Hash, error) {
	path := repo.repoPath("packed-refs")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Hash{}, ErrInvalidObject
		}
		return Hash{}, err
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
		if bytes.Equal(name, want) {
			hex := string(line[:sp])
			id, err := repo.ParseHash(hex)
			if err != nil {
				return Hash{}, err
			}
			return id, nil
		}
	}
	scanErr := scanner.Err()
	if scanErr != nil {
		return Hash{}, scanErr
	}
	return Hash{}, ErrInvalidObject
}

// HeadKind represents the kind of HEAD reference.
type HeadKind int

const (
	// The HEAD reference is invalid.
	HeadKindInvalid HeadKind = iota
	// The HEAD reference points to a detached commit hash.
	HeadKindDetached
	// The HEAD reference points to a symbolic ref.
	HeadKindSymbolic
)

// HeadRef represents a HEAD reference.
type HeadRef struct {
	// Kind is the kind of HEAD reference.
	Kind HeadKind
	// When Kind is HeadSymbolic, Ref is the fully qualified ref name.
	Ref string
	// When Kind is HeadDetached, Hash is the commit hash.
	Hash Hash
}

// ResolveHead reads HEAD into a HEAD reference.
func (repo *Repository) ResolveHead() (HeadRef, error) {
	data, err := os.ReadFile(repo.repoPath("HEAD"))
	if err != nil {
		return HeadRef{Kind: HeadKindInvalid}, err
	}
	line := strings.TrimSuffix(string(data), "\n")
	if strings.HasPrefix(line, "ref: ") {
		refname := strings.TrimSpace(line[5:])
		if !strings.HasPrefix(refname, "refs/") {
			return HeadRef{Kind: HeadKindSymbolic, Ref: refname}, ErrInvalidRef
		}
		return HeadRef{Kind: HeadKindSymbolic, Ref: refname}, nil
	}
	id, err := repo.ParseHash(line)
	if err != nil {
		return HeadRef{Kind: HeadKindInvalid}, err
	}
	return HeadRef{Kind: HeadKindDetached, Hash: id}, nil
}
