package objectid

// SupportedAlgorithms returns all object ID algorithms supported by furgit.
// Do not mutate.
func SupportedAlgorithms() []Algorithm {
	return supportedAlgorithms
}
