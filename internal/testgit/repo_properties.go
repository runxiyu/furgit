package testgit

import "codeberg.org/lindenii/furgit/objectid"

// Dir returns the repository directory path.
func (testRepo *TestRepo) Dir() string {
	return testRepo.dir
}

// Algorithm returns the object ID algorithm configured for this repository.
func (testRepo *TestRepo) Algorithm() objectid.Algorithm {
	return testRepo.algo
}
