# Apply Summary

Source bead: `ga-j195cx`
Apply bead: `ga-hvlofq6`
Attempt: `90`
Synthesis: `.gc/design-reviews/ga-j195cx/attempt-90/synthesis.md`
Global verdict: `block`
Result verdict: `iterate`

Updated `engdocs/design/formula-compiler-requirements.md` in place.

Addressed findings:

- Parser/caller coverage: bound `caller-preflight` generation to the repository-wide per-occurrence caller manifest, added construct presence/materialization semantics before condition filtering, required raw source attribution for every triggering construct, and covered prompt producers, JSON, conflicts, multiple constructs, convergence projections, and unregistered `gc.*` metadata.
- Operator diagnostics: specified `FormulaDiagnosticBurstBudget`, producer policies, defaults, reset semantics, rollup lifecycle states, bounded child retention, payload ownership, and fixtures for repeated scans, restarts, reloads, config flaps, many disabled-host subjects, warning snapshots, mutable counters, and CAS failure.
- Durable writers: made accepted artifacts the only durable-write authority through manifest-driven signature guards, prompt/generated-command coverage, shell-out/probe exclusion from write authorization, and all-or-nothing zero-write fixtures for every multi-write boundary.
- Rollout sequencing: added a phase dependency graph, clarified Phase 4a as shared plumbing rather than completed migration, defined allowed 4b-4g parallelism/order, and split Phase 8 requires-only conversion from Phase 9 parser alias removal with packman provenance and non-placeholder floor gates.
- Migration evidence: centralized JSON-first artifact contracts for compatibility, min-floor, first-party inventory, external support, alias-window start, active roots, alias drain, docs-check, and repair-root dry runs with owners, commands, exit behavior, and consumers.
- Convergence: added forbidden and replacement flows, compile/accept/project/validate pseudocode, same-identity reuse rules, `evaluate_prompt` hashing in artifact identity, and blocked-loop diagnostic metadata that cannot authorize writes.
- Docs and terminology: made `make formula-docs-check` a hard predecessor to user-visible diagnostics and pinned stale-guidance, doctest, generated artifact, and inventory failure classes.
- Future capabilities: made released construct rows immutable release artifacts, required latest registry-owning binary validation before first-party pack publication, separated explicit `>=1` provenance from behavioral identity, and pinned overflow/mixed-manifest old-reader behavior.

Unfixable in this apply bead:

- The minor synthesis finding about persona synthesis files being stamped under `attempt-1/persona-syntheses/` instead of the current `$ATTEMPT_DIR/persona-syntheses/` is a design-review workflow artifact routing bug, not a change to the formula compiler requirements design. This apply bead updated the design document and records the workflow defect as residual workflow follow-up.

Saved artifacts:

- `.gc/design-reviews/ga-j195cx/attempt-90/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-90/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-90/apply-summary.md`
