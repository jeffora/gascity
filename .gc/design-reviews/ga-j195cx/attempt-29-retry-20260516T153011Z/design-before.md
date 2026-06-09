# Approved Bugfix Plan for gascity#1814

## Review Target

- Canonical GitHub issue: https://github.com/gastownhall/gascity/issues/1814
- Canonical bug bead: `ga-5s2lus`
- Bugflow run: `ga-1t0h566`
- Selected path: `confirmed-bug`
- Approval gate: `fix-plan`
- Approval verdict: `approved`
- Human approval artifact: `.gc/bugs/ga-5s2lus/approvals/fix-plan.json`

Human-approved plan, recorded 2026-05-09T21:41:30Z:

> Approved fix plan: add examples/gastown root pack.toml with canonical defaults.rig.imports.gastown, avoid legacy workspace.default_rig_includes, and add focused example fixture regression coverage.

This design document is scoped to the approved implementation plan only. It does not authorize production-code changes unless implementation discovers that the fixture cannot be made valid without a narrowly justified compatibility adjustment.

## Problem Summary

Issue #1814 reports that a user following the documented Gastown example flow:

```bash
bin/gc init --from examples/gastown /tmp/gtest
bin/gc rig add /tmp/gtest-rig
bin/gc status
```

gets a registered rig with only the generic control dispatcher. The rig does not receive the Gastown rig pack binding, so rig-scoped Gastown agents are not stamped for that rig. The reporter suggested adding legacy `[workspace] default_rig_includes = ["packs/gastown"]` to `examples/gastown/city.toml`, but the investigation found that field is deprecated and the canonical V2 mechanism is root `pack.toml` `[defaults.rig.imports.<binding>]`.

## Reproduction Summary

Reported build or baseline:

- Baseline source: `https://github.com/yjspanish/gascity-src.git`
- Commit: `b1843a291ead6c4648ed0fc47174a5537b4dddf4`
- Binary: `/tmp/gascity-historical-baseline/bin/gc`
- Result: not reproduced.
- Evidence: `.gc/bugs/ga-5s2lus/runs/ga-1t0h566/lanes/reported-build-repro/output.json`
- Important caveat: `repro.log` contains an early stale detector line that said reproduced because it expected a literal `polecat` agent. The corrected structured result says the reporter fork already has a default Gastown rig include, `gc rig add` wrote `includes = ["packs/gastown"]`, and rig status listed per-rig agents including `refinery` and `witness`.

Current main:

- `origin/main`: `3a6bcea5aafe6c249a0b77cfba5b86204f1f299f`
- Result: reproduced with high confidence.
- Evidence: `.gc/bugs/ga-5s2lus/runs/ga-1t0h566/lanes/main-repro/output.json`
- Failing artifacts:
  - `.gc/bugs/ga-5s2lus/runs/ga-1t0h566/lanes/main-repro/repro.log`
  - `.gc/bugs/ga-5s2lus/runs/ga-1t0h566/lanes/main-repro/rig-block.txt`
  - `.gc/bugs/ga-5s2lus/runs/ga-1t0h566/lanes/main-repro/rig-status.stdout`
- Key failing evidence: after `gc rig add <repo>` without `--include`, the rig block contains only `[[rigs]] name = "gtest-rig"` with no `imports` or `includes`, and rig status lists only `gtest-rig/control-dispatcher`.

The current-main repro used an isolated direct CLI flow:

```bash
GC_SESSION=fake GC_BEADS=file GC_HOME=<temp> \
  bin/gc init --skip-provider-readiness --from examples/gastown <city>
bin/gc --city <city> unregister
bin/gc --city <city> start --foreground --dry-run
mkdir <rig> && (cd <rig> && git init && git commit --allow-empty -m init)
bin/gc --city <city> rig add <rig>
bin/gc --city <city> rig status gtest-rig
bin/gc --city <city> rig list
bin/gc --city <city> start --foreground --dry-run
```

`gc init` timed out after writing `city.toml` while waiting for a temporary supervisor; the lane unregistered the temp city and continued with explicit `--city`. The config, status, and dry-run artifacts remain valid for this fixture bug.

## Root Cause Evidence

The investigation packet identifies this as template drift in the real `examples/gastown` fixture:

- `examples/gastown/city.toml` imports `packs/gastown` at city scope.
- `examples/gastown` has no root `pack.toml`.
- Therefore the example has no canonical `[defaults.rig.imports.gastown]` entry.
- `gc rig add` applies default pack bindings only from root `pack.toml` `[defaults.rig.imports.*]` or the legacy `[workspace] default_rig_includes`.
- If a rig has neither `Includes` nor `Imports`, `internal/config.HasPackRigs` returns false and pack expansion skips rig-scoped pack agents.
- Existing `cmd/gc` tests already prove that `gc init --from` preserves root-pack default rig imports when the source template has a `pack.toml`, and that `gc rig add` consumes root-pack default rig imports.

Second-wave verification passed:

```bash
go test ./cmd/gc -run 'TestInitFromPreservesCopiedPackDefaultRigImportOrder|TestDoRigAdd_RootPackDefaultRigImports'
```

The remaining gap is not the command implementation. It is the real Gastown example fixture lacking the canonical default-rig import.

Design-review finding: `internal/config.LoadWithIncludes` warns when a root `pack.toml` exists and `city.toml` still declares `[imports]`. `gc start` promotes strict-eligible warnings to failures by default. Therefore the fixed fixture must not keep the Gastown import in both files.

## Approved Fix Shape

Implement the approved fixture-first fix:

1. Add `examples/gastown/pack.toml`.
2. Include a root pack definition with schema 2.
3. Include `[imports.gastown] source = "packs/gastown"` in the root pack manifest.
4. Include `[defaults.rig.imports.gastown] source = "packs/gastown"` in the root pack manifest.
5. Move the existing city-level `[imports.gastown]` declaration out of `examples/gastown/city.toml`; the root `pack.toml` is the single canonical owner for the Gastown city import and default rig import. <!-- REVIEW: added per blocker-import-ownership -->
6. Do not add or reintroduce `[workspace] default_rig_includes` in `examples/gastown/city.toml`.
7. Keep the documented bare `gc rig add <path>` flow true after the fix.
8. Add focused regression coverage in `examples/gastown/gastown_test.go` that loads the real example fixture and asserts the canonical default-rig import exists.
9. Add or extend fast command/config coverage for the copied real fixture: `gc init --from examples/gastown` must preserve the root pack, and bare `gc rig add <path>` must write a rig import binding for Gastown without `--include`. <!-- REVIEW: added per major-real-flow-verification -->

Expected `pack.toml` shape:

```toml
[pack]
name = "gastown"
schema = 2

[imports.gastown]
source = "packs/gastown"

[defaults.rig.imports.gastown]
source = "packs/gastown"
```

Implementation should preserve current example semantics. If adding `[imports.gastown]` to root `pack.toml` while leaving `[imports.gastown]` in `city.toml` creates duplicate import behavior, resolve that by keeping a single canonical import owner and proving the expanded example still contains the same city-scoped pack behavior. Do not silently drop the city-scoped pack import.

Import ownership decision: root `examples/gastown/pack.toml` owns `[imports.gastown]`; `examples/gastown/city.toml` must no longer declare it. This preserves city-scoped pack behavior because root pack imports are merged into the city definition layer before pack expansion, while avoiding the strict warning emitted when both root `pack.toml` and city-level `[imports]` are present. Implementation should prove this with a warning-free `LoadWithIncludes` or dry-run start check.

## Test Strategy

Primary test:

- Add a fast fixture/config test in `examples/gastown/gastown_test.go`.
- The test should inspect the real `examples/gastown/pack.toml` through `config.LoadRootPackDefaultRigImports` or equivalent structured config parsing.
- Assert that a bound import named `gastown` exists and has `Source == "packs/gastown"`.
- Assert the fixture does not use legacy `workspace.default_rig_includes`.
- Assert the fixture does not leave `[imports.gastown]` in `city.toml` once root `pack.toml` exists, and that loading the example does not emit the "city.toml declares [imports]" warning. <!-- REVIEW: added per blocker-import-ownership -->
- Prefer asserting the pack/default binding over checking role-specific agent names, to keep the regression focused on the SDK mechanism and avoid baking orchestration roles into implementation logic.

Required real-flow coverage:

- Add or update a fast test using the real `examples/gastown` source fixture, not a synthetic mini-fixture.
- Copy/init from the real fixture, then run the rig-add logic without an explicit include.
- Assert the resulting rig stanza has `imports.gastown.source == "packs/gastown"` and does not rely on legacy `includes`.
- Assert the fixed copied city loads without strict-eligible warnings caused by city/root duplicate imports.
- The test may use in-process helpers rather than a slow process-backed acceptance test, but it must exercise the same copied-root-pack and bare-rig-add semantics that the user sees.

Recommended verification commands:

```bash
go test ./examples/gastown
go test ./cmd/gc -run 'TestInitFromPreservesCopiedPackDefaultRigImportOrder|TestDoRigAdd_RootPackDefaultRigImports'
make test
```

Optional full process-backed user-flow coverage, only if the implementer decides the extra runtime cost is justified:

```bash
go test ./test/acceptance -run 'TestGastownSmoke_WithRig|TestInitGastown'
```

If adding a new acceptance-a test, place it next to `TestGastownSmoke_WithRig`, initialize from the real `examples/gastown`, call rig add without an explicit include, and assert the rig receives the Gastown default import or expanded rig-scoped pack agents. This is useful but not required by the approved minimal plan.

## Expected Changed-File Boundaries

Expected files:

- `examples/gastown/pack.toml`: new root pack manifest with `imports.gastown` and `defaults.rig.imports.gastown`.
- `examples/gastown/gastown_test.go`: focused regression coverage for the real fixture.
- `examples/gastown/city.toml`: remove the city-level `[imports.gastown]`; add a concise comment or cross-reference only if needed to keep the documented `gc rig add <path>` instruction and two-pack layout understandable; no legacy default-rig include. <!-- REVIEW: added per blocker-import-ownership -->
- `docs/getting-started/coming-from-gastown.md`: audit the section that tells readers which Gastown example files to inspect. Update it if it omits the new root `examples/gastown/pack.toml`, or record in the apply summary why no docs change is needed. <!-- REVIEW: added per major-operator-migration -->

Files that should not need changes:

- `cmd/gc/cmd_rig.go`
- `cmd/gc/cmd_init.go`
- `internal/config/compose.go`
- `internal/config/pack.go`
- API, dashboard, generated schema, and event registry files

If implementation requires changes outside the expected boundaries, pause and document why the existing verified mechanics were insufficient.

## Migration, Rollout, Rollback, and User Impact

Data migration:

- None. This is an example fixture change.
- Existing cities already initialized from the old example are not automatically migrated.
- Existing users can continue using the workaround `gc rig add --include packs/gastown` for old or custom cities that do not have root default-rig imports. This workaround should not be presented as required for new cities initialized from the fixed example.
- Existing cities that have not yet registered rigs should create or update their city root `pack.toml` with the canonical snippet below before running bare `gc rig add`.
- Existing cities that already registered a bare rig without Gastown imports need a rig-level repair as well as the root default for future rigs: add `[rigs.imports.gastown] source = "packs/gastown"` to the affected rig stanza, or remove and re-add that rig after adding the root pack default.

Operator migration snippet for an old city without root `pack.toml`:

```toml
[pack]
name = "gastown"
schema = 2

[imports.gastown]
source = "packs/gastown"

[defaults.rig.imports.gastown]
source = "packs/gastown"
```

If the old city already has a root `pack.toml`, add the `[imports.gastown]` and `[defaults.rig.imports.gastown]` tables there instead of creating a second file.

Rollout:

- New `gc init --from examples/gastown` copies the added root `pack.toml`.
- Subsequent bare `gc rig add <path>` should stamp the Gastown default rig import into the rig stanza.
- The fix should align the copyable example with the already-tested canonical built-in Gastown init shape.
- Operator-facing text must keep bare `gc rig add <path>` as the happy path for the fixed example and describe `--include packs/gastown` only as an old/custom-city workaround. <!-- REVIEW: added per major-operator-migration -->

Rollback:

- Revert `examples/gastown/pack.toml` and the corresponding fixture test.
- If comments were adjusted, revert those comment changes too.
- The previous explicit workaround remains available: `gc rig add <path> --include packs/gastown`.

User impact:

- Positive: new operators following the example instructions get per-rig Gastown behavior without needing an undocumented include flag.
- Low compatibility risk: no runtime code path is expected to change.
- Main implementation risk: duplicate or shifted pack imports if root `pack.toml` and `city.toml` both own the same `gastown` binding. This plan resolves the risk by making root `pack.toml` the single owner; tests must catch any remaining warning, duplicate expansion, or lost city-scoped pack behavior.
- No `gc doctor`-class check is required for this fix. The affected state is an old copy of an example fixture, and the change can be addressed with documentation plus explicit rig-stanza repair guidance rather than a new runtime diagnostic.

## Open Assumptions and Constraints

- The bad reported version was not identified in the issue snapshot; the explicit reporter fork commit appears to be a fixed reference, not the bad baseline.
- The confirmed current-main repro used dry-run start checks and config/status inspection rather than launching provider-backed agents.
- `comments.json` is empty, so no issue comments add extra constraints.
- The approval chooses canonical root-pack defaults over the reporter's legacy `workspace.default_rig_includes` suggestion.
- Preserve the zero-hardcoded-role invariant in SDK code. This plan should not add role-specific branches or role-specific behavior to Go implementation paths.
- Keep the fix minimal. This is a fixture/default coverage gap, not a reason to redesign rig add, pack expansion, or init-from semantics.
