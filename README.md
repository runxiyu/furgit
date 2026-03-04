# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit.svg)](https://builds.sr.ht/~runxiyu/furgit)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/lindenii/furgit.svg)](https://pkg.go.dev/codeberg.org/lindenii/furgit)

Furgit is a somewhat fast Git library in Go.

## Project status

* Initial development
* Frequent breaking changes
* Do not use in production
* May use [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) later

Currently undergoing a massive redesign; see the `legacy` branch too.

## Roadmap

**Done**

* Parsing configs
* Object ID and hash algorithms
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

**Planned**

* Streaming `.pack`/`.idx` hash verifier
* `.idx` and `.rev` writing
* Pack ingestion
* Un-thinning thin packs
* Quarantine areas
* V0/V1 protocol (e.g., `sideband-64k`, `pktline`, etc.)
* Receiving/fetching
* Writing configs
* Delta base selection, e.g., islands
* Compressing deltas
* Writing packfiles
* Commit graph
* Reachability bitmaps
* Thin packs
* DEFLATE optimizations
* Aggressive buffer pooling

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

## Reporting bugs

All problem/bug reports should include a reproduction recipe in form
of a Go program which starts out with an empty repository and runs a
series of Furgit functions/methods and/or Git commands to trigger the
problem, be it a crash or some other undesirable behavior.

Please take this request very seriously; Ask for help with writing your
regression test before asking for your problem to be fixed. Time invested in
writing a regression test saves time wasted on back-and-forth discussion about
how the problem can be reproduced. A regression test will need to be written in
any case to verify a fix and prevent the problem from resurfacing.

If writing an automated test really turns out to be impossible, please
explain in very clear terms how the problem can be reproduced.

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
