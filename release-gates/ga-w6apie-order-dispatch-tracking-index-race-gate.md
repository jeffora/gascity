# Release Gate: order dispatch tracking index race

Bead: ga-w6apie
Source bug: ga-8x5zv4
Review bead: ga-5v7yyh
Branch: builder/ga-8x5zv4
Reviewed commit: 082f504a48bb70ae527591a9fa1a14c10b3343b2
Gate worktree: /tmp/gascity-deploy-ga-w6apie.WOQJfR
Gate date: 2026-06-08

## Manifest

`docs/PROJECT_MANIFEST.md` is absent on `origin/main` and this branch, so no
repo-specific Release Criteria section was available. This gate uses the
deployer criteria plus the repo quality gates in `AGENTS.md` and `TESTING.md`.

## Criteria

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| 1 | Review PASS present | PASS | `bd show ga-5v7yyh` contains `REVIEW VERDICT: PASS`; deploy wrapper `ga-w6apie` says reviewed and passed by `gascity/reviewer`. |
| 2 | Acceptance criteria met | PASS | `orderDispatchTrackingIndex` has `mu sync.Mutex`; `historyEntriesForStore` and `entriesForStore` lock before reading/writing `idx.entries` and `idx.errs`; source bug `ga-8x5zv4` is closed with the expected exit contract. |
| 3 | Tests pass | PASS | See test evidence below. |
| 4 | No high-severity review findings open | PASS | Review notes for `ga-5v7yyh` list correctness/test/security PASS and only one non-blocking style nit; no HIGH findings. |
| 5 | Final branch is clean | PASS | `git status --short` was empty before writing this gate; rechecked after committing the gate before push. |
| 6 | Branch diverges cleanly from main | PASS | `git merge-base --is-ancestor origin/main HEAD` returned success on the reviewed branch. |
| 7 | Single feature theme | PASS | Commit `082f504a4` changes only `cmd/gc/order_dispatch.go` and `cmd/gc/order_dispatch_test.go` for the order-dispatch tracking-index race. |

## Acceptance Evidence

- The fix is scoped to the order dispatch tracking index in `cmd/gc`.
- `cmd/gc/order_dispatch.go` adds a mutex to `orderDispatchTrackingIndex`.
- `historyEntriesForStore` and `entriesForStore` hold that mutex across the
  cache read/check/write path.
- `cmd/gc/order_dispatch_test.go` adds
  `TestOrderDispatchTrackingIndexConcurrentEntriesForStore`, which exercises
  concurrent open-entry and history-entry cache access.

## Test Evidence

| Command | Result | Summary |
|---------|--------|---------|
| `go test -race ./cmd/gc -run TestOrderDispatchTrackingIndexConcurrentEntriesForStore -count=5` | PASS | `ok github.com/gastownhall/gascity/cmd/gc 1.367s` |
| `make test-fast-parallel` | PASS | 8 fast jobs passed: `fsys-darwin-compile`, `unit-core`, and `unit-cmd-gc-1-of-6` through `unit-cmd-gc-6-of-6`. |
| `go vet ./...` | PASS | No output; exit 0. |
| `go build -o /tmp/gc-ga-w6apie ./cmd/gc` | PASS | CLI binary built successfully. |
| `/tmp/gc-ga-w6apie --help` | PASS | Help text rendered and exited 0. |

## Deploy Decision

PASS. Open a pull request for `builder/ga-8x5zv4` and route the merge request
to mayor/mpr. Do not merge from the deployer seat.
