# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/lindenii/furgit.svg)](https://pkg.go.dev/codeberg.org/lindenii/furgit)

Furgit is a low-level Git library in Go.

## Project status

* Initial development
* Frequent breaking changes
* Do not use in production
* Several years away from stable
* Will use [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) starting at 1.0.0

## General goals

* General-purpose Git library for UNIX-like systems
* Prioritize APIs for forges and other server-side uses first
* Aim for clear architecture then high performance

## Implemented

* Parsing configs
* Object ID and hash algorithms (SHA-256, SHA-1)
* Object type enums
* Object representation types
* Object header parsing
* Parsing objects
* Serializing objects
* Diffing lines
* Diffing trees
* Object storer interface
* Reading loose objects
* Applying deltas
* `.idx` lookup
* Reading packed objects
* Object storer chain and mixer
* Ref types
* Ref storer interface
* Reading loose refs
* Reading packed refs
* Ref storer chain
* Reachability iterators
* Repository abstractions
* Adler-32 optimizations
* ZLIB pooling
* Streaming `.pack`/`.idx` hash verifier
* `.idx` and `.rev` writing
* Pack ingestion
* Un-thinning thin packs

## Planned

* Quarantine areas
* V0/V1 protocol (e.g., `sideband-64k`, `pktline`, etc.)
* Receiving/fetching
* Writing configs
* Commit graph
* Multi-pack indexes
* Reachability bitmaps
* Delta base selection, e.g., islands
* Compressing deltas
* Writing packfiles
* Thin packs
* DEFLATE optimizations
* Aggressive buffer pooling
* Ref tables
* Large object promisors
* Repacking
* Compression agility
* Working trees
* Index
* Merge
* Checkout
* Rebase

## Benchmarks

* See [gitbench](https://git.sr.ht/~runxiyu/gitbench).
* `legacy` branch furgit is slightly faster due to buffer reuse and custom
  ZLIB. These will be re-added.
* Alpine edge, i5-10210U, `performance` governor, `linux.git`.
* go-git expects significant speed-ups after [mmap](https://pkg.go.dev/github.com/go-git/go-git/v6/storage/filesystem/mmap)
* These lone tests do not represent all workloads. Test your usage
  pattern yourself (and contribute to gitbench).

### Traversing all trees in `HEAD` and fetching each file size

Mainly tests the packfile object reader.

| Implementation | Total  | User   | System |
| -              | -      | -      | -      |
| Git            | 337 ms | 226 ms | 108 ms |
| libgit2        | 391 ms | 269 ms | 120 ms |
| Furgit         | 487 ms | 457 ms | 49 ms  |
| go-git         | 38 s   | 36 s   | 2 s    |

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

## Reporting bugs

Bug reports ideally include a reproduction recipe: a Go program which starts
out with an empty repository and calls Furgit and/or Git commands to trigger
undesirable behavior.

Please ask for help with writing your regression test before asking for your
problem to be fixed. Time invested in writing a regression test saves time
wasted on back-and-forth discussion about how the problem can be reproduced. A
regression test will need to be written in any case to verify a fix and prevent
the problem from resurfacing.

If writing an automated test really turns out to be impossible, please explain
in very clear terms how the problem can be reproduced.

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
