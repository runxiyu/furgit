# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit?)
[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~runxiyu/furgit.svg)](https://pkg.go.dev/git.sr.ht/~runxiyu/furgit)

Furgit is a fast implementation of Git in pure Go, extracted from
[an internal package of Lindenii Villosa](https://codeberg.org/lindenii/villosa/src/branch/master/villosad/internal/common/git).

Furgit is in initial development, does not have tagged releases yet, and
guarantees that the API will break every now and then. Do not use in
production.

## Features

Currently, Furgit is very basic; it supports reading objects from loose objects
and packfiles. There is some infrastructure for writing loose objects and
packfiles in the tests but they need to be refactored.

## Performance

Furgit is aggressively optimized for performance.
[As of November 2025](https://git.sr.ht/~runxiyu/gitbench),
for large repos such as
[Linux](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git),
it is:

* close to the performance of [upstream git](https://git-scm.com),
* faster than [libgit2](https://libgit2.org), and
* faster than [go-git](https://github.com/go-git/go-git) by 3 orders of magnitude.

## Compatibility

* We only aim to support UNIX-like operating systems that have
  [syscall.Mmap](https://pkg.go.dev/syscall#Mmap).
* Currently, this version of Furgit only supports SHA-1 hashes. However, the
  upstream Villosa project only uses SHA-256 hashes. You may edit two lines in
  `hash.go` to trivially switch to SHA-256.

## Development

* [Main SourceHut repository](https://git.sr.ht/~runxiyu/furgit)
  and [public inbox](https://lists.sr.ht/~runxiyu/public-inbox)
  for [patches](https://git-send-email.io/) and discussions
* [GitHub mirror](https://github.com/runxiyu/furgit)
  (issues and PRs are welcome here too)

## Etymology

I was thinking of names and I accidentally typed "git" as "fur" (i.e., left
shifted one key on my QWERTY keyboard).

## License

Currently, GNU Affero General Public License, version 3 only.

As an exception, `pack_idx.go` (responsible for `.idx` files which index
packfiles) is public domain; I'd be happy to see a port of it to
[go-git](https://github.com/go-git/go-git), although achieving the same
level of performance likely requires memory mapping.
