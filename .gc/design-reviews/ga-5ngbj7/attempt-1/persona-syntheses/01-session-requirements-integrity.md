# Session Requirements Integrity Reviewer

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The artifact does not conform to the requested `gc.mayor.requirements.v1` schema. Both reviewers found missing YAML front matter, missing required top-level sections, a non-schema title and status shape, and a document path/structure that does not match the Mayor plan requirements artifact.
- [Blocker] The review contract and artifact identity conflict. The file declares itself a module-local session behavior ledger, while the Mayor requirements schema explicitly excludes module-local requirements ledgers. This must be resolved before the artifact can be approved against the schema.
- [Blocker] Evidence durability is not strong enough for a reconciliation ledger. Both reviewers identified cited tests that do not resolve in the current checkout. Claude verified that several exist on `origin/main`, while Codex treated them as stale in the active checkout; either way, the document needs an explicit reference frame and live, checkable proof paths.
- [Major] The artifact's status is ambiguous for workflow gates. It uses `Status | Seed draft` in a Markdown table instead of schema-compatible front matter with one of `draft`, `questions`, `approved`, or `blocked`.
- [Major] Several evidence cells are too broad for reliable review. File-level references such as `internal/session/manager_test.go` or `cmd/gc/session_reconcile_test.go` require readers to rediscover which test function or assertion proves each behavior.
- [Major] Important user-visible behaviors are present in ledger rows but are not represented as schema-ready examples or acceptance criteria. Drain-ack behavior, recovered work, confirmed-dead workers, invalid targets, stale snapshots, and repeated controller idempotence need concrete inputs, expected outputs, and verification methods if this artifact remains under a Mayor requirements workflow.
- [Minor] The ledger is valuable as a module behavior artifact: reviewers agreed it has strong vocabulary anchoring, broad scenario coverage, and useful code/test/doc reconciliation intent. The block is on review-contract mismatch and evidence durability, not on the usefulness of the ledger itself.

**Disagreements:**
- The reviewers disagreed slightly on stale evidence. Codex treated missing cited paths in this checkout as a blocker. Claude found that the cited tests exist on `origin/main` and judged the problem as an unpinned reference-frame gap rather than fabricated evidence. My assessment: still blocking for this lane, because a requirements integrity artifact must make its evidence frame explicit enough that future readers can reproduce it.
- Claude recommends declaring this file out of scope for `gc.mayor.requirements.v1` and changing the design-review route or schema. Codex frames the decision as either rewriting to the Mayor schema or using a different schema/review contract. These are compatible: the required decision is whether this is a Mayor requirements artifact or a module-local ledger.
- Claude warns that forcing this ledger into the Mayor schema could destroy its reconciliation value. Codex emphasizes that if it remains under the Mayor workflow, its outcomes must be converted into Example Mapping and Acceptance Criteria. My assessment: preserve the ledger shape if it is intended as a module artifact, and create or route a separate Mayor requirements document when the workflow requires `gc.mayor.requirements.v1`.

**Missing evidence:**
- A declared governing schema for `internal/session/REQUIREMENTS.md`: Mayor requirements artifact, module-local behavior ledger, or a new ledger-specific schema.
- A reference-frame anchor such as `main@<sha>` for evidence citations, so stale checkout, behind checkout, and real citation rot are distinguishable.
- Live replacement paths or restored tests for cited evidence that is missing in the active checkout, including scale-from-zero, provider-health-gate, and session-progress coverage.
- Exact test function names, commands, source guards, issues, or commits for each scenario row rather than broad file-level references.
- Schema-ready examples and acceptance criteria for drain-ack, confirmed-dead sessions, invalid target rejection, recovered work after drain cancellation, stale snapshot avoidance, and repeated controller passes.
- Canonical handling for the legacy `awake` input state, which Claude found used by a scenario row but absent from the canonical vocabulary list.

**Required changes:**
- Decide and document whether `internal/session/REQUIREMENTS.md` is a Mayor requirements artifact or a module-local session behavior ledger. Do not approve it against `gc.mayor.requirements.v1` until that decision is made.
- If it is a ledger, mark it out of scope for the Mayor requirements schema and route it through a ledger-specific schema or review contract. Recommended: preserve the ledger and create a separate `<rig-root>/plans/<slug>/requirements.md` if a Mayor requirements artifact is needed.
- If it must be a Mayor requirements artifact, add schema-compliant front matter and restructure it into `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`, `Out Of Scope`, and `Open Questions`.
- Add a reference-frame anchor for evidence, preferably an explicit `main@<sha>` or equivalent, and repair or pin citations whose paths are absent in the active checkout.
- Tighten evidence rows to cite exact test functions, commands, source assertions, issues, or commits that prove each behavior.
- Add concrete examples and acceptance criteria for high-risk session outcomes if the artifact remains in the Mayor requirements workflow.
- Add `awake` to the canonical vocabulary as a recognized legacy compatibility input that maps to active, or remove the dependency on that term from the relevant scenario row.
