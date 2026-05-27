package signedcommit

import objectid "lindenii.org/go/furgit/object/id"

// Algorithms returns the algorithms for which the commit carries signatures.
func (commit *Commit) Algorithms() []objectid.Algorithm {
	var algorithms []objectid.Algorithm

	for _, algo := range objectid.SupportedAlgorithms() {
		if _, ok := commit.signatures[algo]; ok {
			algorithms = append(algorithms, algo)
		}
	}

	return algorithms
}
