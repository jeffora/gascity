# Kwame Asante

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Major] The target-classification contract risks becoming a broad shared facade unless the active Slice 1 implementation is limited to fields proven by the read-only API query adopter. All reviewers accept the operation-specific classifier direction, but all flag that fields such as lifecycle/label state, materialization flags, stale/partial diagnostics, or later-surface details must remain provisional until a production caller proves exact need.
- [Major] The typed target-classification result currently reads like a flat optional envelope, which conflicts with the design's own rule against flat optional shared types. Claude and DeepSeek explicitly call this out, and Codex raises the same risk as a rich first-adopter schema. Production code must use tagged or per-kind result shapes, or the table must be classified as a provisional field census rather than an implementation contract.
- [Major] Universal Slice 0 still front-loads mutation-oriented evidence and contracts before the first read-only slice proves value. The reviewers agree the universal-vs-per-slice preflight split is the right control, but command appliers, mutation/destructive boundary rows, worker-boundary exceptions, and decider/command diagnostics should move to the mutating slices that consume them unless they are strictly source inventory.
- [Major] Event-log compatibility is correctly deferred in principle, but any durable-event or `TR-007` language must be reconciled with the active requirements. The current target should remain post-commit in-process events plus durable scan recovery; no outbox, event-sourced state model, or generic committed-fact payload should be implied by early slices.
- [Minor] `SCENARIO_PARITY.yaml` and historical review notes need scope boundaries. A universal scenario index and source inventory are useful, but full active parity proof should track the rows touched by the current slice, and uncopied `DESIGN_REVIEW_NOTES.md` content must not become acceptance criteria.

**Disagreements:**
- Claude and Codex both give `approve-with-risks`; DeepSeek gives `block`. My assessment is `approve-with-risks` because two reviewers believe the design text contains the right scope controls, but the risks are material enough that the next revision must tighten the contract language before implementation beads treat these schemas as active build targets.
- The reviewers differ on whether the target-classifier table may remain as a provisional upper-bound sketch. Claude allows that path if Slice 1 implements only the fields its read-only API query adopter produces; DeepSeek wants the contract itself rewritten to per-kind result structures. The shared requirement is that active shared Go/API types must not be flat optional envelopes.
- Codex is more permissive about Slice 0 breadth if `SLICE0_CONTRACT.yaml` hard-rejects behavior-moving code and future-slice compatibility promises. Claude and DeepSeek push to move more artifacts out of Universal Slice 0 entirely. The synthesis view is to keep source-complete inventories universal while moving detailed command, mutation, and diagnostic contracts to the slices that use them.

**Missing evidence:**
- The exact minimal field set the read-only API query adapter must consume to preserve current resolution and error behavior.
- A clear vocabulary-lifecycle classification for the Target Classification Contract and its kind-dependent sub-fields.
- A production shape for classifier outputs that satisfies the flat-envelope ban, or an explicit statement that the current table is not the production shared type.
- The final `SLICE0_CONTRACT.yaml` rules proving which artifacts are evidence only and which may introduce tests or helper code.
- A current, auditable source for `TR-007` or durable-event compatibility if it is still a real requirement.
- A validator rule preventing historical `DESIGN_REVIEW_NOTES.md` content from being treated as normative unless copied into the active design or requirements.

**Required changes:**
- Classify the target-classification schema as provisional except for the exact Slice 1 API-query subset, and forbid non-first-adopter fields from appearing in active Go structs, OpenAPI/generated client types, event payloads, or cross-slice contracts.
- Resolve the flat optional-envelope contradiction by specifying tagged or per-kind production result structures, or by stating that the current table is only a provisional field census and cannot be implemented as one optional struct.
- Apply the universal-vs-per-slice split at the artifact level: keep baseline evidence, source inventory, scenario indexing, and vocabulary checkpoint mechanics universal; defer command-applier details, mutating boundary contracts, worker-boundary exception details, and mutation diagnostics to the slices that consume them.
- Keep event compatibility tied to today's in-process post-commit events and durable recovery scans. Either formalize `TR-007` in `REQUIREMENTS.md` with verification criteria or state that durable event-log compatibility is not a current constraint.
- Add Slice 0 close checks that reject behavior-moving implementation, new public command APIs, future-slice exported types, generic command/event abstractions, and uncopied historical review notes as acceptance criteria.
