# Benchmarks

* See [gitbench](https://git.sr.ht/~runxiyu/gitbench).
* `legacy` branch furgit is slightly faster due to buffer reuse and custom
  ZLIB. These will be re-added.
* Alpine edge, i5-10210U, `performance` governor, `linux.git`.
* go-git may become much faster when
  [#1894](https://github.com/go-git/go-git/pull/1894)
  and such are fully in use.
* These lone tests do not represent all workloads. Test your usage
  pattern yourself (and contribute to gitbench).

## Traversing all trees in `HEAD` and fetching each file size

Mainly tests the packfile object reader.

| Implementation | Total  | User   | System |
| -              | -      | -      | -      |
| Git            | 337 ms | 226 ms | 108 ms |
| libgit2        | 391 ms | 269 ms | 120 ms |
| Furgit         | 487 ms | 457 ms | 49 ms  |
| go-git         | 37 s   | 35 s   | 2 s    |
