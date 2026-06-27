# qhull-go — Roadmap

This repo was extracted from `matplotlib-go`'s `tri/qhull` package (a faithful
port of matplotlib's Qhull 8.0.2 Delaunay backend). The code is copied here as the
**future source of truth**; matplotlib-go still uses its own in-tree copy and is
**not yet** wired to depend on this module. This document tracks everything needed
to make qhull-go a properly published, standalone library and to complete the
cutover.

Status legend: `[ ]` todo · `[~]` in progress · `[x]` done.

---

## 0. Extraction (done in this pass)

- [x] Create repo at `../qhull-go`, `git init`.
- [x] Copy all `tri/qhull/*.go` to the module root (package `qhull`).
- [x] Copy `testdata/` (corpus.json, creation_order.json, gen_*.py).
- [x] Keep the ground-truth oracle at `third_party/qhull-8.0.2/` (patched Qhull
      source + `introspect.c` / `dump_state.c` / `stepdump.c` / `order.py`) as a
      **local, gitignored** dev dependency — not redistributed in this repo.
- [x] `go.mod` → `module github.com/MeKo-Christian/qhull-go`, `go 1.25.0`.
- [x] `.gitignore`, `README.md`, this `PLAN.md`.
- [x] Verify standalone: `go build ./...`, `go vet ./...`, `go test ./...` all green.

---

## 1. Repo / publishing setup

- [x] **Create the GitHub repo** `MeKo-Christian/qhull-go` and push the initial commit.
- [x] **Add a `LICENSE`** — MIT, for the Go code. The Qhull oracle is not
      redistributed here (see vendoring decision below), so the published repo is
      MIT-only.
- [x] Add a short `THIRD_PARTY.md` clarifying that the published repo is MIT-only
      and that Qhull is a local, gitignored dev/test oracle (not redistributed).
- [x] **Confirm the module path.** `github.com/MeKo-Christian/qhull-go` (the package
      name stays `qhull`; consumers import the repo path and refer to it as `qhull`).
      `go.mod` + README + LICENSE copyright updated to match.
- [x] **Vendoring decision:** do **not** redistribute Qhull. `third_party/` is
      gitignored and used only locally for fixture regeneration; the source was
      purged from git history to avoid any licensing entanglement. The build
      recipe (§4) documents how to obtain Qhull 8.0.2 and rebuild the oracle.

## 2. Public API design & freeze

The current exported surface is intentionally minimal:

- `Delaunay(x, y []float64) (triangles, neighbors [][3]int, err error)` — the
  default, Qhull/matplotlib-matched path.
- `DelaunayFast(x, y []float64) (triangles, neighbors [][3]int, err error)` — the
  robust exact-predicate baseline (no cocircular diagonal matching).

Before tagging `v0.1.0`:

- [x] **Public naming decided.** The Qhull-faithful path is the default `Delaunay`;
      the raw exact-predicate baseline is `DelaunayFast`. (In matplotlib-go the
      matched variant is what callers want, so it is the default.)
- [ ] **Add `doc.go`** with a package overview + runnable example.
- [ ] Consider a small result type (`type Triangulation struct{ Triangles, Neighbors [][3]int }`)
      vs. the current multi-return — multi-return matches the matplotlib-go call
      sites today; keep it unless we want a richer surface.
- [ ] Decide whether to expose anything else the port already computes internally
      (convex hull, the vertex creation order, Voronoi dual). Out of scope for
      v0.1.0 unless there's demand.
- [ ] Run an API audit / `go doc` review; freeze for v0.1.0 and tag.

## 3. Finish the algorithm — grid5x4 (the 60/61 holdout)

Carried over from matplotlib-go Phase 12, task "3c.6f". The faithful ridge engine
(`buildHullOrderRidge`) reaches **33/34 cocircular** exact build-order; the lone
holdout is `grid5x4`.

- [ ] **Close grid5x4 → 34/34.** Root cause (re-diagnosed via the QHATTACH oracle,
      *not* the earlier "cross-addPoint coplanarity" theory): the merged quad
      `f1=[19,15,4,0]`'s last two ridges (`[19,4]`, `[19,15]`) get swapped between
      the merge and the Qz-infinity-point visibility step, flipping which cone
      seeds the directed-partition replacement walk. Closing it requires faithful
      **non-simplicial ridge-order tracking across `addPoint`s** (intermediate
      swap-remove / re-append, plus `qh_facetintersect` cone vertex order).
      Deep, fragile, and **cosmetic only** — Qhull's cocircular diagonal is
      arbitrary, so the fallback is already a valid Delaunay triangulation. Was
      explicitly deferred; tackle here only if 34/34 is wanted for the library's
      own correctness story.
- [ ] When closed: raise the ratchets in `order_oracle_test.go`
      (`cocircularRidgeRatchet`) and `buildhull_test.go`
      (`computedCocircularRatchet`) to 34, and drop the holdout notes from README.

## 4. Ground-truth oracle: vendoring & build recipe

The oracle is the real test harness — it captures Qhull's creation order
(`introspect`), projected state (`dump_state`), and per-step merge trace
(`stepdump`, gated on `QHATTACH`/`QHSTEP`) as fixtures.

- [ ] **Add a build target** (`Makefile` or `justfile`) for the tools, e.g.
      `cc -O2 -I third_party/qhull-8.0.2/src third_party/qhull-8.0.2/introspect.c \
      third_party/qhull-8.0.2/src/libqhull_r/*.c -lm -o bin/introspect`
      (and likewise `dump_state`, `stepdump`). Mirror the recipe documented in
      `testdata/gen_creation_order.py`.
- [ ] **Capture the instrumentation patches** applied to the local
      `src/libqhull_r/*.c` (QHATTACH / QHSTEP trace printfs) as a standalone
      `.patch` against a pristine Qhull 8.0.2 tarball + a `.wrap`-style pin
      (mirroring how matplotlib pins FreeType 2.6.1), with the pristine source +
      sha pinned so the oracle is reproducible. **Note:** the full Qhull source is
      no longer committed (it was gitignored and purged from history for licensing
      reasons), so these patches currently exist **only on local disk** — a small
      own-authored `.patch` is the supported way to preserve and version them
      without redistributing Qhull itself.
- [ ] **Document fixture regeneration**: `QHULL_INTROSPECT=bin/introspect python3
      testdata/gen_creation_order.py`, and the corpus via `testdata/gen_corpus.py`.
- [ ] Note the buffering gotcha in docs (libqhull trace printfs are block-buffered;
      capture to a file, not `head`).

## 5. CI & tooling

- [ ] **GitHub Actions**: `go test ./...`, `go vet ./...`, `gofmt`/`gofumpt` check,
      `golangci-lint run`. No cgo and no system deps needed for the Go tests
      (the oracle is only for fixture regeneration, not CI), so CI is simple.
- [ ] Add a `justfile` (build / test / lint / fmt / oracle-build) consistent with
      the MeKo conventions.
- [ ] Add `golangci-lint` config; clear any lints the relocation surfaced.
- [ ] Optional: codecov / coverage badge.

## 6. Tests & quality

- [ ] Sanity-review test names now that the package is at the module root (some
      docstrings reference `third_party/qhull-8.0.2/...` paths — still valid since
      the oracle moved with the same relative layout; confirm after any reshuffle).
- [ ] Consider promoting the corpus comparators (`TestDelaunayConnectivityVsQhull`,
      `TestComputedOrderRidge`) to the documented gates.
- [ ] Add a couple of plain usage tests / examples that don't depend on the corpus,
      so the public API has standalone coverage.

## 7. Cutover in matplotlib-go (do LAST — currently intentionally NOT done)

Once §1–§5 are stable and a version is tagged:

- [ ] Add `require github.com/MeKo-Christian/qhull-go vX.Y.Z` to matplotlib-go's
      `go.mod` (use a `replace => ../qhull-go` during local development).
- [ ] Repoint the two consumers — `tri/delaunay.go` and `tri/triangulation.go` —
      from the in-tree `tri/qhull` to the external module.
- [ ] **Delete** matplotlib-go's `tri/qhull/` and its gitignored
      `third_party/qhull-8.0.2/` oracle.
- [ ] Run the full parity suite (`just test`). The Delaunay connectivity is
      unchanged by the move, so **no goldens should change**; if any do, that's a
      regression to investigate, not a regen.
- [ ] Update matplotlib-go `PLAN.md` Phase 12 to point at this repo and mark the
      extraction complete.

---

## Notes / invariants to preserve

- Parity is validated downstream by matplotlib-go's golden/reference suite, **not**
  by this repo's unit tests alone. Don't change Delaunay connectivity casually.
- Goldens (downstream) are regenerated against matplotlib, never hand-edited.
- The exact-predicate `DelaunayFast` is the cgo-free, always-correct fallback; the
  default `Delaunay` only refines the cocircular diagonal on top of it.
