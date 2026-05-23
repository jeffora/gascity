# Release Gate: Dolt Local-Only Remote Guard

Bead: ga-9usr9
Source bead: ga-d457b
Branch: deploy/ga-d457b
Reviewed branch: origin/builder/ga-d457b
Base: origin/main 8fe54229572b
Merge-base: 1b89b21bbfa8
Reviewed commit: d7b4341379c4
Gate result: PASS

## Criteria

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| 1 | Review PASS present | PASS | `bd show ga-9usr9` notes contain `Reviewer verdict: PASS`; reviewer mail `gm-wisp-d19n3u` confirms PASS for source bead `ga-d457b`, branch `origin/builder/ga-d457b`, commit `d7b434137`. |
| 2 | Acceptance criteria met | PASS | The reviewed diff is scoped to 7 files. It adds `DoltLocalOnlyRemoteCheck`, keys detection on literal `dolt.local-only: true`, exposes `ReadDoltLocalOnly`, registers the check beside `DoltBackupCheck`, removes `bd dolt push/pull/remote add` session-close guidance from `AGENTS.md`, and covers warning, OK, direct data-dir, local-backup, absent/false flag, name, and fix injection cases. Live config has `/home/jaword/projects/gascity/.beads/config.yaml` line `dolt.local-only: true`; branch `gc doctor --json` reports `rig:gascity:dolt-local-only-remote` OK with `dolt.local-only enabled; no off-box remotes registered`. |
| 3 | Tests pass | PASS | `make test` passed via `scripts/go-test-observable`: `observable go test: PASS log=/tmp/gascity-test.jsonl.IfZgvu`. `go vet ./...` passed with no output. `git diff --check origin/main...HEAD` passed with no output. |
| 4 | No high-severity review findings open | PASS | Review notes list three informational, non-blocking observations and no HIGH or blocker findings. |
| 5 | Final branch is clean | PASS | Before writing this release gate, `git status --short --branch` printed only `## deploy/ga-d457b...origin/builder/ga-d457b`; there were no uncommitted files. |
| 6 | Branch diverges cleanly from main | PASS | After refreshing `origin/main`, `git merge-tree --write-tree HEAD origin/main` succeeded. Branch shape is 1 commit ahead and 3 commits behind `origin/main`; there are no merge conflicts. |

## Validation Notes

- `git diff --stat origin/main...HEAD` shows the expected 7-file doctor,
  contract, and AGENTS.md change set.
- The full `gc doctor --json` smoke against the city context exits nonzero due
  to existing unrelated city diagnostics, but the new local-only remote check
  reports OK for the gascity rig.
- Dashboard/API checks are not required; this change does not touch
  `internal/api/`, OpenAPI schema files, dashboard code, or generated
  dashboard types.
