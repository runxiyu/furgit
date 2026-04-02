# Contributing

Refer to the README for community spaces.

## Repos and mirrors

* [Main repository on Codeberg](https://codeberg.org/lindenii/furgit)
* [SourceHut mirror](https://git.sr.ht/~runxiyu/furgit)
* [tangled mirror](https://tangled.org/@runxiyu.tngl.sh/furgit)
* [GitHub mirror](https://github.com/runxiyu/furgit)
* [git.runxiyu.org mirror](https://git.runxiyu.org/furgit.git//)

## Reporting bugs

Bug reports ideally include a reproduction recipe: a Go program which starts
out with an empty repository and calls Furgit and/or Git commands to trigger
undesirable behavior. Feel free to ask for help writing them.

Choose any one of:

* [Open an issue on Codeberg](https://codeberg.org/lindenii/furgit/issues/new/choose)
* Sending email to [my public inbox](https://lists.sr.ht/~runxiyu/public-inbox)
* [Open an issue on GitHub](https://github.com/runxiyu/furgit/issues/new/choose)

## Submitting changes

When fixing a bug, please write a regression test in a separate commit before
your fix. Demonstrate that the regression test fails and that your fix lets it
succeed. Obviously, the regression test must be reasonable, demonstrate a real
issue (however minor), and must not itself contain hacks solely designed for
the purposes of making the old code fail and the new one succeed. Feel free to
ask for help writing regression tests.

Choose any one of:

* Open a [pull request on Codeberg](https://codeberg.org/lindenii/furgit/pulls)
* [Send a patch](https://git-send-email.io) to
  [my public inbox](https://lists.sr.ht/~runxiyu/public-inbox)
* Open a [pull request on GitHub](https://github.com/runxiyu/furgit/pulls)

## DCO sign-off

For the purposes of the Developer Certificate of Origin, the "open source
license" refers to the GNU Affero General Public License, Version 3.0, with
the above proxy designation in the README, pursuant to Section 14.

All contributors are required to "sign‐off" their commits (using `git commit
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
