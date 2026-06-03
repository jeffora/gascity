# Release Gate: supervisor install binary-mismatch guard

**Deploy bead:** ga-7uzy8n
**Source bead:** ga-72abmr
**Review bead:** ga-nooirz
**Feature branch:** builder/ga-72abmr-supervisor-install-guard
**Reviewed commit:** 7314817bd0c7991f5ea26a17a88d94efd12cee26
**Base checked:** origin/main at 84b75173a96f0d328bc485fb6d250a901688350a
**Gate result:** FAIL

`docs/PROJECT_MANIFEST.md` is not present in this checkout, so this gate uses the deployer prompt's seven release criteria.

## Summary

The reviewed commit itself is scoped to the supervisor install guard:

- `cmd/gc/cmd_supervisor_lifecycle.go`
- `cmd/gc/cmd_supervisor_test.go`
- `docs/reference/cli.md`

The feature branch cannot be released as a single-bead PR in its current state. Against current `origin/main`, the branch is 136 commits ahead and 136 commits behind, and the PR diff contains 647 files across unrelated subsystems. This violates the single-feature release requirement and must go back to builder for a clean current-main branch before deploy.

## Criteria

| # | Criterion | Result | Evidence |
|---|-----------|--------|----------|
| 1 | Review PASS present | PASS | `bd show ga-nooirz` contains `Review verdict: PASS`; reviewer mail `gm-wisp-1qfw8` reports PASS for commit `7314817bd`. |
| 2 | Acceptance criteria met | PASS | Code inspection confirms the guard is present in both `installSupervisorSystemd` and `installSupervisorLaunchd`, `--force` is wired to `supervisorInstallForce`, and the stale tmp ExecStart test now opts into force. |
| 3 | Tests pass | FAIL | Deployer did not run the full release test suite because the branch failed the branch/scope gate before it could become a valid release candidate. Reviewer-reported `go vet` and `make test-fast-parallel` are noted but are not a deployer gate substitute. |
| 4 | No high-severity review findings open | PASS | Reviewer notes list two INFO findings only: malformed-unit false-positive safe path and CLI docs default nit. No HIGH findings. |
| 5 | Final branch is clean | FAIL | The feature branch worktree has an untracked `cmd/gc/cmd_supervisor_install_guard_test.go`; the detached verification worktree at `7314817bd` was clean before gate-file creation. |
| 6 | Branch diverges cleanly from main | FAIL | `git rev-list --left-right --count origin/main...HEAD` returned `136 136`. `git merge-tree --write-tree origin/main HEAD` reported no textual conflict, but the branch is not a clean current-main release branch. |
| 7 | Single feature theme | FAIL | `git diff --name-only origin/main...HEAD` lists 647 files spanning `cmd/gc`, `internal/beads`, `internal/api`, `internal/config`, `internal/benchmarks`, examples, docs, workflows, and test trees. A single-bead supervisor install guard PR cannot carry this unrelated history. |

## Additional Evidence

- `git diff --check origin/main...HEAD` fails on unrelated stale release-gate trailing whitespace from `release-gates/ga-2ql1ev-config-comments-gate.md`.
- Top changed areas in the branch diff include 173 paths under `cmd/gc`, 42 under `internal/beads`, 41 under `internal/api`, 33 under `internal/config`, and 30 under `internal/benchmarks`.

## Handoff

Builder should prepare a clean branch from current `origin/main` containing only the supervisor install binary-mismatch guard change, then reroute to deployer. The reviewed patch from `7314817bd` touches only the three supervisor/docs files listed above; do not carry the unrelated branch history into the release PR.
