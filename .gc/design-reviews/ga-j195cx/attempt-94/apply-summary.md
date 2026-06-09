# Apply Summary

Source bead: `ga-j195cx`
Apply bead: `ga-nahw9m7`
Attempt: `94`
Synthesis: `.gc/design-reviews/ga-j195cx/attempt-94/synthesis.md`
Global verdict: `block`
Result verdict: `iterate`

Updated `engdocs/design/formula-compiler-requirements.md` in place.

Addressed blocker and major findings:

- Canonical design artifact: restored the missing configured design path from the latest complete reviewed artifact and added a fail-closed review/snapshot contract so automation cannot silently review `engdocs/proposals/formula-migration.md` or another context file.
- Compatibility boundary: added an attempt-94 closure contract requiring unknown axes, future minima, malformed values, `contract` conflicts, disabled host capability, legacy-only roots, `GC_NATIVE_FORMULA=false`, and `bd` probes to fail before durable writes unless explicitly validation-only.
- Caller and runtime inventory: made the per-occurrence manifest, `CompileResult` to `AcceptedCompileArtifact` path, shared workflow-root predicate, prompt/template coverage, generated TypeScript/API coverage, and legacy `bd` materialization coverage explicit.
- Rollout sequencing: reinforced dual declarations, `[pack] requires_gc` floor enforcement, old-reader fixtures, public notice, JSON release evidence, Phase 8 requires-only conversion, and Phase 9 alias removal as separate gates.
- Parser and validation matrix: reinforced raw capture, byte-exact grammar, TOML/JSON shape rows, duplicate handling, unsupported future capability rows, construct detection before filtering, caller-path zero-write fixtures, count locks, and diagnostic ordering.
- Diagnostic projection: reinforced explicit `HostCapabilities`, typed CLI/Huma/OpenAPI/dashboard/Event Bus/release projections, registered typed event payloads, and append-only-safe grouped diagnostics.
- Deprecation evidence: corrected stale external-support references so canonical release evidence is `docs/release/formula-compiler-external-support.json`, not a Markdown checklist, and clarified JSON schemas/owners/commands/exit contracts as gate inputs.
- Pack provenance and author tooling: reinforced resolver/lockfile provenance prerequisites for refs, SHAs, content hashes, dirty state, transitive sources, and `[pack] requires_gc`, with fail-closed behavior when evidence is unavailable.
- Documentation and terminology: reinforced the glossary, docs/examples inventory, generated help/API/dashboard gates, PackV2 docs, and stale-guidance CI as same-branch requirements for visible surfaces.

Residual workflow note:

- The minor synthesis finding about persona syntheses landing outside the current attempt directory is a design-review workflow artifact propagation issue. This apply step records the current design-after snapshot, design diff, and summary under attempt 94; the workflow routing defect should be fixed in the design-review workflow, not in the formula compiler requirements design.

Saved artifacts:

- `.gc/design-reviews/ga-j195cx/attempt-94/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-94/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-94/apply-summary.md`
