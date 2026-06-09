# Design Review Synthesis

## Overall Verdict: block

The persona review is unanimous across Claude and Codex that `internal/session/REQUIREMENTS.md` is useful as a session behavior ledger, but it cannot be approved against `gc.mayor.requirements.v1`. The remaining block is primarily a review-contract and artifact-identity mismatch: the schema asks for a Mayor plan `requirements.md`, while the reviewed file explicitly declares itself a module-local reconciliation ledger.

## Consensus Strengths
- Both reviewers praised the document as a valuable session behavior source of truth with clear reconciliation rules for keeping code, tests, and requirements aligned.
- Both reviewers found the lifecycle vocabulary strongly anchored to `ProjectLifecycle`, reducing the risk of invented session states or ad hoc state interpretation.
- Both reviewers noted broad coverage of high-risk session surfaces: lifecycle projection, state transitions, identity resolution, worker-boundary behavior, work release, drain safety, runtime submission, and observation.
- Both reviewers agreed the document avoids workflow-launch metadata and implementation-plan scratchpad content, preserving a clean local ledger shape.

## Critical Findings

### [Blocker] Wrong Artifact For The Requested Mayor Requirements Schema
**Sources:** Session Requirements Integrity Reviewer; Claude; Codex
**Actionability:** workflow-defect
**Issue:** The review contract points `gc.mayor.requirements.v1` at `internal/session/REQUIREMENTS.md`, but the schema describes a Mayor plan artifact at `<rig-root>/plans/<plan-slug>/requirements.md` and explicitly says it is not a module-local requirements ledger. The reviewed file declares itself a reconciliation ledger stored beside `internal/session`, so the schema and artifact identity conflict before content quality can be judged.
**Required change:** Decide the governing artifact type before another design-doc edit. If this is a module ledger, route it through a ledger-specific schema or no Mayor output schema. If the workflow needs a Mayor requirements artifact, create or review a separate plan-level `requirements.md` that conforms to `gc.mayor.requirements.v1`.

### [Blocker] Mayor Output Shape Is Absent From The Reviewed File
**Sources:** Session Requirements Integrity Reviewer; Claude; Codex
**Actionability:** workflow-defect
**Issue:** The file lacks the schema-required YAML front matter, `phase: requirements`, schema-compatible `status`, required timestamps, and exact top-level sections: `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`, `Out Of Scope`, and `Open Questions`. These are output-shape mismatches, not evidence that the ledger content is poor.
**Required change:** If the owner intentionally converts this artifact into a Mayor requirements document, add the required front matter and sections and use `draft`, `questions`, `approved`, or `blocked` status. Otherwise, do not iterate the ledger to satisfy this schema; fix the review route or provide the correct plan-level artifact.

### [Major] Evidence Is Not Reproducible From The Active Checkout
**Sources:** Session Requirements Integrity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** Codex found cited paths missing in the active checkout, including `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`. Claude verified those tests on `origin/main` and identified the local checkout as hundreds of commits behind, but the document does not pin an evidence frame, so future readers cannot distinguish a stale citation from a stale checkout.
**Required change:** Provide a reproducible evidence frame, such as `main@<sha>`, and repair or pin citations whose proof paths are absent in the active checkout. This needs external proof inventory and test path verification, not just another blind requirements edit.

### [Major] Evidence Granularity Is Too Broad For Durable Review
**Sources:** Session Requirements Integrity Reviewer; Codex; Claude
**Actionability:** external-prerequisite
**Issue:** Many scenario rows cite broad files or commits instead of exact test functions, commands, source guards, issues, or commits that prove each behavior. Reviewers had to rediscover which assertions support multi-part claims such as drain recovery, confirmed-dead work release, provider health, and progress-aware reconciliation.
**Required change:** Inventory each scenario row against live proof and replace broad citations with exact enforcing tests, commands, or source assertions.

### [Minor] Ledger Vocabulary And Wording Cleanup
**Sources:** Session Requirements Integrity Reviewer; Claude
**Actionability:** document-fixable
**Issue:** `SESSION-LIFE-001` uses `state=awake` as a legacy compatibility input, but the canonical vocabulary section does not name `awake` as a legacy input mapping to active. Claude also found `SESSION-STATE-001` says "any non-none state -> closed", while the reducer behavior is better stated as "any non-closed state -> closed".
**Required change:** Add `awake` as a recognized legacy compatibility input that maps to active, and align the close-transition wording with the reducer's non-closed phrasing.

## Disagreements
- Claude and Codex differed on stale evidence severity. Codex treated missing test paths in the active checkout as direct stale evidence; Claude found the tests on `origin/main` and treated the problem as an unpinned reference-frame gap. Assessment: this is still blocking for an evidence ledger because the document must make proof reproducible from an explicit frame.
- Claude recommended declaring the file out of scope for `gc.mayor.requirements.v1`; Codex framed the decision as either rewriting to the Mayor schema or using a different schema. Assessment: these are compatible, but the recommended path is to preserve the module ledger and fix the review route or create a separate Mayor requirements artifact.
- Claude warned that forcing the ledger into the Mayor schema would destroy its reconciliation value, while Codex emphasized the required Example Mapping and Acceptance Criteria if the Mayor route remains. Assessment: the workflow should not request another edit of this ledger until the artifact identity decision is made.

## Missing Evidence
- A declared governing schema for `internal/session/REQUIREMENTS.md`: Mayor requirements artifact, module-local behavior ledger, or a new ledger-specific schema.
- A reproducible evidence frame, such as `main@<sha>`, for all scenario citations.
- Live replacement paths, restored tests, or pinned commits for scale-from-zero, provider-health-gate, and session-progress evidence.
- Exact test function names, commands, source guards, issues, or commits for broad evidence cells.
- Schema-ready Mayor examples and acceptance criteria if, and only if, the owner decides this should become a Mayor requirements artifact.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: blocked
- Reason: Another automatic edit to `internal/session/REQUIREMENTS.md` will not settle whether the Mayor schema should govern a module-local ledger, and the schema itself says wrong-path or wrong-schema artifacts should block rather than iterate. The evidence issues also need a verified external proof frame before the document can be made durable.
- Next non-design work: decide the governing artifact/schema, update the design-review route or provide a plan-level Mayor requirements artifact, and produce a live evidence inventory for the ledger citations.

## Recommended Changes
1. Preserve `internal/session/REQUIREMENTS.md` as a module-local behavior ledger unless a human explicitly decides to convert it, and stop applying `gc.mayor.requirements.v1` to this path.
2. If Mayor requirements are needed, create or review a separate `<rig-root>/plans/<plan-slug>/requirements.md` with the required front matter and sections.
3. Define or select a ledger-specific review schema for module-local behavior ledgers.
4. Pin the ledger's evidence frame, preferably with a `main@<sha>` reference, and verify all cited tests and commits against that frame.
5. Replace broad evidence cells with exact test functions, commands, source assertions, issues, or commits.
6. Add `awake` as a legacy compatibility input and align the close-transition wording with the reducer.
