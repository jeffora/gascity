# Camille Okafor - Claude

**Verdict:** approve-with-risks

Lane: existing-city upgrade, legacy local paths, compatibility shims, version
skew, two-repo rollout. Reviewed the current
`plans/core-gastown-pack-migration/requirements.md` (`updated_at`
2026-06-09T17:23:58Z) against `requirements.schema.md`. This revision resolved
two prior [Major]s from this lane: a successful existing-city upgrade is now a
first-class **happy-path** row (line 90, with operator-readable success output,
post-fix resolved-config proof, journal/backup evidence, and no-retired-load
verification), and AC15's pin-coherence agreement set now binds **repair output**
alongside fresh-init (line 121), closing the intra-release migration-vs-fresh
pin-drift gap. The remaining findings are narrower and concentrate on the
runtime boot-vs-fail boundary and shim-lifetime scoping.

**Top strengths:**
- Two-repo rollout is bounded against a flag-day release. AC14 (line 120) records
  the two-repository release order so Gas City never ships a public pin lacking
  the validated Gastown behavior manifest, and AC15 (line 121) forces
  pin-ledger, version-skew-matrix, compat proof, lock/cache provenance,
  docs/registry/repair output, and fresh-init to agree on
  source/subpath/immutable-commit/digests, with branch/tag refs demoted to
  fetchability metadata only. Each cross-version cell (old-binary/new-pack,
  new-binary/old-lock, offline, downgrade, rollback) is classified
  supported-diagnostic / unsupported-fail-closed / repair-required rather than
  requiring synchronized releases — red flag #3 (version drift / flag-day) is
  structurally closed.
- The existing-city upgrade machinery is unusually complete and now has its
  success definition. AC10 (line 116) + AC11 (line 117) require report-only by
  default, `gc doctor --fix --non-interactive` as the sole mutating surface
  (idempotent / atomic / resumable / post-verified / preflight- and
  journal/backup-guarded), recursive source attribution with
  file/import-key/line-column, a bootstrap-only diagnostic mode that runs with no
  packs resolved, in-flight session refusal from direct runtime observation, and
  old-binary post-marker reconciliation. The happy-path row (line 90) supplies
  the operator-visible success signal that was previously missing.
- No-silent-fallback posture is consistent, addressing red flag #2's runtime
  form. Retired Maintenance is "neither auto-included nor silently materialized"
  (How, line 79; AC5, line 111); stale `.gc/system/packs/maintenance` is
  ignored/pruned/reported and "never loaded as an active Maintenance pack" (line
  99); offline resolution "never falls back to in-tree examples, system packs, or
  bundled synthetic aliases" (line 96, AC16 line 122); diamond/same-name ties
  "never select an embedded alias, stale cache, or in-tree example as a tie
  breaker" (line 100).

**Critical risks:**
- **[Major] The failure mode for a *present, declared* retired import during a
  normal (non-doctor) command in the pre-`doctor --fix` window is unspecified.**
  Missing *Core* explicitly fails closed (line 63), and stale *materialized
  directories* are explicitly non-fatal ignored state (line 99). But a retired
  optional import edge still written in `city.toml`/`pack.toml`/a transitive
  config sits between those, and the doc only specifies its *diagnostic* and
  *repair* behavior (negative path line 93), never its runtime gating. The "When"
  dimension (line 76) acknowledges that otherwise "operator workflows fail from
  retired pack paths," but does not require that failure to be a *clean*
  fail-closed with a stable condition code pointing at `gc doctor --fix` — so the
  most common upgrade trigger (an operator simply running their normal workflow on
  an un-migrated city) could surface as a clean actionable stop, a silent drop of
  the Gastown import, or an arbitrary downstream error. The silent-drop branch
  directly violates the Problem Statement promise that "Gastown operators must not
  lose supported behavior when Gastown behavior moves to the public pack
  authority" (line 47). A prior pass of this lane rated this Minor; I weight it
  Major because (a) the degrade branch strips promised Gastown behavior, and (b)
  the doc asserts "Open Questions: None" (line 148) while this boot-vs-fail
  contract is a product decision an implementer would otherwise guess.
- **[Minor] "Compatibility-pin windows are explicit and temporary" (line 98) has
  no named closure trigger or post-activation inertness check.** AC15 (line 121)
  requires the version-skew-matrix to *classify* compatibility-pin states, but a
  state can stay classified "supported diagnostic" with nothing demanding the
  window ever close. The in-tree path's inertness is separately guarded by AC4's
  non-resolvability isolation and AC5/AC8 closure/role-neutrality scans, which
  mitigates the permanent-alternate-source risk for the *path*; the gap is
  specifically the *pin window's* expiry trigger. Worth naming so the temporary
  guarantee is enforceable rather than aspirational.
- **[Minor] The "compatibility shim" disposition for the legacy
  `internal/bootstrap/packs/core` root (Problem Statement, lines 33-39) is
  under-scoped and asymmetric with AC4.** AC4 (line 110) demands the retired
  in-tree *Gastown and Maintenance* roots be "deleted or isolated as
  non-resolvable fixtures excluded from runtime resolution" with explicit
  absence/non-resolvability checks, but the legacy *Core* root — the most likely
  second active Core source — gets only AC2's "classified as non-runtime fixture"
  treatment (line 108) with no symmetric runtime-exclusion test and no sunset.
  "Non-runtime compatibility shim" is also self-contradictory; left unscoped it is
  the textbook shape of red flag #2.
- **[Minor] Version-skew (AC15, line 121) is scoped to the public Gastown pin and
  is silent on bundled-Core-vs-stale-locked-Core skew across a binary upgrade.** A
  city whose lock/cache pins an old Core digest while the new binary bundles a
  newer Core payload is only implicitly covered by AC3's generic "stale
  materialized copies" / "conflicting pins fail closed" (line 109). Whether the
  bundled Core payload always supersedes a stale locked/materialized Core digest
  on upgrade (expected) or fails closed as a conflict is not stated, even though
  Core is a bundled system layer, not a remote.

**Missing evidence:**
- The required runtime contract for a *present, declared* retired optional import
  hit by a normal behavior-changing command in the pre-repair window: clean
  fail-closed-with-condition-code-and-pointer vs. silent-drop vs. arbitrary
  downstream failure.
- The decided closure/expiry trigger for the compatibility-pin window and a check
  proving the compatibility/in-tree path is inert once the activation pin is
  consumed.
- An explicit bundled-Core-vs-stale-locked-Core precedence rule at binary-upgrade
  time, and a symmetric runtime-non-resolvability check for
  `internal/bootstrap/packs/core` (parallel to AC4's for the Gastown/Maintenance
  roots).
- Whether `gc doctor --fix --non-interactive` for a Gastown city itself performs
  the network fetch + digest-verify of the pinned public pack, or only rewrites
  config and relies on a separately-seeded cache. AC10 lists "air-gapped cache
  seeding" / "install failures" and AC16 covers runtime offline behavior, but the
  repair-step sequencing for an offline operator's migration is not stated, so an
  offline operator's upgrade outcome is undetermined.

**Required changes:** (product decisions, to land before requirements approval)
- State the runtime-resolution contract for a present, declared retired optional
  import in the pre-`doctor --fix` window. Recommend: fail closed before
  behavior-changing operations with a stable condition code that names the
  introducing source and points at `gc doctor --fix`, mirroring missing-Core, and
  explicitly distinguish this from a never-declared absent optional pack. If
  genuinely undecided, move it to Open Questions and adjust `status`, since as
  written it makes "Open Questions: None" untrue for a product-behavior decision
  on the upgrade path. Extend the negative-path row (line 93) with the
  boot-vs-fail evidence.
- Tighten AC15 to require the version-skew-matrix define the compatibility-pin
  window's closure/expiry trigger and a post-activation inertness check, so the
  "temporary" guarantee is provable.
- Add a symmetric runtime-non-resolvability check for the legacy Core root to AC4
  (or AC2) and either drop the "compatibility shim" disposition in favor of
  "deleted / non-runtime fixture", or require an explicit sunset gate and owner if
  any runtime grace shim is genuinely intended.
- State the bundled-Core-vs-stale-locked-Core precedence at binary-upgrade time in
  AC3/AC15 (expected: bundled payload wins; stale materialized/locked Core is
  pruned or reported as retired state, never loaded).
- State whether migration repair fetches+verifies the pinned public pack or relies
  on a pre-seeded cache, and extend the offline edge case (line 96) or AC10 to
  cover the repair step itself for an air-gapped operator.

**Questions:**
- For an existing Gastown city on a new binary but not yet repaired: does a normal
  `gc` workflow fail closed (with a condition code and a doctor pointer) on the
  dangling retired import, or run with Gastown behavior silently dropped?
- What event closes the compatibility-pin window, and is there a check proving the
  compatibility/in-tree source path is no longer selectable once the activation
  pin is consumed?
- Does the bundled Core payload always supersede a stale locked/materialized Core
  digest on binary upgrade, or can that combination fail closed?
- For an offline/air-gapped Gastown city, does migration repair perform the
  network fetch of the pinned public pack, or require a separate seeding step
  first?

**Schema conformance (`gc.mayor.requirements.v1`, schema non-empty):** Conforms.
Front matter carries only allowed keys with `phase: requirements` and `status:
draft` (no `requirements_file` / `implementation_plan_file` / `design_file`, no
bead IDs or formula targets). The six required body sections are present and in
order (Problem Statement, W6H, Example Mapping, Acceptance Criteria, Out Of
Scope, Open Questions); W6H covers all seven dimensions (lines 72-80); Example
Mapping now has happy + negative + edge rows including the previously-missing
existing-city-upgrade happy path (line 90); ACs are testable; the `support/*.yaml`
references are acceptance-evidence contracts, not implementation file
assignments, so they do not breach the iteration rules; Out Of Scope names real
scope creep; Open Questions is a bare `None` (line 148). The one coherence note
against AC1's own readiness bar: the [Major] boot-vs-fail item above is an
unresolved product decision an implementer would have to infer, so per the
schema's Implementation-Plan-Readiness rule it should be resolved inline (or moved
to Open Questions) rather than left implicit before requirements approval.
