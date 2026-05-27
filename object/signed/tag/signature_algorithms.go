package signedtag

import objectid "lindenii.org/go/furgit/object/id"

// Algorithms returns the algorithms for which the tag carries signatures.
func (tag *Tag) Algorithms() []objectid.Algorithm {
	var algorithms []objectid.Algorithm

	for _, algo := range objectid.SupportedAlgorithms() {
		if _, ok := tag.signatures[algo]; ok {
			algorithms = append(algorithms, algo)
		}
	}

	return algorithms
}
