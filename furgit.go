// Package furgit provides low-level Git operations.
//
// Furgit provides absolutely no guarantees on correctness, performance,
// API stability. In particular, before version 1.0.0, no attempt at
// API stability is made at all, and breaking changes may be introduced
// in patch-level releases. See also the warranty and liability disclaimers
// in the license.
//
// Git libraries often center on a repository type that owns objects, refs,
// worktree state, and configuration behind a single facade. Furgit inverts
// that: objects are plain values, stored objects are separate types that
// associate objects with their object IDs, object storage and ref storage
// are sets of narrow interfaces consisting only of things that are truly
// reasonable for all implementations to satisfy, and every higher-level
// operation, such as commit traversal, reachability analysis, and
// recursive peeling, is built over those interfaces.
//
// While the [codeberg.org/lindenii/furgit/repository] package is where
// most users should begin, it only exists as one convenient composition of
// those pieces for the standard on-disk repository layout. Nothing inside
// furgit should depend on it; extensions to furgit such as alterntaive
// object stores must not depend on it either.
//
// # Contract labels
//
// Many furgit APIs document concurrency, dependency ownership, value lifetime,
// and close behavior using short labels.
// These labels summarize the API contract, but they do not replace the full
// doc comment on a package, type, function, method, constant, or variable.
//
// When both a type and one of its methods specify labels, the method-level
// labels take precedence for that operation.
//
// Concurrency labels:
//
//   - MT-Safe: safe for concurrent use.
//   - MT-Unsafe: not safe for concurrent use without external synchronization.
//
// Dependency labels:
//
//   - Deps-Owned: the receiver takes ownership of all supplied dependencies
//     where ownership is a reasonable concept.
//   - Deps-Borrowed: the value borrows supplied dependencies. Also Life-Parent
//     in most cases, unless those dependencies are not retained past
//     construction.
//   - Deps-Mixed: some supplied dependencies are owned and others are borrowed.
//
// Lifetime labels:
//
//   - Life-Independent: returned values remain valid independently of the
//     parent or provider.
//   - Life-Parent: returned values are only valid while the parent or provider
//     remains valid.
//   - Life-Call: returned values are only valid for the duration of the
//     current call, callback, or hook invocation.
//
// Close labels:
//
//   - Close-Caller: the caller must close the returned value.
//   - Close-No: the caller must not close the returned value directly.
//   - Close-Idem: repeated Close calls are safe.
//
// Mutation labels:
//
//   - Mut-Never: returned values must not be mutated.
//
// Unless Close-Idem is specified, repeated Close calls are undefined behavior.
//
// Unless a doc comment explicitly states otherwise, these labels describe the
// API contract only. They do not imply any specific implementation strategy.
package furgit
