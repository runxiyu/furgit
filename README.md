# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit?)
[![Go Reference](https://pkg.go.dev/badge/git.sr.ht/~runxiyu/furgit.svg)](https://pkg.go.dev/git.sr.ht/~runxiyu/furgit)

Furgit is a fast Git library in pure Go.

## Project status

Furgit is in initial development, does not have tagged releases yet, and we can
guarantee that the API will break every now and then. Do not use in
production. When we do have tagged releases, we will likely follow
[Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

## History

Furgit's lineage is from [Villosa](https://codeberg.org/lindenii/villosa), a
new, work-in-progress software development forge. It started as an internal
package inside Villosa, and was later extracted into a standalone module.

Villosa's old internal Git package, in turn, based on my work in
[hare-git](https://forge.lindenii.org/hare/-/repos/hare-git/) for the Hare
programming language. Features in here will likely be ported back to
hare-git in the future.

## Goals and current features

We do not focus on command-line utilities; in particular, Furgit does not
intend to replace [upstream git](https://git-scm.com). It is intended to be
used as a library.

We intend for repository objects to be freely usable across goroutines, which
may enable long-running applications such as forges to keep a pool of recently
used repos (including their `.idx` and `.pack` cache) for rapid access.

There is no specific plan for features yet, but we'll initially focus on
developing what forges like [Villosa](https://codeberg.org/lindenii/villosa) and
[tangled](https://tangled.org/@tangled.org/core) (and other forges if
interested) requires. Afterwards, we'll take a look at what other usages (such
as writing Git clients, IDE integration, etc) would need.

Currently, furgit is very basic; it supports reading and writing loose objects
and reading from packfiles. There is some infrastructure for writing packfiles
in the tests but they need to be refactored.

## Dependencies

Furgit has no dependencies other than the standard library and some packages
from `golang.org/x`. It is unlikely that other dependencies will be introduced.

In some cases, external code is introduced but maintained in-tree.

## Environment requirements

We currently do not intend to support flexible storage backends such as
[storers in go-git](https://pkg.go.dev/github.com/go-git/go-git/v5/plumbing/storer);
a standard UNIX-like filesystem with
[syscall.Mmap](https://pkg.go.dev/syscall#Mmap) is expected.

## Performance

Furgit is being aggressively optimized for performance.

It is difficult to optimize Go code to be as performant as libgit2 (or
for that matter, upstream git). However, we are making tiny steps
towards it.

The first step that has been arguably a success is the packfile parser.
By using memory-mapped I/O, relatively optimized delta resolution, and
zero-copy techniques, Furgit is able to perform the equivalent to
`git ls-tree --long HEAD` on the Linux repository in about 2ms on
a ThinkPad T14, which is comparable to Git, faster than libgit2,
and significantly faster than go-git.

However this is a microbenchmark and does not reflect all real-world
performance. For example, when recursively listing tree entires and
commits, Furgit's performance is only slightly faster than libgit2;
both lack behind Git by multiple orders of magnitude.

Things we might consider in the future include:

* [commit-graph](https://git-scm.com/docs/commit-graph)
* Using a custom zlib implementation to amortize decompression overhead
  (see the `zlib` branch)
* More optimizations to delta resolution

## Hash algorithm

Furgit supports both SHA-256 and SHA-1.

The default tests run with SHA-256. To run tests with SHA-1, use the `sha1`
build tag.

## Active services using Furgit

There's an experimental instance of [Villosa](https://codeberg.org/lindenii/villosa)
hosting [a copy of Linux](https://villosa.lindenii.org/test//repos/linux/)
([tree](https://villosa.lindenii.org/test//repos/linux/HEAD/tree/)) using
Furgit as the Git backend.

## Repos and community resources

The [main repository](https://forge.lindenii.org/furgit/-/repos/furgit/) is
hosted on [Lindenii Forge](https://forge.lindenii.org/forge/-/repos/server/)
(the previous iteration of [Villosa](https://codeberg.org/lindenii/villosa)).

To contribute, clone the repository from the SSH remote
`ssh://forge.lindenii.org/forge/-/repos/server`, create a unique branch that
begins with `contrib/`, and push. Your branch will be associated with your SSH
key and a merge request will be created, and the maintainers will be notified
on IRC.

Anonymous SSH cloning is supported with or without a key. Pushing requires an
SSH key: no key pre-registration is required, but you have to ensure that your
key is consistent throughout pushes if you push multiple times.

```
git clone ssh://forge.lindenii.org/furgit/-/repos/furgit
cd furgit
git checkout -b contrib/name_of_your_contribution
# edit and commit stuff
git push -u origin HEAD
```

There are also a few mirrors. The maintainer may poll them for
issues/patches/PRs/etc., but pushing to Lindenii Forge is recommended.

* [SourceHut](https://git.sr.ht/~runxiyu/furgit)
* [Codeberg](https://codeberg.org/runxiyu/furgit)
* [tangled](https://tangled.org/@runxiyu.tngl.sh/furgit)
* [GitHub](https://github.com/runxiyu/furgit)

We discuss in `#chat` on [irc.runxiyu.org](https://irc.runxiyu.org)
([web chat](https://webirc.runxiyu.org/kiwiirc/#chat))

The maintainer is working through college applications and IBDP coursework and
may not necessarily respond in time.

## Etymology

I was thinking of names and I accidentally typed "git" as "fur" (i.e., left
shifted one key on my QWERTY keyboard).

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

## Random internal to-do list

* Consider adding repository methods that attempt to resolve objects
  of a particular type. They would attempt to resolve the object's
  header and return an error if the type mismatches; if it matches,
  they continue from that point (passing along some state such as
  the packLocation to avoid re-resolving the location from index
  files).
