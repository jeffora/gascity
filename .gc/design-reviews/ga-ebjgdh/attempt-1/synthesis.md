# Design Review Synthesis

## Overall Verdict: block

The review blocks because the workflow is gating `internal/session/DESIGN.md` against the non-empty Mayor implementation-plan schema, and both model reviewers identify that as the wrong artifact. The session refactor direction is broadly sound, but the current artifact cannot satisfy this schema and the correct implementation-plan artifact still needs traceability, caller inventory, operation-boundary detail, and test proof before decomposition.

## Consensus Strengths
- Both Claude and Codex praised the narrow refactor scope around session target classification instead of a broad `SessionService`, generic command bus, event-sourcing rewrite, or mail/provider/pool policy move.
- Both reviewers found the described resolver behavior aligned with the current core contract: direct session bead ID, open exact `session_name`, open exact current `alias`, caller-specific allow-closed behavior, and no ordinary config-name or `template:<name>` fallback.
- Both reviewers agreed that the document's boundary intent keeps work assignment, mail delivery, provider execution, pool scaling, and reconciler policy outside `internal/session`.

## Critical Findings

### [Blocker] Wrong artifact for implementation-plan schema
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** workflow-defect
**Issue:** The workflow root reviews `internal/session/DESIGN.md` while the attached schema is `gc.mayor.implementation-plan.v1`, which requires a Mayor plan artifact at `<rig-root>/plans/<plan-slug>/implementation-plan.md`. The schema explicitly says a module-local reference document is the wrong artifact and that this case should stop as wrong-artifact rather than iterate the document.
**Required change:** Retarget the workflow/schema if the intended review target is `internal/session/DESIGN.md`, or provide a schema-conforming Mayor `implementation-plan.md` and point the review at it. Do not retrofit `internal/session/DESIGN.md` into the Mayor plan shape.

### [Major] Required implementation-plan output shape is absent
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** The reviewed artifact lacks the required YAML front matter and the required top-level sections `Summary`, `Current System`, `Proposed Implementation`, `Data And State`, `Testing`, `Rollout And Recovery`, and `Open Questions`. These are schema conformance failures caused by the artifact mismatch, not proof that the module design content should be reshaped in place.
**Required change:** Create or supply the correct `implementation-plan.md` with the schema-required front matter and exact section order before rerunning this schema gate.

### [Major] Target-classification parity is not traceable enough
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** The design describes the right high-level resolver behavior, but it does not map the first slice to concrete `internal/session/REQUIREMENTS.md` rows such as `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-008`, and `SESSION-ID-009`, nor does it mark adjacent named-session rows as in or out of scope.
**Required change:** In the correct implementation plan, add a scenario-row matrix that names the exact requirements rows and global invariants the target-classification slice preserves, plus nearby rows intentionally excluded from the slice.

### [Major] Caller-surface parity proof is underspecified
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** The reviewers found no complete inventory of caller surfaces that currently resolve or classify session targets, and no named characterization tests for each wrapper behavior. The missing surfaces include API, CLI, worker, mail, graphroute, sling, dispatch, and reconciler paths, with API orphan rejection called out as a distinct wrapper policy.
**Required change:** Add a caller-surface inventory naming the relevant files, current wrapper behavior, existing tests, missing characterization tests, and exact proof commands for each caller moved behind the session-owned classifier.

### [Major] First operation-specific API boundary is too loose
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** The design rejects a broad service facade, but it does not precisely define whether the first implementation slice adds a pure classifier, caller adapter, command wrapper, or replacement for existing `ResolveSessionID` call sites. That leaves room for caller-specific allow/deny semantics to leak into `internal/session`.
**Required change:** Define the first narrow API or adapter shape in the correct plan, including what it returns and which caller-owned policies remain outside it, especially allow-closed handling, orphan rejection, mail recipient policy, provider execution, pool scaling, and work assignment.

### [Major] Close and work-release ownership is ambiguous
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude
**Actionability:** external-prerequisite
**Issue:** Backlog language around close and work-release recovery could authorize moving caller-owned work-release policy into a session-owned close decider, contradicting the stated boundary that work assignment and release stay outside `internal/session`.
**Required change:** State that a session-owned close decider may emit close or lifecycle facts, while work release remains caller-owned unless a separate approved design changes that boundary.

### [Minor] Anti-SessionService checkpoint is not enforceable
**Sources:** Morgan Schema Refactor Parity Reviewer; Claude; Codex
**Actionability:** external-prerequisite
**Issue:** The no-broad-service constraint is currently prose only. Reviewers asked for either a mechanical guard or an explicit acknowledgement that reviewer discipline is the only control.
**Required change:** Add a concrete checkpoint, such as a lint/import test or review gate, that prevents the classifier from growing into a universal session service, or explicitly document that no mechanical guard exists.

## Disagreements
- There is no material verdict disagreement: Claude and Codex both return `block` because the implementation-plan schema is attached to a module-local design document.
- Claude gives more detail on the schema's wrong-artifact rule, API orphan rejection, and close/work-release ownership. Codex emphasizes the missing Mayor front matter, exact implementation-plan sections, first caller selection, and proof commands. These are complementary, not conflicting.
- Both reviewers approve of the engineering direction as a module-local refactor design; the block is about artifact/schema routing and missing decomposition proof, not rejection of the refactor concept.

## Missing Evidence
- Required Mayor front matter: `plan_slug`, `phase: implementation-plan`, `rig`, `rig_root`, `artifact_root`, `requirements_file`, `status`, `created_at`, and `updated_at`.
- Required schema sections: `## Summary`, `## Current System`, `## Proposed Implementation`, `## Data And State`, `## Testing`, `## Rollout And Recovery`, and `## Open Questions`.
- A scenario-row matrix mapping the target-classification slice to `internal/session/REQUIREMENTS.md` rows and explicitly marking adjacent rows in or out of scope.
- A caller-surface inventory naming files and wrapper behavior for API, CLI, worker, mail, graphroute, sling, dispatch, and reconciler paths.
- Named characterization tests and exact commands proving resolver and caller behavior before and after the first moved caller.
- A concrete first-slice API or adapter contract that keeps caller-specific allow-closed, orphan-rejection, work-release, mail, provider, pool, and reconciler policy out of `internal/session`.
- An enforceable checkpoint, or an explicit no-mechanical-guard statement, for preventing a narrow classifier from becoming a universal session service.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: blocked
- Reason: Another edit to `internal/session/DESIGN.md` cannot make this review pass while the workflow is applying the Mayor implementation-plan schema to a module-local design document. The content gaps are real, but they should be addressed in the correct implementation-plan artifact or after the workflow/schema target is corrected.
- Next non-design work: Retarget this review to a schema appropriate for `internal/session/DESIGN.md`, or author and review a schema-conforming Mayor `implementation-plan.md`; build the caller inventory and characterization-test proof list needed by that plan.

## Recommended Changes
1. Resolve the artifact/schema mismatch: either review `internal/session/DESIGN.md` without the Mayor implementation-plan schema, or provide a correct `<rig-root>/plans/<plan-slug>/implementation-plan.md`.
2. In the correct implementation plan, add the required YAML front matter and exact schema sections, especially `Current System`, `Proposed Implementation`, `Testing`, and `Rollout And Recovery`.
3. Add a target-classification requirements matrix covering `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-008`, `SESSION-ID-009`, relevant global invariants, and intentionally excluded nearby rows.
4. Inventory every caller surface that currently resolves or classifies session targets, including API orphan rejection and API/mail factory-target rejection, and tie each to characterization tests and proof commands.
5. Define the first narrow classifier or adapter boundary precisely, including what remains caller-owned.
6. Clarify that session close facts and caller-owned work release are separate responsibilities.
7. Add or document the checkpoint that keeps the classifier from becoming a broad `SessionService`.
