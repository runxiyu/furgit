// Package ingest writes one incoming pack stream
// into an objects/pack directory
// as a finalized pack, index, and reverse index.
//
// WritePack streams the pack to a temporary file
// while scanning its entries,
// resolves every delta against in-pack bases,
// optionally completes thin packs from an external base reader,
// and publishes the artifacts under content-addressed names
// derived from the pack trailer hash.
package ingest
