# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/lindenii/furgit.svg)](https://pkg.go.dev/codeberg.org/lindenii/furgit)

Furgit is a low-level Git library in Go.

## Status

* Several years away from stable
* Do not use in production
* Mature alternative: [go-git](https://github.com/go-git/go-git)
* Will use [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) starting at 1.0.0

## Goals

* General-purpose Git plumbing library for UNIX-like systems
* Aim for extremely clear and modular architecture
* Then aim for high performance
* Expect familiarity with Git internals

## Finding your way around

If you are working with a normal on-disk repository, start with
`repository.Open(...)`. It opens the repository and wires together the refs
storage, object storage, fetcher, (optional) commit-graph, commit queries, and
reachability helpers. Note that it requires either a bare repository or a
`.git` directory. Then,

* `repo.Refs()` is for branch names, tags, `HEAD`, and ref updates.
  * Use it when you are starting from names rather than object IDs.
  * A common pattern is to resolve a ref first, then pass the resulting object
    ID to the fetcher.

* `repo.Fetcher()` is the main object-facing API for most callers.
  * Use it when you want commits, trees, blobs, or tags as typed values.
  * It also handles peeling through annotated tags, resolving objects to the
    type you actually want, and walking paths inside trees.
  * It even allows you to access a tree as an `io/fs.FS`.
  * If your goal is "show me this commit", "read this tree", "follow this tag",
    or "get me the file at this path", this is usually the right layer.

* `repo.Objects()` is the storage layer underneath `Fetcher`.
  * Use it when you need to read object headers, read raw object contents,
    stream object data, or otherwise look up objects directly by ID.
  * Most callers who want to work with Git objects as commits, trees, blobs, or
    tags should prefer the fetcher instead.
  * However, checking an object ID's size and type are somewhat common
    operations that should be done here.

* `repo.CommitQueries()` is the main graph-query API.
  * Use it when you need ancestor checks or merge-base computation.

* `repo.Reachability()` is the graph-traversal API.
  * Use it when you need to walk reachable commits or objects, or to perform
    connectivity checks.

* `repo.CommitGraph()` exposes the low-level commit-graph reader.
  * Not all repositories have commit graphs, so it may be nil.
  * Most callers should prefer `repo.CommitQueries()` or `repo.Reachability()`
    unless they specifically need direct commit-graph access.

Note that:

* `object` contains parsed Git object values such as blobs, trees, commits, and
  tags. These are the decoded contents of Git objects and do not tell you
  anything about the object's identity.

* `object/stored` wraps a parsed object together with the object ID it was
  loaded from. This is used when you need both the parsed value and the
  identity it was loaded under.

As a rule of thumb:

* If you have a ref name, start with `repo.Refs()`.
* If you want typed objects or path-based access, use `repo.Fetcher()`.
* If you need raw object lookup by ID, object headers, or object streams, use
  `repo.Objects()`.
* If you need ancestor checks or merge bases, use `repo.CommitQueries()`.
* If you need commit or object graph traversal, use `repo.Reachability()`.

Some operations remain available if you want to work below those accessors:

* To accept pushes on the server side, construct `network/receivepack` or
  `network/receivepack/service` with the repository's ref store, object store,
  commit-graph reader, and object ID algorithm.
  * Push handling also needs the repository's object storage root so incoming
    objects can be quarantined and later promoted.
  * `Repository` does not currently expose that root directly (we'll consider
    possible solutions sometime later), so a push server usually keeps the
    repository path or object root handle alongside the `Repository` value.
  * Hook-based checks are just Go functions; for example, a fast-forward check
    can use `commitquery.Queries` over the existing and quarantined object
    stores. Some hooks are provided.

## Alternatives

Not endorsements.

* [github.com/go-git/go-git](https://github.com/go-git/go-git) (by far the most mature)
* [github.com/driusan/dgit](https://github.com/driusan/dgit)
* [github.com/Nivl/git-go](https://github.com/Nivl/git-go)
* [github.com/unkn0wn-root/git-go.git](https://github.com/unkn0wn-root/git-go.git)
* [github.com/speedata/gogit](https://github.com/speedata/gogit)

## Community

* [#lindenii](https://webirc.runxiyu.org/kiwiirc/#lindenii)
  on [irc.runxiyu.org](https://irc.runxiyu.org)
* [#lindenii](https://web.libera.chat/#lindenii)
  on [Libera.Chat](https://libera.chat)

See the CONTRIBUTING document for bug reports and patch submissions.

## Acknowledgements

Partly inspired by [upstream Git](https://git-scm.com),
OpenBSD's [Game of Trees](https://gameoftrees.org), and
[9front Git](https://git.9front.org/plan9front/9front/HEAD/sys/src/cmd/git/f.html).

## License

This project is licensed under the GNU Affero General Public License,
Version 3.0 only.

Pursuant to Section 14 of the GNU Affero General Public License, Version 3.0,
[Runxi Yu](https://runxiyu.org) is hereby designated as the proxy who is
authorized to issue a public statement accepting any future version of the
GNU Affero General Public License for use with this Program.

Therefore, notwithstanding the specification that this Program is licensed
under the GNU Affero General Public License, Version 3.0 only, a public
acceptance by the Designated Proxy of any subsequent version of the GNU Affero
General Public License shall permanently authorize the use of that accepted
version for this Program.
