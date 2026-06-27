package qhull_test

import (
	"testing"

	qhull "github.com/cwbudde/qhull-go"
)

// This file exercises the public API (Delaunay, DelaunayFast) directly, without
// the Qhull oracle or the captured corpus fixtures, so the exported surface has
// standalone coverage and the package's invariants are checked on their own.

// checkTriangulation asserts the structural invariants every result of Delaunay /
// DelaunayFast must satisfy: valid, distinct, anticlockwise-wound vertex indices;
// every input point used; and a consistent neighbour graph (symmetry + shared-edge
// correspondence, with -1 only on the hull boundary).
func checkTriangulation(t *testing.T, x, y []float64, tris, nbrs [][3]int) {
	t.Helper()
	n := len(x)
	if len(nbrs) != len(tris) {
		t.Fatalf("neighbors length %d != triangles length %d", len(nbrs), len(tris))
	}
	used := make([]bool, n)
	for i, tr := range tris {
		for j := range 3 {
			v := tr[j]
			if v < 0 || v >= n {
				t.Fatalf("triangle %d vertex %d out of range [0,%d): %d", i, j, n, v)
			}
			used[v] = true
			if tr[j] == tr[(j+1)%3] {
				t.Fatalf("triangle %d has a repeated vertex: %v", i, tr)
			}
		}
		// Anticlockwise winding: signed area > 0.
		ax, ay := x[tr[0]], y[tr[0]]
		bx, by := x[tr[1]], y[tr[1]]
		cx, cy := x[tr[2]], y[tr[2]]
		if area := (bx-ax)*(cy-ay) - (by-ay)*(cx-ax); area <= 0 {
			t.Fatalf("triangle %d %v is not anticlockwise (2*area=%g)", i, tr, area)
		}
	}
	for v := range used {
		if !used[v] {
			t.Errorf("input point %d does not appear in any triangle", v)
		}
	}
	checkNeighbors(t, tris, nbrs)
}

// checkNeighbors verifies that nbrs[i][j] is the triangle across the edge from
// vertex j to (j+1)%3: the reverse edge must belong to that neighbour and point
// back at i, and a -1 edge must be unshared (a true hull boundary).
func checkNeighbors(t *testing.T, tris, nbrs [][3]int) {
	t.Helper()
	// Map each directed edge (u,v) to the triangle that owns it.
	owner := map[[2]int]int{}
	for i, tr := range tris {
		for j := range 3 {
			e := [2]int{tr[j], tr[(j+1)%3]}
			if prev, dup := owner[e]; dup {
				t.Fatalf("directed edge %v shared by triangles %d and %d (non-manifold)", e, prev, i)
			}
			owner[e] = i
		}
	}
	for i, tr := range tris {
		for j := range 3 {
			u, v := tr[j], tr[(j+1)%3]
			m := nbrs[i][j]
			reverse, shared := owner[[2]int{v, u}]
			switch {
			case m < 0 && shared:
				t.Errorf("triangle %d edge %d (%d->%d) marked hull (-1) but is shared by triangle %d", i, j, u, v, reverse)
			case m >= 0 && !shared:
				t.Errorf("triangle %d edge %d (%d->%d) claims neighbour %d but the reverse edge is unshared", i, j, u, v, m)
			case m >= 0 && shared && m != reverse:
				t.Errorf("triangle %d edge %d neighbour is %d, want %d", i, j, m, reverse)
			}
		}
	}
}

// bothEngines runs a check against both public entry points; for general-position
// inputs they must agree, and both must satisfy the invariants.
func bothEngines(t *testing.T, name string, x, y []float64) (triFast, triFull [][3]int) {
	t.Helper()
	tf, nf, err := qhull.DelaunayFast(x, y)
	if err != nil {
		t.Fatalf("%s: DelaunayFast: %v", name, err)
	}
	checkTriangulation(t, x, y, tf, nf)

	td, nd, err := qhull.Delaunay(x, y)
	if err != nil {
		t.Fatalf("%s: Delaunay: %v", name, err)
	}
	checkTriangulation(t, x, y, td, nd)
	return tf, td
}

func TestSingleTriangle(t *testing.T) {
	x := []float64{0, 1, 0}
	y := []float64{0, 0, 1}
	tris, nbrs, err := qhull.Delaunay(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if len(tris) != 1 {
		t.Fatalf("want 1 triangle, got %d: %v", len(tris), tris)
	}
	checkTriangulation(t, x, y, tris, nbrs)
	if nbrs[0] != [3]int{-1, -1, -1} {
		t.Errorf("lone triangle should have no neighbours, got %v", nbrs[0])
	}
}

func TestUnitSquareIsCocircular(t *testing.T) {
	// Four corners of a square lie on a common circle: a single cocircular cell
	// that Delaunay splits with Qhull's deterministic diagonal.
	x := []float64{0, 1, 1, 0}
	y := []float64{0, 0, 1, 1}
	tris, nbrs, err := qhull.Delaunay(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if len(tris) != 2 {
		t.Fatalf("a square triangulates into 2 triangles, got %d: %v", len(tris), tris)
	}
	checkTriangulation(t, x, y, tris, nbrs)
	// The two triangles share exactly one diagonal: each has exactly one non-hull
	// edge, and they reference each other.
	interior := 0
	for i := range tris {
		for j := range 3 {
			if nbrs[i][j] >= 0 {
				interior++
			}
		}
	}
	if interior != 2 {
		t.Errorf("square should have one shared diagonal (2 directed interior edges), got %d", interior)
	}
}

func TestGeneralPositionEnginesAgree(t *testing.T) {
	// No four points cocircular, so Delaunay is unique and both engines must agree.
	x := []float64{0, 3, 7, 5, 9, 2}
	y := []float64{0, 1, 0, 4, 5, 6}
	fast, full := bothEngines(t, "generic", x, y)
	if !sameTriangleList(fast, full) {
		t.Errorf("general position: Delaunay and DelaunayFast disagree\n fast=%v\n full=%v", fast, full)
	}
}

func TestDeterministic(t *testing.T) {
	x := []float64{0, 1, 2, 0, 1, 2, 0, 1, 2}
	y := []float64{0, 0, 0, 1, 1, 1, 2, 2, 2}
	t1, n1, err := qhull.Delaunay(x, y)
	if err != nil {
		t.Fatal(err)
	}
	checkTriangulation(t, x, y, t1, n1)
	for range 5 {
		t2, n2, err := qhull.Delaunay(x, y)
		if err != nil {
			t.Fatal(err)
		}
		if !equalTriList(t1, t2) || !equalTriList(n1, n2) {
			t.Fatalf("Delaunay is not deterministic across calls")
		}
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		name string
		x, y []float64
	}{
		{"length mismatch", []float64{0, 1, 2}, []float64{0, 1}},
		{"too few points", []float64{0, 1}, []float64{0, 1}},
		{"empty", nil, nil},
		{"all collinear", []float64{0, 1, 2, 3}, []float64{0, 1, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := qhull.Delaunay(tc.x, tc.y); err == nil {
				t.Errorf("Delaunay(%v,%v): want error, got nil", tc.x, tc.y)
			}
			if _, _, err := qhull.DelaunayFast(tc.x, tc.y); err == nil {
				t.Errorf("DelaunayFast(%v,%v): want error, got nil", tc.x, tc.y)
			}
		})
	}
}

func sameTriangleList(a, b [][3]int) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[[3]int]int{}
	for _, t := range a {
		seen[normTri(t)]++
	}
	for _, t := range b {
		seen[normTri(t)]--
	}
	for _, c := range seen {
		if c != 0 {
			return false
		}
	}
	return true
}

// normTri rotates a triangle to its smallest-first representation so the same
// anticlockwise triangle compares equal regardless of which vertex is listed first.
func normTri(t [3]int) [3]int {
	lo := 0
	for i := 1; i < 3; i++ {
		if t[i] < t[lo] {
			lo = i
		}
	}
	return [3]int{t[lo], t[(lo+1)%3], t[(lo+2)%3]}
}

func equalTriList(a, b [][3]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
