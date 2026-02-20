package testgit

import "codeberg.org/lindenii/furgit/objectid"

// Dir returns the repository directory path.
func (repo *TestRepo) Dir() string {
	return repo.dir
}

// Algorithm returns the object ID algorithm configured for this repository.
func (repo *TestRepo) Algorithm() objectid.Algorithm {
	return repo.algo
}
