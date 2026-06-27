# Third-party licenses

This repository combines code under two different licenses.

## qhull-go (the Go package) — MIT

All Go source at the module root (`package qhull`) is original work, licensed
under the MIT License. See [`LICENSE`](LICENSE).

## Vendored Qhull 8.0.2 — Qhull license

`third_party/qhull-8.0.2/` is a vendored copy of [Qhull](http://www.qhull.org)
8.0.2 (Copyright © 1993–2020 C.B. Barber and The Geometry Center, University of
Minnesota). It is included as a porting reference and as the ground-truth test
oracle. It is **not** compiled into the importable Go package and carries no cgo
dependency for consumers.

This vendored tree is governed by its own permissive license, not by the MIT
license above. See [`third_party/qhull-8.0.2/COPYING.txt`](third_party/qhull-8.0.2/COPYING.txt)
for the full terms (the key conditions are that copyright notices remain intact
and that the origin of the software is not misrepresented).

### Instrumentation harnesses

The files `introspect.c`, `dump_state.c`, `stepdump.c`, and `order.py` under
`third_party/qhull-8.0.2/` were added by this project. They build and link
against the vendored Qhull library to capture Qhull's vertex creation order and
per-step merge trace as test fixtures (`testdata/creation_order.json`,
`testdata/corpus.json`). As derivative tooling distributed alongside Qhull, they
are made available under the same Qhull license terms. They are development-only
and are not part of the importable Go package.
