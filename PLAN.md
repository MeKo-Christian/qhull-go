# qhull-go — Roadmap

This repo was extracted from `matplotlib-go`'s `tri/qhull` package (a faithful
port of matplotlib's Qhull 8.0.2 Delaunay backend). The code is copied here as the
**future source of truth**; matplotlib-go still uses its own in-tree copy and is
**not yet** wired to depend on this module. This document tracks everything needed
to make qhull-go a properly published, standalone library and to complete the
cutover.

Status legend: `[ ]` todo · `[~]` in progress · `[x]` done.

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
- [x] **Add `doc.go`** with a package overview (`doc.go`) + runnable, verified
      example (`ExampleDelaunay` in `example_test.go`, with `// Output:`). The
      package comment was consolidated into `doc.go` (removed from `delaunay.go`)
      so `go doc` shows a single clean overview.
- [x] **Result type decided: keep the multi-return.** No `Triangulation` struct —
      `(triangles, neighbors, err)` matches the matplotlib-go call sites and keeps
      the surface minimal. Revisit only if a richer surface is wanted.
- [x] **Extra exposure decided: nothing else for v0.1.0.** Convex hull, vertex
      creation order, and the Voronoi dual stay internal — out of scope unless
      there's demand.
- [x] **API audit / `go doc` review done.** Exported surface is exactly
      `Delaunay` + `DelaunayFast`; `go doc`, `go vet`, `gofmt`, and the full test
      suite (incl. the example) are green.
- [ ] **Freeze + tag `v0.1.0`.** Deferred until a v0.1.0 release is explicitly
      wanted (Go's module proxy caches tags immutably). §3 (grid5x4) and §5 (CI)
      are now both settled, so the only thing holding the tag is the go-ahead. The
      API itself is frozen.

## 3. Finish the algorithm — grid5x4 (the 60/61 holdout) — DONE

Carried over from matplotlib-go Phase 12, task "3c.6f". The faithful ridge engine
(`buildHullOrderRidge`) now reaches **34/34 cocircular** exact build-order (61/61
across the combined order lock). `grid5x4` is closed.

- [x] **Close grid5x4 → 34/34.** Root cause (pinned with the QHSTEP/QHATTACH oracle
      on the real merging build — the earlier `stepdump`/`TA` runs were misleading
      because `TAn` suppresses merging): the divergence was **not** a ridge-order or
      replacement-choice bug. The replacement of the visible facet `[7,15,0]` is the
      cone `[5,15,0]` in both engines. The gap was the **partition search itself**.
      `qh_addpoint` (libqhull_r.c:246-258) switches `qh_partitionvisible` from the
      directed `qh_findbest` walk to the linear-scan `qh_findbestnew` the moment
      premerge produces a non-simplicial new facet. On the apex-`5` merge step,
      coplanar point `6` is outside both the cone triangle `[5,7,15]` (dist ~0.25)
      and the merged quad `[5,2,7,0]` (dist ~0.23); the directed walk reached the
      quad first and returned it, whereas `qh_findbestnew` scans the new-facet list
      in order and the merged-horizon quad is appended at the tail, so the cone is
      hit first. Fix: `findBestNew` (faithful `qh_findbestnew` port) + a `findbestnew`
      flag set in `partitionVisible` when any new facet is non-simplicial. No change
      to general position (27/27) or the other 33 cocircular cases.
- [x] Raised the ratchets in `order_oracle_test.go` (`cocircularRidgeRatchet`) and
      `buildhull_test.go` (`computedCocircularRatchet`) to 34, and dropped the
      holdout notes from the README.

## 4. Ground-truth oracle: vendoring & build recipe

The oracle is the real test harness — it captures Qhull's creation order
(`introspect`), projected state (`dump_state`), and per-step merge trace
(`stepdump`, gated on `QHATTACH`/`QHSTEP`) as fixtures.

- [ ] **Add a build target** (`justfile`) for the tools, e.g.
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

- [x] **GitHub Actions** (modular `workflow_call` layout, mirroring the MeKo
      `algo-dsp` convention): `.github/workflows/ci.yml` orchestrates
      `test-unit.yml` (matrix Go 1.24.x/1.25.x: `go mod verify`, `go vet`,
      `go test`), `test-lint.yml` (`golangci-lint-action@v8`, pinned v2.12.2), and
      `test-format.yml` (gofumpt via `golangci-lint fmt` + `git diff --exit-code`).
      No cgo / no system deps — the oracle is only for fixture regeneration.
- [x] Add a `justfile` (build / test / test-race / test-coverage / bench / vet /
      lint / lint-fix / fmt / check-formatted / check-tidy / ci / oracle-build /
      clean / fix), consistent with the MeKo conventions; self-contained on
      `golangci-lint` for both lint and format (no treefmt dependency).
- [x] Add `golangci-lint` config (`.golangci.yml`, v2: standard set + misspell,
      gocritic, revive; gofumpt formatter). Cleared the one surfaced lint
      (unchecked `os.Stderr.WriteString` in the debug trace). `just ci` is green.
- [ ] Optional: codecov / coverage badge. (`just test-coverage` produces the
      profile; wiring a badge/upload is deferred.)

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
