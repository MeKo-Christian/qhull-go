# Third-party licenses

## qhull-go (this repository) — MIT

Everything published in this repository is original work licensed under the MIT
License (see [`LICENSE`](LICENSE)). The importable Go package has **no external
dependencies and no cgo** — standard library only.

## Qhull — local-only dev/test oracle, not redistributed here

This package was built to reproduce the exact connectivity of
[Qhull](http://www.qhull.org) 8.0.2. Qhull's source is used **locally** as a
porting reference and as the ground-truth oracle that regenerates the test
fixtures in `testdata/` (creation order and per-step merge trace).

That Qhull source is **not vendored or redistributed in this repository** — the
`third_party/` directory is gitignored. To regenerate fixtures, obtain Qhull
8.0.2 separately and place it under `third_party/qhull-8.0.2/`; the exact pinned
tarball (URL + sha256), the one-time setup, our own instrumentation harnesses
(`oracle/introspect.c`, `oracle/dump_state.c`, `oracle/stepdump.c`), and the
trace-`printf` patch (`oracle/instrumentation.patch`) are all documented in
[`oracle/README.md`](oracle/README.md). Qhull retains its own license (a
permissive license from C.B. Barber and The Geometry Center) — consult the
`COPYING.txt` shipped with the Qhull source you download.

Because the Qhull source is not part of this repository, running `go build`,
`go test`, and `go get` against qhull-go never touches it.
