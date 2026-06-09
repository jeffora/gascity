# Morgan Schema Refactor Parity Reviewer

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The workflow is reviewing the wrong artifact for the requested schema. Both sources agree that `internal/session/DESIGN.md` is a module-local design reference, while the attached `gc.mayor.implementation-plan.v1` schema requires a Mayor `implementation-plan.md` with YAML front matter and exact top-level sections. Under the current gate this is `blocked:wrong-artifact`; the right fix is to retarget or replace the artifact, not reshape the module-local design doc into a Mayor plan.
- [Major] Target-classification parity is not traceable enough for decomposition. The design describes the right high-level resolver behavior, but it does not cite the concrete `internal/session/REQUIREMENTS.md` scenario rows such as `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-008`, and `SESSION-ID-009`, nor does it identify nearby rows that are intentionally out of scope.
- [Major] Caller-surface parity proof is underspecified. Both reviewers ask for an inventory of actual API, CLI, worker, mail, graphroute, sling, and reconciler paths that currently resolve or classify session targets, plus the characterization tests that pin each wrapper's behavior before migration.
- [Major] The first operation-specific API boundary is not precise enough. The document rejects a broad `SessionService`, but it does not define whether the first slice adds a pure classifier, caller adapter, command wrapper, or replacement for existing `ResolveSessionID` call sites. This leaves room for caller-specific allow/deny policy to leak into `internal/session`.
- [Major] Boundary ownership around close and work release needs to be explicit. Claude flags that backlog text could move caller-owned work-release policy into a session-owned close decider; Codex similarly asks for a checkpoint preventing the classifier from absorbing work assignment, mail delivery policy, provider execution, pool scaling, or generic lifecycle orchestration.

**Disagreements:**
- There is no material verdict disagreement: Claude and Codex both block on the artifact/schema mismatch.
- Claude emphasizes the schema's explicit instruction to stop as wrong-artifact and highlights API orphan rejection plus close/work-release ownership. Codex emphasizes the missing Mayor front matter/sections, the need to name the first implementation caller, and the exact test plan. These are complementary findings.
- Both reviewers approve of the engineering direction in the module doc: narrow target classification, no broad service facade, no mail/provider/pool policy movement, and behavior aligned with the current resolver. The block is about routing/schema and missing parity proof, not about rejecting the refactor concept.

**Missing evidence:**
- Required Mayor front matter: `plan_slug`, `phase: implementation-plan`, `rig`, `rig_root`, `artifact_root`, `requirements_file`, `status`, `created_at`, and `updated_at`.
- Required schema sections: `## Summary`, `## Current System`, `## Proposed Implementation`, `## Data And State`, `## Testing`, `## Rollout And Recovery`, and `## Open Questions`.
- A scenario-row matrix mapping the target-classification slice to `internal/session/REQUIREMENTS.md` rows and explicitly marking adjacent rows in or out of scope.
- A caller-surface inventory naming the files and wrapper behaviors to preserve, including API orphan rejection and API/mail factory-target rejection.
- Named characterization tests and exact package/command targets proving resolver and caller behavior before and after the first moved caller.
- A concrete first-slice API shape that keeps caller-specific allow-closed, orphan-rejection, work-release, mail, provider, pool, and reconciler policy out of `internal/session`.
- An enforceable checkpoint or stated review gate that prevents the narrow classifier from becoming a universal session service.

**Required changes:**
- Resolve the artifact-type mismatch before any content iteration. Either provide a schema-conforming Mayor `implementation-plan.md` for this workflow, or change the review workflow/schema if the intended target is `internal/session/DESIGN.md`.
- Do not retrofit `internal/session/DESIGN.md` into the Mayor schema shape; preserve it as a module-local reference document if it remains useful.
- In the correct implementation plan, add `Current System` coverage for the resolver code, requirements rows, current tests, and every caller surface that re-derives target resolution or adds wrapper policy.
- In `Proposed Implementation`, define the first narrow API/adapter precisely and state that caller-specific allow/deny policy remains outside the core session classifier.
- In `Testing`, list the exact requirements rows, existing characterization tests, missing tests to add, and commands/package targets that prove parity after each caller move.
- In `Rollout And Recovery`, state whether any durable session metadata changes are in scope. If the first slice is pure classification, say that explicitly and define rollback as reverting the caller move.
- Disambiguate close versus work-release ownership: the session-owned close decider may emit close/lifecycle facts, while work release remains caller-owned unless a separate approved design changes that boundary.
