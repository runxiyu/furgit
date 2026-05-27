// Package commit provides parsed commit objects and commit serialization.
//
// It parses commits into ordinary Go values for reading and construction.
// It does not preserve the exact original byte layout
// needed for signature verification;
// callers that need signature-verification payload fidelity
// should use [lindenii.org/go/furgit/object/signed/commit].
package commit
