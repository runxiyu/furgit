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

## Features

* Configuration
  * [X] Parsing
  * [ ] Includes
  * [ ] Writing
* [X] Object IDs
  * [X] SHA-256
  * [X] SHA-1
* [X] Object model (incl., parse, serialize)
  * [X] Blobs
  * [X] Trees
    * [X] File mode definitions
    * [X] Entry insertion ordering
    * [X] Traversal
    * [ ] Pathspec
  * [X] Commits
  * [X] Annotated tags
  * [X] Stored objects
* Further cryptography
  * [ ] OpenPGP signatures
  * [ ] SSH signatures
* [X] Reading object stores
  * [X] Pluggable interface
  * [X] Chain lookup store
  * [X] Bundle store
  * [X] MRU lookup store
  * [X] Reading loose objects
  * [ ] Promisor remotes
  * [ ] Alternates
  * [X] Reading packed objects
    * [X] Pack index lookups
    * [X] Delta caching
    * [X] Delta application
    * [ ] Pack-wide bloom filters
    * [ ] Multi pack indexes
* [ ] Writing objects
  * [X] Loose object writing
* Misc bundle features
  * [ ] Writing bundles
* Misc packfile features
  * [X] Writing pack indexes
  * [X] Writing reverse pack indexes
  * [ ] Writing packfiles
    * [ ] Writing thin packs
    * [ ] Compressing deltas
      * [ ] Delta islands
  * [ ] Pack verification
* Compression
  * [ ] Plugabble compression algorithms
  * [X] ZLIB support
  * [ ] DEFLATE optimizations
  * [X] Adler-32 SIMD optimizations
* [X] References
  * [X] Detached references
  * [X] Symbolic references
  * [X] Name verification/resolution
  * [X] Annotated tag ref peeling
  * [ ] Describe
  * [ ] Revision syntax
  * [ ] Namespaces
  * [ ] Repalce refs, grafts
* [X] Reference stores
  * [X] Chain lookup store
  * [X] Files reference store
    * [X] Reading loose refs
    * [X] Reading packed refs
    * [X] Atomic writes
    * [X] Batched writes
    * [ ] Packing refs
    * [ ] Reflogs
  * [ ] Reftable
* Reachability
  * [X] Have/wants walks
  * [X] Is ancestor
  * [X] Merge bases
  * [X] Commit graph
    * [X] Changed path bloom filters
    * [X] Chained graphs
    * [ ] Writing
  * [ ] Reachability bitmaps
    * [ ] For a single packfile
    * [ ] For Multi pack indexes
* Misc repository
  * [X] Opening relevant stores
  * [ ] Creating repositories
  * [ ] Filter branch/repo
  * [ ] Fast import/export
  * [ ] Git notes
  * [ ] Git attributes
  * [ ] Pseudorefs
  * Integrity and maintainence
    * [ ] Fsck
    * [ ] Repacking
    * [ ] Garbage collection
    * [ ] Cruft packing
    * [ ] Expiration
  * [ ] Grep
  * [ ] Submodules
  * [ ] Worktrees
  * [ ] Archive
  * [ ] LFS
  * [ ] Revision log walk
    * [ ] Topological ordering
    * [ ] Date ordering
    * [ ] Path-limited
* [ ] Diffing
  * [ ] Blame
  * [ ] Annotate
  * [X] Tree diffing
    * [ ] Similarity/rename/copy detection
  * [ ] Multi-way diffs
  * [ ] Patch-id
  * [ ] Range-diff
  * Blob diffing
    * [ ] Word diffs
    * [X] Myers
    * [ ] Patience
    * [ ] Histogram
    * [ ] Tree-way
  * [ ] Format patch
  * [ ] Apply/amend patch
* Branch integration/rewrite/etc methods
  * [ ] Merge
    * [ ] Recursive
    * [ ] ORT
  * [ ] Rebase
  * [ ] Cherry pick
  * [ ] Revert
  * [ ] Rerere
* Network protocols and related features
  * [X] pkt-line
  * [X] side-band-64k
  * [X] Ingesting packfiles
    * [X] Quarantine areas
    * [X] Un-thinning thin packs
  * Version 0, version 1 protocols
    * [X] Server side
      * [X] Reference advertisement
      * [X] Capability negotiation
      * [X] Receive
      * [ ] "Upload"
    * [ ] Client side
      * [ ] Send
      * [ ] Fetch
  * Version 2 protocol
    * [ ] Server side
      * [ ] "Upload"
    * [ ] Client side
      * [ ] Fetch
  * Protocol-independent logic
    * Common
      * [X] Progress meters
    * Client side
      * [ ] Refspec
      * [ ] Fetch
        * [ ] Partial clones
          * [ ] Object filtering
        * [ ] Bundle URI
        * [ ] Packfile URI
        * [ ] Shallow clones
      * [ ] Send
    * Server side
      * [ ] Upload
        * [ ] Object filtering
      * [X] Receive
        * [ ] Signed push
        * Hooks
          * Slots
            * [ ] After ref negotiation
            * [X] After object unpacking
          * Provided samples
            * [X] Chain
            * [X] Force push rejection
* [ ] Working trees
  * [ ] Stashing
  * [ ] Ignore rules
  * [ ] Checkouts
    * [ ] Sparse checkouts
    * [ ] CR/LF conversions
    * [ ] File mode conversions
  * [ ] Indexes
    * [ ] Conflict resolution
    * [ ] Split index
    * [ ] Sparse index
    * [ ] Untracked cache
  * [ ] Status listing
  * [ ] Filesystem monitor
  * [ ] Worktree
    * [ ] Common directory
    * [ ] Worktree-specific references
* Research
  * [ ] Dynamic packfiles
    * [ ] Compaction; page-sized hole punching
    * [ ] Dynamic indexing
      * [ ] Linear/extendible/spiral hashing
    * [ ] Dynamic reachability bitmaps

## Not planned

* Any CLI tools whatsoever
* Clone
* Anything reasonably considered "porcelain"
* Credential helper
* Transports
* Auth
* Remote management
* Bisect
* Any use of env vars
* Repository discovery walking

I might make a second project that supports these.
Furgit will probably not, and will remain sans-IO.

## Benchmarks

* See [gitbench](https://git.sr.ht/~runxiyu/gitbench).
* `legacy` branch furgit is slightly faster due to buffer reuse and custom
  ZLIB. These will be re-added.
* Alpine edge, i5-10210U, `performance` governor, `linux.git`.
* My copy of go-git is v6 with
  [#1894](https://github.com/go-git/go-git/pull/1894)
  and other optimizations merged in.
* These lone tests do not represent all workloads. Test your usage
  pattern yourself (and contribute to gitbench).

### Traversing all trees in `HEAD` and fetching each file size

Mainly tests the packfile object reader.

| Implementation | Total  | User   | System |
| -              | -      | -      | -      |
| Git            | 337 ms | 226 ms | 108 ms |
| libgit2        | 391 ms | 269 ms | 120 ms |
| Furgit         | 487 ms | 457 ms | 49 ms  |
| go-git         | 4.2 s  | 2.2 s  | 2.0 s  |

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
* Some architectual elements inspired by [upstream Git](https://git-scm.com),
  OpenBSD's [Game of Trees](https://gameoftrees.org), and
  [9front Git](https://git.9front.org/plan9front/9front/HEAD/sys/src/cmd/git/f.html).

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
