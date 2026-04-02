// Package furgit provides low-level Git operations.
//
// Furgit provides absolutely no guarantees,
// including but not limited to correctness, performance, API stability, etc.
// In particular, before version 1.0.0,
// no attempt at API stability is made at all:
// breaking changes may even be introduced in patch-level releases.
// See also the warranty and liability disclaimers in the license.
//
// Git libraries often center on a central repository type where all
// data structures and operations are hosted.
// Furgit inverts that:
// objects are plain values that do not provide access to the storer that loaded them,
// object storage and ref storage are sets of narrow interfaces
// that consist only of methods that are reasonable for all implementations,
// and every higher-level operation,
// such as commit traversal, reachability analysis, and recursive peeling,
// is built over those interfaces as independent subsystems.
//
// # Contract labels
//
// Many furgit APIs use short labels to document
// concurrency, dependency ownership, value lifetime, and other behavior.
//
// When both a type and one of its methods specify conflicting labels,
// the method-label takes precedence.
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
//   - Deps-Borrowed: the value borrows supplied dependencies.
//     Usually also labelled with Life-Parent
//     unless those dependencies are not retained past construction.
//   - Deps-Mixed: some are mixed while others are borrowed.
//     The documentation should specify further.
//
// Close labels:
//
//   - Close-Caller: the caller must close the returned value.
//   - Close-No: the caller must not close the returned value,
//     even if the returned value has a Close method.
//
// Idempotency labels:
//
//   - Idem-Yes: action is idempotent.
//   - Idem-2UB: action is not idempotent;
//     repeated use is undefined behavior.
//
// Mutation labels:
//   - Mut-No: returned values must not be mutated.
//
// These labels describe the API contract only,
// and do not imply any specific implementation strategy.
//
// Concrete implementations of capability interfaces
// generally inherit the contract documented by the interfaces they satisfy.
// Implementation docs focus on additional guarantees
// and implementation-specific behavior,
// but users are advised not to reply on them
// in order to maintain the flexibility of switching implementations.
package furgit
