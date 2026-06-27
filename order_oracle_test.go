package qhull

import (
	"testing"
)

// TestComputedOrderRidge measures the faithful ridge-graph engine
// (buildHullOrderRidge) against the captured ground truth, the same way as the
// vertex-set engine above. This is the development gate for Stage 3c; it logs
// per-category exact-order match counts and the divergent cases.
func TestComputedOrderRidge(t *testing.T) {
	corp := loadCorpus(t)
	orders := loadCreationOrders(t)

	byCat := map[string]*struct{ pass, total int }{}
	var diverged []string
	for _, tc := range corp.Cases {
		st := byCat[tc.Category]
		if st == nil {
			st = &struct{ pass, total int }{}
			byCat[tc.Category] = st
		}
		st.total++
		want := orders.Order[tc.Name]
		got, built := buildHullOrderRidge(project(tc.X, tc.Y))
		if built && sameIntSlice(got, want) {
			st.pass++
		} else {
			diverged = append(diverged, tc.Name)
		}
	}
	for _, cat := range []string{"general", "cocircular"} {
		if st := byCat[cat]; st != nil {
			t.Logf("ridge %-10s: %d/%d exact-order match", cat, st.pass, st.total)
		}
	}
	if len(diverged) > 0 {
		t.Logf("ridge divergences (%d): %v", len(diverged), diverged)
	}
	// General position has a unique build with no merges; the faithful engine must
	// reproduce its creation order exactly. Hard-gate it to prevent regression.
	if st := byCat["general"]; st != nil && st.pass != st.total {
		t.Errorf("ridge general exact-order regressed: %d/%d (must be %d)", st.pass, st.total, st.total)
	}
	// Cocircular ratchet: the faithful coplanar-horizon merge (newfacet2 leak +
	// per-merge qh_makeridges propagation + qh_mergesimplex swap-remove + pre-existing
	// ridge ordering) plus the qh_findbestnew partition switch on merge steps
	// (qh_addpoint uses the linear new-facet scan, not the directed walk, once a merge
	// produces a non-simplicial facet) reaches 34/34 exact-order. Never lower.
	const cocircularRidgeRatchet = 34
	if st := byCat["cocircular"]; st != nil && st.pass < cocircularRidgeRatchet {
		t.Errorf("ridge cocircular exact-order regressed below ratchet: %d/%d < %d",
			st.pass, st.total, cocircularRidgeRatchet)
	}
}

// sameIntSlice reports whether a and b are element-wise equal.
func sameIntSlice(a, b []int) bool {
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
