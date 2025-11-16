package furgit

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"strings"
)

// ResolveRef resolves a fully qualified ref name to its object ID.
func (repo *Repository[T]) ResolveRef(refname string) (Hash[T], error) {
	id, err := repo.resolveLooseRef(refname)
	if err == nil {
		return id, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Hash[T]{}, err
	}

	return repo.resolvePackedRef(refname)
}

func (repo *Repository[T]) resolveLooseRef(refname string) (Hash[T], error) {
	data, err := os.ReadFile(repo.repoPath(refname))
	if err != nil {
		if os.IsNotExist(err) {
			return Hash[T]{}, ErrNotFound
		}
		return Hash[T]{}, err
	}
	line := strings.TrimSpace(string(data))
	id, err := ParseHash[T](line)
	if err != nil {
		return Hash[T]{}, err
	}
	return id, nil
}

func (repo *Repository[T]) resolvePackedRef(refname string) (Hash[T], error) {
	path := repo.repoPath("packed-refs")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Hash[T]{}, ErrInvalidObject
		}
		return Hash[T]{}, err
	}
	defer func() { _ = f.Close() }()

	hashSize := repo.hashSize()
	want := []byte(refname)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 || line[0] == '#' || line[0] == '^' {
			continue
		}
		sp := bytes.IndexByte(line, ' ')
		if sp != hashSize*2 {
			continue
		}
		name := line[sp+1:]
		if bytes.Equal(name, want) {
			hex := string(line[:sp])
			id, err := ParseHash[T](hex)
			if err != nil {
				return Hash[T]{}, err
			}
			return id, nil
		}
	}
	scanErr := scanner.Err()
	if scanErr != nil {
		return Hash[T]{}, scanErr
	}
	return Hash[T]{}, ErrInvalidObject
}

// ResolveHEAD reads HEAD and returns the ref that HEAD points to.
func (repo *Repository[T]) ResolveHEAD() (string, error) {
	data, err := os.ReadFile(repo.repoPath("HEAD"))
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(data))
	const prefix = "ref: "
	if strings.HasPrefix(line, prefix) {
		ref := strings.TrimSpace(line[len(prefix):])
		if ref == "" {
			return "", ErrInvalidRef
		}
		return ref, nil
	}
	return "", ErrInvalidRef
}
