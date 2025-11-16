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

// ResolveHEAD reads HEAD and returns the fully qualified
// ref name it points to.
func (repo *Repository) ResolveHEAD() (string, error) {
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
