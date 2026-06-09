# Kwame Asante

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash/Gemini slot

**Consensus findings:**
- [Blocker] The active vocabulary contract still mixes next-slice vocabulary with future-slice vocabulary. All sources flagged ambiguity around `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` appearing beside Slice 1 `TargetCandidate`/`TargetSelection` material. Future names must remain design-only until an owning slice has a first delegated caller and exact field demand.
- [Blocker] The Slice 0 vocabulary gate is not executable as written. The design requires `TestVocabularyCheckpoints` to have nonzero matches, but also forbids checkpoint rows before first-caller evidence. The sources converge on the same resolution: Slice 0 must seed baseline rows from existing contracts, while new vocabulary still requires first-caller evidence.
- [Blocker] `TestVocabularyCheckpoints` needs concrete fail-first mechanics. Claude called out that a YAML well-formedness test could pass vacuously; Codex asked for unused-field rejection around `TargetCandidate`/`TargetSelection`; the DeepSeek/Gemini slot tied the same ambiguity to premature downstream checkpointing.
- [Blocker] Slice 1 target classification still risks becoming a universal resolver. Codex and DeepSeek/Gemini flagged the breadth of the proposed all-surface candidate collector; Claude flagged that tests are not a valid first delegated caller. The first Slice 1 bead must name one production adapter, the candidate kinds it needs, the fields deliberately excluded, and the untouched legacy surfaces.
- [Blocker] The proposed target result shape needs stronger anti-envelope constraints. DeepSeek/Gemini judged the flat `TargetCandidate`/`TargetSelection` shape to be a flat optional envelope; Codex and Claude did not require a full split, but their first-caller and excluded-field requirements support the same guard. A tagged/per-kind or private minimal structure should be required before any shared shape is approved.
- [Major] Event-log deferral is directionally correct, but `SessionFactEvent` remains a vocabulary backdoor while listed as an active checkpoint. Claude and Codex accept durable-scan-first recovery and thin current events; DeepSeek/Gemini blocks on the generic `committed facts` field. The synthesis accepts the recovery model but requires future event vocabulary to be provisional and to forbid generic committed-facts payloads.
- [Major] Raw target classification must not own adapter policy. DeepSeek/Gemini uniquely flagged `ordinary-config-target` and `requires-materialization` as policy-shaped negative kinds. This is accepted as in-lane because it directly protects the design's stated classifier/adaptor boundary.
- [Minor] Backlog scope needs one more bound. Claude flagged concurrent mutation-owning slice overlap; Codex separately stressed that Slice 0 must remain artifact-only. The design should prevent broad Slice 0 implementation and make overlap of bake-period slices explicit.

**Disagreements:**
- Claude and Codex gave `approve-with-risks`; the DeepSeek/Gemini slot gave `block`. My assessment is `block` because the shared blockers are contract-level ambiguities that can cause premature exported types, checkpoint rows, event payload fields, or broad resolver code before decomposition starts.
- Claude and Codex view TR-007/event recovery as sufficiently deferred; DeepSeek/Gemini treats `SessionFactEvent.committed facts` as a current blocker. I assess the durable-scan-first event strategy as acceptable, but only if `SessionFactEvent` is moved out of active checkpoints and generic committed facts are explicitly forbidden.
- DeepSeek/Gemini requires splitting `TargetCandidate`; Claude and Codex frame the issue as first-caller and field-boundary discipline. I do not require a specific type spelling, but the design must mechanically prevent a flat optional envelope through private first-use types, tagged/per-kind structs, or equivalent field-exclusion checkpoints.
- Only DeepSeek/Gemini flagged raw classifier negative kinds as policy leakage, and only Claude flagged open-slice concurrency. Both are accepted as required or residual changes because they are direct YAGNI/scope-control risks.

**Missing evidence:**
- Whether `VOCABULARY_CHECKPOINTS.yaml` starts with baseline rows for the existing contracts listed in the design, including their current files and callers.
- The exact assertion and fixture list for `TestVocabularyCheckpoints`, including failure cases for missing checkpoint rows, undeclared fields, expansion without an adopting caller, and unused fields introduced for later surfaces.
- The exact first production caller for Slice 1 target classification, including the API/Huma handler or helper, candidate kinds, policy fields, and parity fixtures it owns.
- The package visibility rule for new Slice 1 target classification types, especially what counts as "leaving a private slice package."
- Whether raw classifier outputs will be limited to physical store identity/type facts, with materialization and rejection policy owned by surface adapters.
- Whether mutation-owning slices may overlap during bake periods, and how the shrink-only ledger converges if they do.

**Required changes:**
- Split the vocabulary material into active allowed vocabulary for the next slice and design-only future bounds. Move `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` out of active checkpoints, with an explicit no-code/no-checkpoint/no-generated-field rule until their owning slices start.
- State that Slice 0 seeds `VOCABULARY_CHECKPOINTS.yaml` with the existing-contract rows as baseline active vocabulary, while new rows still require first-delegated-caller evidence.
- Define `TestVocabularyCheckpoints` as a fail-first gate with fixtures for missing rows, undeclared fields, provisional-bound expansion without an adopting caller, and fields included only for later surfaces.
- Add a Slice 1 first-caller constraint: the first target-classification bead may implement only one adopting production surface, its required candidate kinds, its required policy/output fields, and explicit excluded fields; all other surfaces remain legacy until their own adoption beads.
- Replace or constrain the flat target candidate shape so it cannot become a universal optional envelope. Acceptable forms include private first-use types, tagged/per-kind candidate structs, or equivalent checkpoint enforcement that rejects unused fields.
- Remove generic `committed facts` from future event vocabulary and require each event payload to carry only the slice-specific fields demanded by immediate typed subscribers.
- Keep raw target classification policy-free: collect physical identity/type facts only, and move `ordinary-config-target`, `requires-materialization`, and similar rejection/materialization decisions into surface-specific adapters.
- Keep Slice 0 artifact-only and add an explicit concurrency bound or overlap policy for open mutation-owning slices.
