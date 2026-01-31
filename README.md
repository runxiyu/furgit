# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/lindenii/furgit.svg)](https://pkg.go.dev/codeberg.org/lindenii/furgit)

Furgit is a fast Git library in pure Go
(and a little bit of optional Go Assembly).

## Project status

* Initial development
* Frequent breaking changes
* Do not use in production
* Will likely use [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) later

## Current features

* SHA-256 and SHA-1\
  (runtime supports both; tests are SHA-256 by default,
  but the `sha1` build tag makes it test SHA-1)
* Reading loose objects
* Writing loose objects
* Reading packfiles
  * Delta resolution
* Writing packfiles
  * Delta compression
  * Thin packs
* Reading commit-graph's
* General support for blobs, trees, commits, and tags

## Future features

* Compression algorithm agility
* Multi pack indexes
* Repack
* Better delta base selection strategy
* Delta reuse; delta islands
* Reading reachability bitmaps
* Writing reachability bitmaps when writing packfiles
* [commit-graph](https://git-scm.com/docs/commit-graph) (in progress)
* Network protocols
* Reftables
* And much more

## General goals

Furgit intends to be a general-purpose Git library.

For now, Furgit primarily prioritize APIs and optimizations that are
likely to be used by software development forges and other
server-side usages; in particular, Furgit follows the needs of
[Villosa](https://villosa.lindenii.org/villosa//repos/villosa/) and
to some extent [tangled](https://tangled.org/@tangled.org/core).

## Performance optimizations

* Aggressive pooling of byte buffers
* Aggressive pooling of custom zlib readers
* Minor SIMD optimizations for Adler-32
* Memory-mapping packfiles and their indexes

## Performance

See [gitbench](https://git.sr.ht/~runxiyu/gitbench) for details on methods.

All tests below were run on `linux.git` with `HEAD` at `6da43bbeb6918164`
on a `Intel(R) Core(TM) i5-10210U CPU @ 1.60GHz`.

| Task                     | [git](https://git-scm.com) | Furgit | [libgit2](https://libgit2.org) | [go-git](https://github.com/go-git/go-git) |
| -                        | -                          | -      | -                              | -                                          |
| Traversing all trees     | 0.1s                       | 9s     | 19s                            | 122s                                       |
| Traversing the root tree | 4ms                        | 1ms    | 11ms                           | 1800ms                                     |

**Note:** go-git is expected to perform much better after
[storage: filesystem/mmap, Add PackScanner to handle large repos](https://github.com/go-git/go-git/pull/1776).

## Architectural considerations

Furgit heavily relies on memory mappings of packfiles, and assume relatively
predictable fault handling behavior. In distributed systems, we advise *not*
using Furgit on top of distributed network filesystems such as CephFS or NFS;
consider solutions where redundancy and distributions belong *above* the Git
layer, e.g., using an RPC protocol over a set of Git nodes each running Furgit
on local repositories.

## Dependencies

* The standard library
* Some things from `golang.org/x`
* `github.com/cespare/xxhash/v2` (may move in-tree at some point)

Some external code is also introduced and maintained in-tree.

## Environment requirements

We currently do not intend to support flexible storage backends such as
[storers in go-git](https://pkg.go.dev/github.com/go-git/go-git/v5/plumbing/storer);
a standard UNIX-like filesystem with
[syscall.Mmap](https://pkg.go.dev/syscall#Mmap) is expected.

## Repos and mirrors

* [Codeberg](https://codeberg.org/lindenii/furgit) (with the canonical issue tracker)
* [SourceHut mirror](https://git.sr.ht/~runxiyu/furgit)
* [tangled mirror](https://tangled.org/@runxiyu.tngl.sh/furgit)
* [GitHub mirror](https://github.com/runxiyu/furgit)

## Community

* [#lindenii](https://webirc.runxiyu.org/kiwiirc/#lindenii)
  on [irc.runxiyu.org](https://irc.runxiyu.org)
* [#lindenii](https://web.libera.chat/#lindenii)
  on [Libera.Chat](https://libera.chat)

## History and lineage

* I wrote Lindenii Forge
* I wrote [hare-git](https://codeberg.org/lindenii/hare-git)
* I wanted a faster Git library for
  [Lindenii Villosa](https://codeberg.org/lindenii/villosa)
  the next generation of Lindenii Forge
* I translated hare-git and put it into `internal/common/git` in Villosa
* I extracted it out into a general-purpose library, which is what we
  have now
* I was thinking of names and I accidentally typed "git" as "fur" (i.e., left
  shifted one key on my QWERTY keyboard), so, "Furgit"

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

For the purposes of the Developer Certificate of Origin, the "open source
license" refers to the GNU Affero General Public License, Version 3.0, with the
above proxy designation pursuant to Section 14.

All contributors are required to "sign-off" their commits (using `git commit
-s`) to indicate that they have agreed to the [Developer Certificate of
Origin](https://developercertificate.org), reproduced below.

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```
