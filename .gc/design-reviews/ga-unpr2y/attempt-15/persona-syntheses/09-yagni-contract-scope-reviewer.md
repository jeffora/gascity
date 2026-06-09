# Kwame Asante

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Slice 0 has re-inflated into an all-future-slices gate. Codex and DeepSeek block because the first behavior extraction is a read-only API query classifier, yet Slice 0 requires detailed artifacts for wake, close, retire, drain, runtime start, rollback, provider side effects, event emission, repair/backfill, and broad boundary matrices before that narrow read path proves value. Claude rates the same issue major rather than blocker, but agrees the 13-artifact preflight and full taxonomy are a serious regression from the reset.
- [Blocker] The target-classification active schema includes speculative future surfaces before a first caller proves the fields. All reviewers flag that only `api-query` is the first adopter, while the design already enumerates mail, extmsg, assignee, nudge, attach, transcript, CLI, and api-command surfaces as if they were active vocabulary. This contradicts the document's own Vocabulary Lifecycle rule that future terms stay provisional until a production caller proves exact need.
- [Blocker] The candidate shape is a flat optional envelope, which the design itself forbids. Claude and DeepSeek cite the candidate table with many kind-conditional fields; Codex raises the same active-schema risk. This recreates the `SessionFacts` facade problem inside the first classifier contract unless the table is explicitly provisional or production code uses tagged per-kind structures.
- [Major] Diagnostics and route inventories risk becoming speculative registries. Reviewers agree source-complete inventory is useful, but active diagnostic codes, operation IDs, event relationships, renderer coverage, subprocess budgets, and hot-loop constraints should be seeded from existing behavior and the next adopting slice only. Future operations should not reserve active vocabulary before delegation.
- [Major] The distinction between source inventory and active compatibility contract is unclear. A full route/source inventory can be universal Slice 0 evidence, but full parity contracts should be required only for the first adopting route and then grown per slice.
- [Major] `DESIGN_REVIEW_NOTES.md` is correctly described as historical in the design, but the execution rules should also forbid validators or implementation beads from treating uncopied archived notes as acceptance criteria.
- [Major] The design contains good YAGNI controls but does not apply them to its new contracts. All reviewers praise the Vocabulary Lifecycle, event-sourcing deferral, anti-`SessionService` stance, and one-caller-at-a-time refactor rules; the block is that the Target Classification Contract and Slice 0 gate bypass those controls.

**Disagreements:**
- Claude gives `approve-with-risks`, while Codex and DeepSeek give `block`. My assessment is `block` because the overbroad Slice 0 gate and active multi-surface taxonomy are not peripheral cleanup; they define what implementers will build first and can directly undo the reset's YAGNI posture.
- Claude would accept keeping the Target Classification Contract as a provisional upper-bound sketch if the design says Slice 1 implements only the `api-query` subset. Codex and DeepSeek push harder to move non-first-adopter surfaces out of active enums and generated/shared fields. Either is acceptable if active Go/API vocabulary remains limited to the first delegating caller.
- Reviewers differ on how much of `COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and `DIAGNOSTICS_MANIFEST.yaml` can remain in universal Slice 0. The synthesis view: source and writer inventories can stay universal; detailed command contracts, active diagnostics, and mutation-specific boundaries should move to per-slice entry gates.
- DeepSeek calls the candidate table a blocker and asks for tagged per-kind structures; Claude permits treating the table as a provisional field census. The required invariant is that production shared types must not be flat optional envelopes.

**Missing evidence:**
- A split between universal Slice 0 evidence and per-slice preflight contracts.
- A declaration that the Target Classification Contract is provisional upper-bound material except for the exact `api-query` first-adopter subset.
- The minimal active kind, field, and source-surface set required by the read-only API query adopter.
- A binding rule that non-first-adopter surfaces must not appear in active Go code, exported types, generated API fields, event payloads, or diagnostics until their own delegating slice.
- A production shape for classifier outputs that avoids flat optional envelopes, or an explicit statement that the current candidate table is not a production shared type.
- A diagnostic-manifest activation rule that distinguishes existing behavior and the next adopting slice from future provisional codes.
- A validator/implementation rule that uncopied `DESIGN_REVIEW_NOTES.md` content is rationale only, not acceptance criteria.

**Required changes:**
- Split Slice 0 into universal evidence and per-slice preflight. Keep baseline, validator harness, source inventory, scenario parity index, mutation-writer inventory, route/source inventory, and vocabulary checkpoint mechanism universal. Move detailed command-applier rows, mutation boundary contracts, runtime/provider side-effect details, repair/backfill contracts, and mutation diagnostics to the slices that actually adopt them.
- Restrict the active target-classifier contract to the first read-only API query adopter. Mark non-`api-query` surfaces and unused kinds/fields as provisional examples or deferred rows, and state they must not generate active code or wire/event vocabulary in Slice 1.
- Resolve the flat optional-envelope contradiction by using tagged per-kind result structures for production, or by explicitly classifying the candidate table as a provisional census with a rule that production shared types must be per-kind/tagged.
- Scope `DIAGNOSTICS_MANIFEST.yaml` to existing behavior plus the next adopting slice. New reason/outcome codes should become active only when a concrete caller delegates and renderer tests exist.
- Separate source-complete inventories from contract-complete parity proof. Slice 0 should prove hidden routes and writers are discovered; only the first adopting route needs full active parity fixtures before Slice 1.
- Add an explicit rule that `DESIGN_REVIEW_NOTES.md` may be cited as rationale only. Any rule that governs implementation or validation must be copied into the active design or approved artifact in current minimal language.
