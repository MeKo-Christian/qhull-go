# AGENTS.md

This file provides guidance to ai agents (claude code, codex etc.) when working with code in this repository.

## What this is

A pure-Go (no cgo, no external deps — stdlib only) 2-D Delaunay triangulator that reproduces the **exact connectivity** of Qhull 8.0.2 under matplotlib's options (`qhull d Qt Qbb Qc Qz`). It was extracted from `matplotlib-go`'s `tri/qhull` package as the future source of truth; matplotlib-go is not yet wired to depend on it (see PLAN.md §7).

The point of the project: most Delaunay libraries return *a* valid triangulation, which suffices for general-position inputs where the triangulation is unique. But on **cocircular** inputs (≥4 points on a common empty circle) the triangulation is non-unique, and the specific diagonal Qhull picks is fixed by its internal vertex *creation order*, not by geometry. This package reproduces that choice so it can be a drop-in for matplotlib-parity work.

## Commands

All workflows go through `just` (see `justfile`):

```bash
just build            # go build ./...
just test             # go test ./...  (includes the runnable doc example)
just test-race        # go test -race ./...
just test-coverage    # coverage.out + coverage.html
just bench            # go test -run='^$' -bench=. -benchmem ./...
just vet              # go vet ./...
just lint             # golangci-lint run ./...   (v2; config .golangci.yml)
just lint-fix         # golangci-lint run --fix ./...
just fmt              # golangci-lint fmt          (gofumpt)
just check-formatted  # golangci-lint fmt --diff   (CI gate)
just check-tidy       # go mod tidy + git diff --exit-code
just ci               # check-formatted + vet + test + lint + check-tidy  (run before pushing)
just fix              # lint-fix + fmt
```

Run a single test (everything is in the root `qhull` package):

```bash
go test -run TestDelaunayMatchesQhullCorpus ./...
go test -run TestComputedOrderRidge ./...
go test -run ExampleDelaunay ./...
```

CI (`.github/workflows/ci.yml`) runs unit tests on the Go **1.24.x and 1.25.x** matrix, golangci-lint **v2.12.2**, and a gofumpt format gate. There is no cgo and no system dependency — the Qhull oracle (below) is needed only to *regenerate* fixtures, never to build or test.

## Architecture

Single package `qhull`, all files at the repo root. The public surface is intentionally just two functions (frozen for v0.1.0; do not expand without reason — see PLAN.md §2):

- `Delaunay(x, y []float64) (triangles, neighbors [][3]int, err error)` — default, Qhull/matplotlib-matched.
- `DelaunayFast(x, y []float64) (...)` — robust exact-predicate baseline; valid but arbitrary cocircular diagonal.

Both take parallel `x`/`y` slices (equal length, ≥3 points) and return `[][3]int`: `triangles[i]` is three input-point indices in anticlockwise winding; `neighbors[i][j]` is the triangle across the edge from vertex `j` to `(j+1)%3`, or `-1` on the convex-hull boundary.

**Two-layer design.** The exported `Delaunay` (`fanfromorder.go:79`) is layered on top of the exact baseline:

1. `DelaunayFast` (`delaunay.go`) — exact-predicate incremental insertion. Predicates (`predicates.go`) use `math/big.Rat`, so `orient2d`/`inCircle` return the mathematically exact sign with no epsilon. This yields the unique, correct Delaunay connectivity for general-position inputs.
2. `Delaunay` then groups cocircular cells. If none exist (general position), the exact result is already canonical and is returned as-is — the fast path.
3. If cocircular cells exist, it computes Qhull's vertex **creation order** via `buildHullOrderRidge` (`ridgebuild.go`) and re-fans each cocircular cell from its last-created vertex (`fanfromorder.go`), reproducing Qhull's exact diagonal. If the order computation bails on a degeneracy, it falls back to the valid baseline result.

Key files:

- `predicates.go` — exact `orient2d` / `inCircle` over `math/big.Rat`.
- `delaunay.go` — `DelaunayFast`, the exact-predicate baseline + neighbour derivation.
- `build.go` — `project()`: paraboloid lift, Qz "infinity" point, `qh_maxmin` tolerance pass, Qbb last-coordinate scaling. Faithful to Qhull's float64 arithmetic so the downstream creation order matches bit-for-bit.
- `ridgebuild.go` — `buildHullOrderRidge`, the faithful re-port of Qhull's incremental hull. The decisive detail is the data layout: each facet keeps Qhull's own vertex set (inverse-sorted by creation id) and a parallel neighbour array, so neighbour iteration order matches Qhull and the computed creation order matches exactly. This is the largest and most delicate file.
- `fanfromorder.go` — exported `Delaunay` + `delaunayFromOrder` (the cocircular fan geometry that consumes the creation order).
- `computed.go` — `delaunayComputed`, a fully self-contained Qhull-faithful engine (no fixtures) used by tests/development; chains `buildHullOrderRidge` → `delaunayFromOrder` unconditionally.
- `doc.go` — package overview (the canonical narrative; `go doc` shows it).

## The oracle and parity tests

Parity is validated against Qhull 8.0.2's actual output, captured as **committed** fixtures in `testdata/` (`corpus.json`, `creation_order.json`). Because they're committed, nothing here needs Qhull installed.

Test gates (don't lower these casually — connectivity is also validated downstream by matplotlib-go's golden suite):

- `TestDelaunayMatchesQhullCorpus` (`corpus_test.go`) — **hard gate**: public `Delaunay` must reproduce Qhull's exact connectivity (diagonal included) for all corpus cases: 27/27 general position, 34/34 cocircular.
- `TestComputedOrderRidge` (`order_oracle_test.go`) — creation-order ratchet (`cocircularRidgeRatchet = 34`); never lower it.
- `TestDelaunayComputed` (`buildhull_test.go`) — end-to-end computed engine vs the differential corpus.
- `TestDelaunayConnectivityVsQhull` — intentionally a **loose** gate: it covers `DelaunayFast`, which does not match the cocircular diagonal.
- `usage_test.go` — fixture-independent public-API tests (external `qhull_test` package): error paths, determinism, structural invariants (anticlockwise winding, every point used, neighbour-graph symmetry).

The `oracle/` directory holds only our own C instrumentation tools (`introspect.c`, `dump_state.c`, `stepdump.c`) plus `instrumentation.patch`. Qhull itself is **not redistributed** — it's a local, gitignored dev dependency under `third_party/qhull-8.0.2/`. To regenerate fixtures: do the one-time setup in `oracle/README.md` (download + sha256-verify the pinned 8.0.2 tarball, apply the patch), then `just oracle-build`, then run the `testdata/gen_*.py` scripts. Gotcha noted in PLAN.md/oracle README: `stepdump`'s `TA<n>` stop option *suppresses merging* — use `introspect` with `QHSTEP=1` for the real merging per-step trace.

## Conventions

- gofumpt formatting (stricter than gofmt); enabled linters add `misspell`, `gocritic`, `revive`. `revive` enforces a 1500-line file-length limit — `ridgebuild.go` is the file to watch.
- Don't change Delaunay connectivity casually: it's a parity contract validated downstream against matplotlib goldens, which are regenerated from matplotlib and never hand-edited.
- `PLAN.md` is the live roadmap. §1–§6 are done; §7 (the matplotlib-go cutover + `v0.1.0` tag) is intentionally deferred pending an explicit go-ahead.
