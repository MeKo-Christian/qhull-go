package qhull_test

import (
	"fmt"
	"log"

	qhull "github.com/cwbudde/qhull-go"
)

// ExampleDelaunay triangulates the four corners of a unit square. The corners are
// cocircular, so the diagonal is not determined by geometry — Delaunay reproduces
// the specific diagonal that matplotlib's Qhull backend chooses.
func ExampleDelaunay() {
	// 3---2
	// |   |
	// 0---1
	x := []float64{0, 1, 1, 0}
	y := []float64{0, 0, 1, 1}

	triangles, neighbors, err := qhull.Delaunay(x, y)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("triangles:", triangles)
	fmt.Println("neighbors:", neighbors)
	// Output:
	// triangles: [[3 0 1] [3 1 2]]
	// neighbors: [[-1 -1 1] [0 -1 -1]]
}
