# Apply Summary

Attempt: 4

Global verdict: block

Applied verdict: iterate

Updated `.gc/design-review-inputs/core-gastown-pack-migration/design.md` to
turn the attempt-4 blocker and major findings into concrete design contracts.

Changes made:

- Added `Attempt 4 Review Resolution Contracts` with a source-discovered,
  machine-readable behavior manifest, generated manifest CI checks, and
  executable `test/packcompat` row-level witnesses.
- Chose the public Gastown host-Core model: Core is an auto-included host
  system pack, public Gastown does not import Core, and Gastown patches Core
  `dog` only through normal resolved-config patch semantics.
- Added pre-resolution doctor/import recovery for stale local Gastown,
  Maintenance, `../maintenance`, and older public Gastown pins before full
  config expansion can fail.
- Added a direct production loader call-site inventory and refined runtime vs.
  partial-read loader interfaces for the Core loading slice.
- Completed bootstrap extraction details: non-nil empty production
  `bootstrapAssets`, inline `fstest.MapFS` fixtures, empty
  `bootstrapManagedImportNames`, `GC_BOOTSTRAP=skip` containment, and hardcoded
  old-path test updates.
- Replaced undecided Maintenance runtime policy with concrete Core migration
  behavior for JSONL state/archive, push fields, spawn-storm ledgers,
  `GC_PACK_STATE_DIR`, and `jsonl_archive_doctor_check.go` fallbacks.
- Declared Go role de-roling in scope for SDK/Core surfaces and added scanner
  requirements for tmux themes, default scaffolding, prompt fallback, sling
  heuristics, `classifyAgentKind`, mail/nudge targets, TOML, docs, and tests.
- Fixed rollout gates so cross-pack ownership audits happen before the public
  pin, pin adoption runs current-loader compatibility only, and the
  no-Maintenance production-loader gate runs when Maintenance is actually
  removed from `requiredBuiltinPackNames`.
- Added docs/operator-DX anchoring to `docs/reference/system-packs.md`,
  version-skew diagnostics, and a wording matrix for Core, provider system
  packs, retired Maintenance, public Gastown, Core `dog`, and stale state.

Artifacts:

- Design after: `.gc/design-reviews/ga-2404qu/attempt-4/design-after.md`
- Diff: `.gc/design-reviews/ga-2404qu/attempt-4/design.diff`
- Synthesis: `.gc/design-reviews/ga-2404qu/attempt-4/synthesis.md`

Residual status:

The design must iterate because the global verdict was `block`. The updated
design records the required choices and gates, but the workflow should rerun
review to verify that reviewers accept the new contracts.
