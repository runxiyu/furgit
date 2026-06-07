// Package tag provides parsed annotated tag objects and tag serialization.
//
// It parses annotated tags into ordinary Go values for reading and construction.
// It does not preserve the exact original byte layout
// needed for signature verification;
// callers that need signature-verification payload fidelity
// should use [lindenii.org/go/furgit/object/signed/tag].
package tag
