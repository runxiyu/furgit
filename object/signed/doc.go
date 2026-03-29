// Package signed encapsulates raw signed-object processing.
//
// Its subpackages extract verification payloads and embedded signatures from
// raw commit and tag object bodies, without depending on the parsed
// object models in [codeberg.org/lindenii/furgit/object/commit] and
// [codeberg.org/lindenii/furgit/object/tag].
package signed
