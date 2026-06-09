# Faisal Khoury - Claude

**Verdict:** approve-with-risks

Lane: doctor diagnostics; import-state warnings; safe remediation; operator
messaging. Reviewed `plans/core-gastown-pack-migration/requirements.md` (updated
2026-06-09T17:23Z) against my mandate and, per AC1, against
`assets/skills/mayor/requirements.schema.md`. None of my three red flags fire,
and this revision closed two of my prior risks: AC11 now makes the
condition-code registry "pack-independent Go-side" with schema rendering and the
exit-code matrix able to "run with no packs resolved" (a corrupt Core can no
longer suppress its own diagnostic), and AC10 plus its verification now cover
"interrupted or concurrent repairs." AC3 now also binds `import-state` to the
shared registry. The two carryover risks below are pre-implementation
sharpenings, not blockers — the doc has strong defense-in-depth (report-only
default, durable preflight/journal/backup, non-destructive-to-unrelated-TOML,
refusal semantics, live-state evidence from direct observation).

**Top strengths:**
- **Diagnostics name the exact source, from one pack-independent substrate.**
  AC11 (line 117) requires the exact resolved config source or nested import
  chain for every missing-Core/duplicate-Core/version-skew/retired-path/cache-integrity/public-pack
  condition, with source chain and file/import key/line-column "when available,"
  stable codes from a pack-independent Go-side registry, and a bootstrap-only
  mode that renders "with no packs resolved." AC3 (line 109) binds init, doctor,
  import-state, CLI, and runtime to the same registry and source-attribution
  model for the same broken state. Answers lane Q1 and forecloses my
  "generic diagnostics" red flag.
- **The mutating repair surface is bounded, fail-closed, and crash-safe.** AC10
  (line 116) keeps repair report-only by default; mutation is confined to
  `gc doctor --fix --non-interactive` and must be idempotent, atomic, resumable,
  post-verified, and non-destructive to unrelated TOML, guarded by durable
  preflight/journal/backup or refusal semantics, with offline/cache-incomplete
  repair failing *before* mutation, read-only/transitive cases reported rather
  than edited, and refusal evidence drawn from direct runtime/process/session
  observation (not stale status files). AC11 forbids hiding errors by
  substituting defaults and pins stdout/stderr separation and an exit-code
  matrix. Answers lane Q2 and forecloses my "unsafe/unwired fix" red flag.
- **Required/optional/retired messaging is enforced wherever operators read.**
  AC12 requires doctor, import-state, help, docs, and the public pack to
  consistently describe Core as required, Gastown as external/optional, and the
  Maintenance pack as retired, backed by a terminology matrix separating the
  retired Maintenance *pack* from store/supervisor "maintenance" language; AC5
  states Maintenance is not "presented as an implicit dependency." Answers lane
  Q3 and forecloses my "Maintenance is implicit / blurred" red flag.

**Critical risks:**
- **[Minor] The missing-Core `--fix` mutation target is still undefined
  (carryover).** Core is a host-materialized required system layer (AC2/AC3),
  but the negative-path missing-Core row (line 91) only promises "the explicit
  idempotent repair action" without saying what it mutates. "Omits Core"
  conflates two safety-distinct cases: (a) a corrupt/absent *materialized* Core
  tree — safe to idempotently re-materialize from the embed — versus (b) a
  user/root/rig/transitive config that drops or shadows Core — which must be
  report-only and never auto-edit operator TOML. Left merged, a reader could
  implement `--fix` as injecting a system pack into user config, which is exactly
  my red flag "fix commands mutate custom config unsafely." AC10's
  read-only/local-modified/transitive report-only rule mitigates this, but the
  missing-Core path is not explicitly mapped onto it.
- **[Minor] Multi-condition aggregation and partial-fix outcomes are
  unspecified (carryover).** AC11 frames "a ... condition" singularly, yet the
  common upgrade hits missing-Core + retired-Maintenance + version-skew at once.
  Operators running one non-interactive pass need every detected condition in
  that pass, and `--fix` must distinguish fully-repaired vs
  partially-repaired-with-enumerated-manual-steps vs nothing-fixable via the
  exit-code matrix. Neither all-conditions-per-run nor partial-fix surfacing is
  required, which weakens operator messaging for the multi-fault upgrade case.
- **[Minor] Cross-command messaging consistency is bound but not verified by an
  equality check.** AC3 now requires doctor and import-state to share one
  registry and attribution model, but the verification columns provide
  per-command goldens (AC11), not an explicit "same broken city → identical
  condition code, severity, and source attribution from both commands" check.
  For my lane (consistent operator messaging across doctor AND import-state),
  that equality should be pinned so the two surfaces cannot drift.

**Missing evidence:**
- What the missing-Core `--fix` concretely mutates (re-materialize the Core
  system tree vs. edit config vs. refuse), and the supported outcome when the
  omission lives in a read-only or transitive source.
- Whether `gc doctor`/`gc import-state` report all detected conditions per run,
  and how partial `--fix` success is surfaced to non-interactive automation
  (exit code plus a machine-readable remaining-actions list).
- AC11's binding diagnostic contract `migration-diagnostics.schema.json` does
  not yet exist (the `support/` dir currently holds only
  `maintenance-asset-classification.md`). This is expected and gated — the doc
  blocks implementation approval until the named support artifacts exist and
  pass — but AC11's per-field adequacy (live-state, pre-fix, attempted-fix,
  post-fix-verification, journal/backup, manual-reconciliation fields) is
  unverifiable until the schema is written.

**Required changes:**
- Make the missing-Core repair concrete and safe: split "Core system payload
  absent/corrupt → idempotent re-materialize from the embed" from
  "user/root/rig/transitive config omits or shadows Core → report-only, never
  auto-edited," consistent with the duplicate-Core fail-closed row and AC10's
  read-only/transitive rule. Add an Example Mapping row or extend AC10/AC11.
- Require doctor/import-state to report all detected conditions in a single
  non-interactive run, and require `--fix` to distinguish fully-repaired /
  partially-repaired (manual steps remaining, enumerated) / nothing-fixable via
  the documented exit-code matrix.
- Add a cross-command consistency verification to AC3/AC11: for one fixed broken
  city state, doctor and import-state must emit the same condition code,
  severity, and source attribution.

**Questions:**
- For missing Core, is the canonical repair always re-materialization of the
  host system pack, or may `--fix` ever edit a config to reintroduce Core — and
  when the omission is in a read-only/transitive source, is
  report-with-manual-guidance the only supported outcome?
- Does `gc doctor` aggregate multiple co-occurring conditions in one pass, and
  how is partial `--fix` success surfaced to automation (exit code plus a
  machine-readable remaining-actions list)?
