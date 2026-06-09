# Mira Acharya - Claude

**Verdict:** approve-with-risks

Lane: invariant tests; absence scans; regression harnesses; acceptance
traceability. Reviewed `plans/core-gastown-pack-migration/requirements.md`
(`status: draft`, `updated_at 2026-06-09T17:23:58Z`, `Open Questions: None`)
against `gc.mayor.requirements.v1` and the live test surface. Schema conformance
is clean from my vantage: correct front matter (`phase: requirements`, no
downstream `*_file` keys, no bead IDs), six sections in required order,
happy/negative/edge coverage, `Open Questions: None`, and every AC carries a
behavior-focused Verification column. The named `support/*.yaml|*.json` paths
read as acceptance-evidence contracts, not implementation file assignments. I
confirmed the support artifacts do not exist yet (only
`support/maintenance-asset-classification.md` is present) — but the doc
deliberately sequences them as implementation-approval gates (Problem Statement;
Out Of Scope), so their absence at the requirements phase is by design, not a
defect. Adjacent matters (resolution-matrix internals, doctor exit-code matrix,
rollout pins) are left to their reviewers.

**Top strengths:**
- The anti-cheating test contract is unusually strong and directly neutralizes
  my red flags. AC13 requires coverage-transfer validators to "parse active
  execution evidence such as `go test -json`, prove each mapped witness actually
  runs, include negative/regression controls that fail when moved behavior is
  broken, fail closed on skipped, empty, or no-op witnesses, and reject
  post-deletion walks that pass only because no assertions ran." That is the
  precise antidote to "tests check only file presence instead of loaded
  behavior."
- Acceptance traceability has a single executable backbone. AC17's
  `acceptance-proof-matrix.yaml` maps AC1–AC16 to concrete commands/tests/gates,
  names gate placement (local/pre-commit/integration/deterministic public-pack
  CI/live release/manual), requires "at least one executable gate … before
  decomposition," and "fails validation when any AC lacks evidence or when
  external evidence is referenced but unavailable." This makes lane-question-1
  (every AC maps to a real proof) enforceable rather than aspirational.
- Role-neutrality is invariant-driven, not vigilance-driven. AC8 mandates an
  absence-scan validator with denied-token matching, **positive and negative
  controls**, an **allowlist rot guard**, substituted-executor fixtures, and a
  no-executor proof; AC9 adds config tests for default/renamed/omitted/undefined/
  disabled `dog` plus a Go/source/generated scan for literal `dog` fallback.
  Together these answer lane-question-2 (fail if Core reintroduces role names,
  with documented exceptions and configurable `dog` defaults).

**Critical risks:**
- **[Major] Absence/behavior scans are not pinned to *freshly generated* output,
  so a stale committed snapshot can produce a false pass.** AC8 requires
  "generated/materialized output coverage" and AC7 requires the Gastown pack to
  actually "load, resolve, render, trigger, route, notify, run scripts," but
  neither AC states that the gate must **regenerate/materialize the artifacts
  inside the gate and scan the live output**. A scan over checked-in generated
  files (`.gc/system/packs/*`, synthetic caches, generated TS types) would pass
  even if regeneration reintroduced a role token or a Maintenance import — the
  exact "passes because nothing actually ran" failure this lane exists to catch.
  The doc should require the absence-scan and behavior gates to produce their
  inputs by fresh generation/materialization, with a control proving that a
  token planted into a *regenerated* artifact fails the gate.
- **[Major] The "active execution evidence / reject no-op witness" guarantee is
  scoped to AC13 only and is not generalized to the other invariant harnesses.**
  AC8 (role-neutrality) and AC14 (public Gastown validation) carry controls and
  fallback-disabled checks but do **not** carry AC13's "prove each mapped witness
  actually runs … fail closed on skipped, empty, or no-op witnesses" guarantee.
  An absence scan or a public-pack load check that silently runs zero assertions
  (skipped fixture, empty matrix, disabled job) would pass. Require the
  execution-evidence / no-skip / no-no-op guard to apply to every invariant and
  release gate, not just coverage-transfer.
- **[Minor] AC13's frozen baseline is "frozen historical baseline *or*
  digest-verified snapshot" — the non-digest branch can drift.** The
  coverage-transfer denominator is only trustworthy if it is immutable. A
  baseline that is "frozen" by convention but not digest-pinned can be edited so
  that retired assertions silently disappear from the denominator. Require the
  baseline to be digest-pinned (the AC6 ledger already mandates a digest-frozen
  source snapshot; AC13 should reuse that discipline).
- **[Minor] No drift guard binding the AC set to the AC17 proof-matrix.** AC17
  fails when an AC lacks evidence, but nothing requires the matrix validator to
  fail when the AC list changes (an AC added/renumbered or a Verification column
  edited) without a corresponding matrix row update. Without an AC↔matrix
  bidirectional completeness check, the traceability backbone can quietly fall
  out of sync. (AC6/AC7 already require bidirectional links for behavior rows;
  the proof matrix deserves the same.)

**Missing evidence:**
- No frozen-baseline commit/digest is named for AC13's "retired test file,
  assertion, fixture, and behavior witness" denominator. This is correctly an
  implementation-phase artifact, but the requirements leave the *kind* of
  immutability ambiguous (see Minor #1).
- The doc does not state which existing tests assert implicit Maintenance
  inclusion. That inventory is implementation work, but the live tree already
  has clear candidates (`internal/builtinpacks/registry_test.go`,
  `internal/packman/bundled_test.go`, `cmd/gc/embed_builtin_packs_test.go`,
  `cmd/gc/import_state_doctor_check_test.go`); AC13's frozen baseline must be
  proven to capture them rather than a hand-picked subset.

**Required changes:**
- Amend AC7 and AC8 verification to require that generated/materialized/rendered
  artifacts are produced by **fresh generation inside the gate** and scanned
  live, with an explicit positive control: a planted role token (and a planted
  retired-Maintenance import) in a regenerated artifact must fail the gate.
- Generalize AC13's execution-evidence contract — "parse `go test -json` (or
  equivalent), prove the witness ran, fail closed on skipped/empty/no-op" — to
  AC8 (absence scan), AC9 (executor binding tests), and AC14 (public-pack load/
  trigger checks), so no invariant gate can pass by running zero assertions.
- Require AC13's coverage-transfer baseline to be **digest-pinned** (not merely
  "frozen"), reusing the AC6 source-snapshot digest discipline, and require the
  baseline to be proven complete against the live retired-test set rather than a
  curated list.
- Add an AC17 (or AC13) requirement for an **AC↔proof-matrix completeness
  guard** that fails when any AC has no matrix row or any matrix row references a
  nonexistent AC, so traceability cannot drift as ACs evolve.

**Questions:**
- For AC8/AC7, is the intent that the absence/behavior gates regenerate inputs
  from source on every run, or may they scan checked-in generated artifacts? The
  answer determines whether a stale-snapshot false pass is possible.
- Should the "fail on skipped/empty/no-op witness" guarantee be a single shared
  harness primitive reused across AC8/AC9/AC13/AC14, or per-AC bespoke checks? A
  shared primitive is the only way to keep the guarantee from eroding as new
  gates are added during decomposition.
- Is the AC13 frozen baseline expected to be a git commit/tag, a content digest
  manifest, or both? Only a content-addressed form makes the coverage-transfer
  denominator tamper-evident.
