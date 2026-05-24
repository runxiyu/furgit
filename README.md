# Furgit

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/furgit/commits/next.svg)](https://builds.sr.ht/~runxiyu/furgit/commits/next)
[![Go Reference](https://pkg.go.dev/badge/codeberg.org/lindenii/furgit.svg)](https://pkg.go.dev/codeberg.org/lindenii/furgit)

A low‐level Git plumbing library in Go.

## Status

Furgit is an unfinished research-ish project,
and is **not suitable for production**.

The architecture is rather deliberate
and the parts that exist have been
built carefully to our capability,
and tested against git as the canonical oracle, etc.

However, the API is not yet settled at all;
we may revise interfaces and behaviour at any time,
to incorporate findings from ongoing audits,
including full rearchitectures.
Consumers that depend on a 0.x version of furgit
will need to revise their usages
every few days or weeks.

Therefore, **you should probably not use furgit**.
Instead, use [go-git](https://github.com/go-git/go-git),
which has API stability promises,
or simply `os/exec` to git,
which is sufficient for many purposes.

## Community

* [#lindenii](https://webirc.runxiyu.org/kiwiirc/#lindenii)
  on [irc.runxiyu.org](https://irc.runxiyu.org)
* [#lindenii](https://web.libera.chat/#lindenii)
  on [Libera.Chat](https://libera.chat)

See also `CONTRIBUTING.md`.

## Acknowledgements

Partly inspired by [upstream git](https://git-scm.com),
OpenBSD's [Game of Trees](https://gameoftrees.org), and
[9front git](https://git.9front.org/plan9front/9front/HEAD/sys/src/cmd/git/f.html).

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


## Repos and mirrors

* [Codeberg](https://codeberg.org/lindenii/furgit)
* [SourceHut mirror](https://git.sr.ht/~runxiyu/furgit)
* [hjgit mirror](https://hjgit.org/runxiyu/furgit/HEAD/info.html)
* [tangled mirror](https://tangled.org/@runxiyu.tngl.sh/furgit)
* [GitHub mirror](https://github.com/runxiyu/furgit)
* [git.runxiyu.org mirror](https://git.runxiyu.org/furgit.git//)
* [cgit.space mirror](https://cgit.space/~runxiyu/furgit.git/)
