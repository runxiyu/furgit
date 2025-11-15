# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit?)
[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~runxiyu/furgit.svg)](https://pkg.go.dev/git.sr.ht/~runxiyu/furgit)

Furgit is a fast implementation of Git in pure Go, extracted from
[an internal package of Lindenii Villosa](https://codeberg.org/lindenii/villosa/src/branch/master/villosad/internal/common/git).

Furgit is in initial development, does not have tagged releases yet, and we can
guarantee that the API will break every now and then. Do not use in
production.

We do not focus on command-line utilities; in particular, Furgit does not
intend to replace [upstream git](https://git-scm.com). It is intended to be
used as a library.

## Features

Currently, furgit is very basic; it supports reading and writing loose objects
and reading from packfiles. There is some infrastructure for writing packfiles
in the tests but they need to be refactored.

We intend for repository objects to be freely usable across goroutines, which
may enable long-running applications such as forges to keep a pool of recently
used repos (including their `.idx` and `.pack` cache) for rapid access.

We currently do not intend to support flexible storage backends such as
[storers in go-git](https://pkg.go.dev/github.com/go-git/go-git/v5/plumbing/storer);
a standard UNIX-like filesystem is expected.

## Performance

Furgit is aggressively optimized for performance.
[As of November 2025](https://git.sr.ht/~runxiyu/gitbench),
for the task of `git ls-tree --long HEAD` on large repos such as
[Linux](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git),
it is:

* approximately the same performance as [upstream git](https://git-scm.com),
* approximately 10 times faster than [libgit2](https://libgit2.org), and
* approximately 1000 times faster than [go-git](https://github.com/go-git/go-git).

## Compatibility

* We only aim to support UNIX-like operating systems that have
  [syscall.Mmap](https://pkg.go.dev/syscall#Mmap).
* Currently, this version of Furgit only supports SHA-1 hashes.
  You may edit two lines in `hash.go` to trivially switch to SHA-256.
  The upstream Villosa project only uses SHA-256 hashes.
  We intend to make this library easily interoperable across hash
  formats, but that's not the primary goal for now.

## Development

* [Main SourceHut repository](https://git.sr.ht/~runxiyu/furgit)
  and [public inbox](https://lists.sr.ht/~runxiyu/public-inbox)
  for [patches](https://git-send-email.io/) and discussions
* [GitHub](https://github.com/runxiyu/furgit)
  (issues and PRs are welcome here too)
* [tangled](https://tangled.org/@runxiyu.tngl.sh/furgit)
  (issues and pulls are welcome here too)
* `#chat` on [irc.runxiyu.org](https://irc.runxiyu.org)
  ([web chat](https://webirc.runxiyu.org/kiwiirc/#chat))

## Etymology

I was thinking of names and I accidentally typed "git" as "fur" (i.e., left
shifted one key on my QWERTY keyboard).

## License

Currently, GNU Affero General Public License, version 3 only.

As an exception, `pack_idx.go` (responsible for `.idx` files which index
packfiles) is public domain; I'd be happy to see a port of it to
[go-git](https://github.com/go-git/go-git), although achieving the same
level of performance likely requires memory mapping.
