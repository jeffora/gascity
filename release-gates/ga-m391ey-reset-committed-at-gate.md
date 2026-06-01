# Release Gate: ga-m391ey - reset_committed_at restart handoff

Deploy bead: `ga-m391ey` - `needs-deploy: persist reset_committed_at on restart request`
Source review bead: `ga-7071hn`
Reviewed commit: `64aa34ed0` on `builder/ga-2znrco.3-beadmail-assignees`
Release branch: `release/ga-m391ey-reset-committed-at-v11`
Release commit before gate: `9d751d392f608a16005e8e5872202f95bbe7798c`
Base: `origin/main` at `882955678403fc4327ac57422bbbf668ac0231de`
Evaluated: `2026-06-01T04:25:06Z`

`docs/PROJECT_MANIFEST.md` is not present in this checkout, so this gate uses
the release criteria from the active deployer instructions plus the repository
testing guidance in `TESTING.md`.

## Scope

The builder branch was stacked with unrelated coordstore and assignee work. This
release branch was cut fresh from current `origin/main` and cherry-picked only
the reviewed reset handoff commit. The resulting code diff before this gate file
contains exactly:

- `cmd/gc/session_reconciler.go`
- `internal/session/lifecycle_transition.go`
- `internal/session/lifecycle_transition_test.go`

## Criteria

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| 1 | Review PASS present | PASS | `ga-7071hn` is closed with `REVIEWER PASS` for `64aa34ed0`; `ga-m391ey` records the reviewer PASS handoff and deploy request. |
| 2 | Acceptance criteria met | PASS | `RestartRequestPatch` now accepts a clock value, stamps `reset_committed_at` in UTC RFC3339 format, and the reconciler passes `clk.Now()`. `PreWakePatch` does not clear this diagnostic timestamp, verified by `TestPreWakePatchPreservesResetCommittedAt`. |
| 3 | Tests pass | PASS | Focused `internal/session` tests passed; restart-request reconciler tests passed with `GC_FAST_UNIT=0`; `make test-fast-parallel` passed all fast shards; `go vet ./...` exited clean. |
| 4 | No high-severity review findings open | PASS | The review notes on `ga-7071hn` list no security issues and only a non-blocking observation about retaining diagnostic metadata after successful restart. Unresolved HIGH count is 0. |
| 5 | Final branch is clean | PASS | `git status --short` was empty after the cherry-pick and before writing this release-gate file. This file is the only deployer-authored addition and is committed with the gate. |
| 6 | Branch diverges cleanly from main | PASS | Branch was created from current `origin/main`; `git merge-base origin/main HEAD` equals `882955678403fc4327ac57422bbbf668ac0231de`. `git merge-tree --write-tree HEAD origin/main` exited 0 and produced tree `57d5feeb28ef7d40d6f7d44fee37f9c187537e41`. |
| 7 | Single feature theme | PASS | The release branch touches one subsystem theme: session restart handoff metadata used by reset-stall diagnostics. The diff is limited to the lifecycle patch, its reconciler call site, and tests. |

## Acceptance Evidence

- `git diff --stat origin/main...HEAD` before this gate showed only the three
  expected files and 32 insertions / 4 deletions.
- `git diff --check origin/main...HEAD` exited 0.
- `rg -n "RestartRequestPatch\(" cmd internal` found one production call site,
  updated to `sessionpkg.RestartRequestPatch(newSessionKey, clk.Now())`, plus
  the function definition and tests.
- `TestLifecycleTransitionPatchesSetCompleteMetadata` covers both restart patch
  shapes with a fixed non-UTC time and expects UTC RFC3339 output.
- `TestPreWakePatchPreservesResetCommittedAt` proves the timestamp survives the
  fresh-wake reset path.

## Commands

```text
gh auth status
git fetch origin main
git worktree add -b release/ga-m391ey-reset-committed-at-v11 /tmp/gascity-deploy-ga-m391ey-v11 origin/main
git cherry-pick 64aa34ed0
git status --short --branch
git diff --stat origin/main...HEAD
git diff --check origin/main...HEAD
rg -n "RestartRequestPatch\(" cmd internal
git merge-tree --write-tree HEAD origin/main
GOTOOLCHAIN=auto go test ./internal/session -count=1
GOTOOLCHAIN=auto GC_FAST_UNIT=0 go test ./cmd/gc -run 'TestReconcileSessionBeads_RestartRequest|TestReconcileAndWake_RestartRequest|TestReconcileSessionBeads_PreservedRunningNamedSessionHonorsRestartRequest|TestReconcileSessionBeads_BeadMetadataRestartRequestedWhenSessionDead' -count=1
GOTOOLCHAIN=auto make test-fast-parallel
GOTOOLCHAIN=auto go vet ./...
git config core.hooksPath
```

## Test Summary

```text
go test ./internal/session -count=1
ok  	github.com/gastownhall/gascity/internal/session	7.191s

GC_FAST_UNIT=0 go test ./cmd/gc -run 'TestReconcileSessionBeads_RestartRequest|TestReconcileAndWake_RestartRequest|TestReconcileSessionBeads_PreservedRunningNamedSessionHonorsRestartRequest|TestReconcileSessionBeads_BeadMetadataRestartRequestedWhenSessionDead' -count=1
ok  	github.com/gastownhall/gascity/cmd/gc	8.349s

make test-fast-parallel
All fast jobs passed

go vet ./...
clean

git config core.hooksPath
.githooks
```
