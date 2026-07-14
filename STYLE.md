# Style guide

These only really document things that the linter won't check.

## Symbol naming

Try to avoid package aliases.
`object/store` should simply be imported implicity as `store`,
`object/id` as `id`, `object/typ` as `typ`, etc.
In cases like `ref/name`
where `name` really does conflict with variable names a lot,
`refname` may be considered.

Avoid letting local variables shadow an imported package.

Name things like `loose.Loose` rather than `loose.Store`.

## File organization

Each package should have a `doc.go`.

Usually also a `<pkg>.go` with
the package's central type, `New`, `Close`, etc.

Options structs should be close to `New`.
There should not be an `options.go`.

## Sizes and offsets

In-memory sizes and offsets are `int`.
They are bounded by `len()`,
and anything too large for `int` cannot be held in memory anyway.

On-disk and wire-format fields stay `uint64`,
faithful to their serialized representation.
Convert at the boundary with `intconv`.

Convert a value where it is stored or crosses into a domain,
not for a transient one:
a short-lived result can keep its producer's type,
widening the other operand at a comparison
rather than being narrowed to match.

## Comments

Use semantic line breaks,
breaking at sentence ends and major clause boundaries.

## Contract labels

See the `doc.go` at the root of the `furgit` package.

## Tests

Test file base-names are `_test`
appended to the base-name of the file
whose functionality is being tested.

`roundtrip_test.go` is used
for roundtrip tests with ourselves.

`helpers_test.go` is used
for helpers shared between many tests.

`testgit` mainly wraps plumbing commands from git.
Non-git commands may be added in the future.

For things that may plausibly depend on object formats:

```go
for _, objectFormat := range id.SupportedObjectFormats() {
	t.Run(objectFormat.String(), func(t *testing.T) {
		// Stuff.
	})
}
```

## Errors

See the `errs` package for some guidance on globally common errors.

Define sentinel errors
for genuinely distinct and reasonably matchable conditions.

Error stings should begin
with our current package path starting at the module root,
such as `object/store/packed: whatever`.

Non-sentinel error types should only be defined for cases where
it is reasonable to expect callers to act on the information.

Underlying errors should almost always be wrapped and passed along.
For cases where our own sentinel error is reasonable to match on, use
`fmt.Errorf("%w: %w", ErrOurSentinal ,err)`.
Otherwise, use
`fmt.Errorf("path/to/our/package: optionally some context: %w", err)`.


