# qhull-go

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
import qhull "github.com/MeKo-Tech/qhull-go"

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
- **Cocircular:** 33/34 corpus cases match Qhull's exact build order / diagonal
  (60/61 across the combined order lock). One case (`grid5x4`) is a graceful,
  still-valid-Delaunay fallback — see `PLAN.md`.

## Ground-truth oracle

`third_party/qhull-8.0.2/` holds the vendored Qhull source (porting reference)
plus small instrumentation tools (`introspect.c`, `dump_state.c`, `stepdump.c`)
used to capture Qhull's creation order and per-step merge trace as test fixtures
(`testdata/creation_order.json`, `testdata/corpus.json`). See `PLAN.md` for the
build recipe and vendoring plan.

## License

The Go code is MIT-licensed (see `LICENSE`, to be added). The vendored Qhull
source under `third_party/qhull-8.0.2/` retains its original Qhull license
(`third_party/qhull-8.0.2/COPYING.txt`).
