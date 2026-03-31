// Package fetch loads typed Git objects from object storage and provides
// higher-level object queries.
//
// Fetching is above [objectstore]: it parses stored objects into blobs, trees,
// commits, and tags, exposes object metadata, peels tree-ish or commit-ish
// objects, resolves paths within trees, and can expose one tree as an [io/fs]
// view.
package fetch
