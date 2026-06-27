# Ground-truth oracle

The fixtures in `testdata/` (`creation_order.json`, `corpus.json`) are captured
from **Qhull 8.0.2** running matplotlib's options (`qhull d Qt Qbb Qc Qz`). They
are committed, so building and testing this package needs nothing here — this
directory is only for **regenerating** those fixtures and for diffing the Go port
against Qhull's internals during development.

Qhull itself is **not redistributed** in this repository (see `../THIRD_PARTY.md`).
This directory holds only our own code:

| File | What it is |
| --- | --- |
| `instrumentation.patch` | Our `getenv`-gated trace `printf`s added to Qhull's `libqhull_r` (no Qhull source — diff context only). |
| `introspect.c` | Tool: vertex creation order + facet sets before/after `qh_triangulate`. |
| `dump_state.c` | Tool: projected state dump. |
| `stepdump.c` | Tool: per-step facet/outside-set dump via the `TA<n>` stop option. |

## Pristine source pin

```
version : qhull_r 8.0.2 (2020.2.r 2020/08/31)
tarball : https://github.com/qhull/qhull/archive/refs/tags/v8.0.2.tar.gz
sha256  : 8774e9a12c70b0180b95d6b0b563c5aa4bea8d5960c15e18ae3b6d2521d64f8b
```

## Setup (one-time)

Download the pristine source, verify it, extract it to `third_party/` (gitignored),
and apply our instrumentation:

```sh
cd "$(git rev-parse --show-toplevel)"
mkdir -p third_party && cd third_party
curl -sL -o qhull-8.0.2.tar.gz \
  https://github.com/qhull/qhull/archive/refs/tags/v8.0.2.tar.gz
echo "8774e9a12c70b0180b95d6b0b563c5aa4bea8d5960c15e18ae3b6d2521d64f8b  qhull-8.0.2.tar.gz" \
  | sha256sum -c -
tar xzf qhull-8.0.2.tar.gz                      # -> third_party/qhull-8.0.2/
cd qhull-8.0.2
git apply -p1 ../../oracle/instrumentation.patch # or: patch -p1 < ...
```

The instrumentation is inert unless its env var is set (below), so a patched tree
still builds and behaves like stock Qhull by default.

## Build the tools

```sh
just oracle-build      # compiles oracle/*.c against third_party/qhull-8.0.2 -> bin/
```

## Regenerate the fixtures

```sh
QHULL_INTROSPECT=bin/introspect python3 testdata/gen_creation_order.py
python3 testdata/gen_corpus.py        # corpus.json (uses introspect's FACETS_POST)
```

## Tracing the build (development)

The instrumentation is gated by env vars. Set one and run a tool:

| Env var | Effect |
| --- | --- |
| `QHATTACH` | Per-cone `REPL` / `CONE` / `MERGEFACET` trace from `qh_makenewfacets` & `qh_mergefacet` — the merge/replacement decisions. |
| `QHSTEP`   | Per-pick facet list with vertices + outside sets in `qh_buildhull` (the **real** merging build). |

Example — dump the real per-step state of the grid5x4 build:

```sh
{ echo 20; printf '%s\n' "0 0" "1 0" ... ; } | QHSTEP=1 bin/introspect
```

### Gotchas (learned the hard way)

- **`stepdump` / `TA<n>` suppresses merging.** Qhull's `TA<n>` stop option exits
  `qh_buildhull` before the merge-bearing steps complete, so `stepdump`'s
  intermediate facet lists look **all-simplicial** even though the real build does
  coplanar-horizon merges. Do **not** use `stepdump` to reason about merge state —
  use `introspect` with `QHSTEP=1`, which traces the genuine merging build. (This
  cost real time while closing grid5x4: the false "Qhull never merges" reading came
  straight from `stepdump`.)
- **Block buffering.** The trace `printf`s go to stdout, which is block-buffered
  when piped. Capture the **full** output to a file (or read it whole); truncating
  mid-run with `head` can drop trailing lines / SIGPIPE the process before a flush.
- **Centroid projection.** All three tools subtract the input centroid before
  lifting (matching `gen_creation_order.py`), so point ids — not coordinates — are
  what to compare against the Go engine.
