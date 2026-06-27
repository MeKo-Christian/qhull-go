# qhull-go

[![CI](https://github.com/MeKo-Christian/qhull-go/actions/workflows/ci.yml/badge.svg)](https://github.com/MeKo-Christian/qhull-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/MeKo-Christian/qhull-go.svg)](https://pkg.go.dev/github.com/MeKo-Christian/qhull-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A pure-Go 2-D Delaunay triangulator that aims to reproduce the **exact
connectivity** produced by [Qhull](http://www.qhull.org) 8.0.2 with matplotlib's
options (`qhull d Qt Qbb Qc Qz`).

Most Go Delaunay libraries return *a* valid triangulation. For points in general
position that is enough — the triangulation is unique. But on **cocircular**
inputs (≥4 points on a common empty circle) the triangulation is non-unique, and
the specific diagonal Qhull picks is fixed by its internal construction order, not
by geometry. This package reproduces that choice, which is what makes it suitable
as a drop-in for matplotlib-parity work (it was extracted from
[`matplotlib-go`](https://github.com/cwbudde/matplotlib-go)).

No cgo. No external dependencies — standard library only.

## API

```go
import qhull "github.com/MeKo-Christian/qhull-go"

// Default: matplotlib/Qhull-matched connectivity, with the cocircular diagonal
// resolved from the computed vertex creation order. This is what you want for
// parity. Falls back to the exact baseline if the order computation ever bails.
tris, neighbors, err := qhull.Delaunay(x, y)

// Robust exact-predicate baseline. Identical to Delaunay for general position;
// on cocircular cells it returns a valid but arbitrary diagonal (no Qhull
// matching) and is cheaper on cocircular-heavy inputs.
tris, neighbors, err := qhull.DelaunayFast(x, y)
```

Both return `triangles [][3]int` (anticlockwise vertex indices) and `neighbors
[][3]int`, where `neighbors[i][j]` is the triangle across the edge from vertex `j`
to `(j+1)%3`, or `-1` on the convex-hull boundary.

## Status

- **General position:** 27/27 corpus cases match Qhull's connectivity exactly.
- **Cocircular:** 34/34 corpus cases match Qhull's exact build order / diagonal
  (61/61 across the combined order lock).

## Ground-truth oracle

Parity is validated against Qhull's actual output, captured as fixtures in
`testdata/` (`creation_order.json`, `corpus.json`). The fixtures are committed, so
building and testing this package needs nothing beyond the Go toolchain.

Regenerating the fixtures requires the Qhull 8.0.2 source plus the small
instrumentation tools (`introspect.c`, `dump_state.c`, `stepdump.c`). That source
is **not redistributed here** — it is a local, gitignored dev dependency under
`third_party/qhull-8.0.2/`. Obtain Qhull from <http://www.qhull.org>; see
`THIRD_PARTY.md` and `PLAN.md` for the layout and build recipe.

## License

All code published here is MIT-licensed (see `LICENSE`). Qhull is used only as a
local dev/test oracle and is **not redistributed** in this repository; see
`THIRD_PARTY.md`.
