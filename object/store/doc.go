// Package objectstore provides interfaces for object storage backends.
//
// Reading stores only respond to object-ID queries in terms of headers (type
// and size), raw bytes, and streaming payloads, but they do not parse commits,
// trees, blobs, or tags into typed values. Turning stored objects into typed
// objects is the job of [codeberg.org/lindenii/furgit/object/fetch].
//
// This package does not define one unified writing interface. Backends have
// very different write models: writing one loose object is natural, while
// writing one object into a packfile backend is wasteful. Instead, we define
// distinct optional capabilities for object-wise writes, pack-wise writes,
// and compose them against quarantined writes.
//
// Concrete implementations generally inherit the contract documented by the
// interfaces they satisfy. Implementation docs focus on additional guarantees
// and implementation-specific behavior.
package objectstore
