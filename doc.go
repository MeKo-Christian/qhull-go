// Package qhull computes 2-D Delaunay triangulations whose connectivity matches
// matplotlib's Qhull backend (Qhull 8.0.2, options "d Qt Qbb Qc Qz").
//
// # Why this exists
//
// Most Delaunay libraries return *a* valid triangulation. For points in general
// position that is enough: the Delaunay triangulation is unique, so every correct
// construction agrees. But on cocircular inputs — four or more points on a common
// empty circle — the triangulation is non-unique, and the particular diagonal
// Qhull picks is fixed by its internal construction order, not by geometry. This
// package reproduces that choice, which makes it a drop-in for matplotlib-parity
// work (it was extracted from matplotlib-go).
//
// # API
//
// Two entry points, both taking parallel x and y slices and returning the
// triangles and per-triangle neighbours:
//
//   - [Delaunay] is the default, parity-matching path. It reproduces Qhull's
//     cocircular diagonal by resolving it from the computed vertex creation order,
//     and falls back to the exact baseline if that computation cannot complete.
//     Use it whenever you want matplotlib/Qhull connectivity.
//   - [DelaunayFast] is the robust exact-predicate baseline. It returns a valid
//     Delaunay triangulation but makes an arbitrary cocircular diagonal choice. It
//     is identical to [Delaunay] in general position and cheaper on
//     cocircular-heavy inputs.
//
// Both return triangles and neighbors as [][3]int. triangles[i] holds the three
// input-point indices of triangle i in anticlockwise winding; neighbors[i][j] is
// the triangle across the edge from vertex j to vertex (j+1)%3, or -1 on the
// convex-hull boundary. Inputs must have equal-length x and y and at least three
// points.
//
// The implementation is pure Go: no external dependencies and no cgo.
package qhull
