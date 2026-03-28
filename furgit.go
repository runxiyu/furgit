// Package furgit provides low-level Git operations.
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
//   - MT-Safe: safe for concurrent use, excluding Close unless documented
//     otherwise.
//   - MT-ReadSafe: safe for concurrent read-only use.
//   - MT-Unsafe: not safe for concurrent use without external synchronization.
//
// Dependency labels:
//
//   - Deps-Owned: the receiver takes ownership of all supplied dependencies
//     where ownership is a reasonable concept.
//   - Deps-Borrowed: the receiver borrows supplied dependencies.
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
//   - Close-UB: repeated Close calls are undefined behavior.
//
// Unless a doc comment explicitly states otherwise, these labels describe the
// API contract only. They do not imply any specific implementation strategy.
package furgit
