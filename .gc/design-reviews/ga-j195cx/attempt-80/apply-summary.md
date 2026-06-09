# Apply Summary

Verdict source: `.gc/design-reviews/ga-j195cx/attempt-80/synthesis.md`

Global verdict was `block`, so this apply pass addresses blocker and major
findings in `engdocs/design/formula-compiler-requirements.md` and sets the
workflow verdict to `iterate`.

Changes applied:

- Added host source line/column to typed diagnostics, event payload fields, and
  parity fixtures so direct CLI, API-routed CLI, Huma JSON, generated
  TypeScript, dashboard, and Event Bus projections share the same remediation
  surface.
- Defined the bounded operator rollup for host/config diagnostics, including
  grouping keys, child detail fields, mutable producer-owned cadence state,
  restart/reload behavior, and burst-budget fixtures.
- Made the caller inventory repository-wide across Go packages, first-party
  packs, examples, tutorials, docs snippets, generated prompt templates, and
  `gc bd mol ...` command references, with narrow allowlists and deletion
  tests for retired helpers.
- Locked rollout sequencing: Phase 2 now carries first-party dual declarations,
  green inventory, docs/check bundle, generated surfaces, and proposal
  supersession before visible diagnostics; Phase 3 is dormant floor metadata;
  Phase 4 producer migration is blocked by a writer-lockdown gate.
- Completed parser/construct-registry executability by separating authored
  `step.expand` from compose traversal, requiring checked coverage intent ids,
  literal count locks after generation, and registry minimum/capability
  invariants.
- Defined the external-support JSON schema, malformed-row behavior,
  public-notice and unknown/unreachable expiry semantics,
  `formula.migration.pin_pack_revision`, SHA-pinned edit targets, and
  `[pack] requires_gc` comparator evidence consumption.
- Strengthened convergence preflight for create, retry, next iteration, manual
  iterate, speculative/fallback pour, missing-state repair, missing-child
  repair, pending-wisp adoption, and pending-wisp burn, plus active-loop
  blocked status, event/status payload, retry semantics, and pending-wisp
  cleanup.
- Added host capability and `RequirementSource` authority guards with property
  tests for `CompilerCapability(0)` non-escape and same-identity host downgrade
  write-intent limits.

Unfixable blocker/major items: none documented; all synthesis blocker and
major findings were addressed by design text additions.

Artifacts saved:

- `.gc/design-reviews/ga-j195cx/attempt-80/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-80/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-80/apply-summary.md`
