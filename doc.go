// Package furgit provides low-level git operations.
//
// Furgit is an unfinished research project,
// and is not suitable for production.
//
// The architecture is rather deliberate
// and the parts that exist
// have been built carefully
// (to the best of our capability)
// and tested against git as the canonical oracle, etc.
//
// However, the API is not yet settled at all;
// we may revise interfaces and behaviour at any time,
// to incorporate findings from ongoing audits,
// including full rearchitectures.
// Consumers that depend on a 0.x version of furgit
// will need to revise their usages
// every few days or weeks.
//
// Therefore, you should probably not use furgit.
// Consider using [go-git],
// which is mature and has API stability promises,
// or simply use [os/exec] to upstream git,
// which is sufficient for many purposes.
//
// # Overall architecture
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
// When both a type and one of its methods
// specify conflicting labels,
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
// Lifetime labels:
//
//   - Life-Independent: returned values remain valid
//     independently of the parent or provider.
//   - Life-Parent: returned values are only valid
//     while the parent or provider remains valid.
//   - Life-Call: returned values are only valid
//     for the duration of the current call, callback, or hook invocation.
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
//
//   - Mut-No: returned values must not be mutated.
//
// These labels describe the API contract only,
// and do not imply any specific implementation strategy.
//
// Concrete implementations of capability interfaces
// generally inherit the contract documented by the interfaces they satisfy.
// Implementation docs focus on additional guarantees
// and implementation-specific behavior,
// but users are advised not to rely on them
// in order to maintain the flexibility of switching implementations.
//
// [go-git]: https://github.com/go-git/go-git
package furgit
