# Morgan Schema Refactor Parity Reviewer - Codex

**Verdict:** block

**Top strengths:**
- The document keeps the refactor narrow and explicitly rejects a broad `SessionService`, generic command bus, event-sourcing rewrite, or movement of mail/provider/pool policy into `internal/session`.
- The first slice is correctly aimed at session target resolution, which is already represented by `internal/session/resolve.go` and covered by `internal/session/resolve_test.go`.
- The behavior summary matches the key targeting invariant in `internal/session/REQUIREMENTS.md`: direct session bead ID, open exact `session_name`, open exact current `alias`, and no template/config-name fallback.

**Critical risks:**
- [Blocker] The reviewed artifact is the wrong artifact for the requested output schema. The workflow root declares `design_review.design_doc=internal/session/DESIGN.md` and `design_review.output_schema=/data/projects/gascity-packs-worktrees/gc-plan-pack/gascity/assets/skills/mayor/implementation-plan.schema.md`, but the schema requires a Mayor `implementation-plan.md` artifact with YAML front matter and exact top-level sections: `Summary`, `Current System`, `Proposed Implementation`, `Data And State`, `Testing`, `Rollout And Recovery`, and `Open Questions`. `internal/session/DESIGN.md` is explicitly a module-local design reference and has none of that required front matter or section shape, so this review should stop as `blocked:wrong-artifact` rather than iterate the module document.
- [Major] The first target-classification slice is not traceable to concrete `REQUIREMENTS.md` scenario rows. The design gestures at the right behaviors, but does not name `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-008`, `SESSION-ID-009`, or any excluded rows. Without that mapping, decomposition can move caller logic while missing the exact API, mail, and allow-closed semantics that must remain stable.
- [Major] Caller parity proof is underspecified. The document says "several surfaces already re-derive it" and mentions API, CLI, worker, and reconciler code, but it does not inventory the actual caller paths or tests that must be preserved before moving behavior behind a session-owned classifier. For example, `internal/session/resolve_test.go` proves the core resolver, and `internal/api/session_model_phase0_interface_spec_test.go` proves API/mail factory-target rejection, but the design does not say which existing CLI or worker tests characterize their current resolution policy before refactoring.
- [Major] The design states that different callers may keep different allowed behaviors, but it does not specify the first operation-specific API boundary enough to prevent a universal target policy from emerging during implementation. A later task writer still has to infer whether the first slice adds a pure classifier, a caller adapter, a command, or a replacement for existing `ResolveSessionID` call sites.

**Missing evidence:**
- Required Mayor front matter for an implementation plan: `plan_slug`, `phase: implementation-plan`, `rig`, `rig_root`, `artifact_root`, `requirements_file`, `status`, `created_at`, and `updated_at`.
- The exact implementation-plan sections required by the schema, especially `Current System`, `Proposed Implementation`, `Data And State`, `Testing`, `Rollout And Recovery`, and `Open Questions`.
- A scenario-row matrix for the target-classification slice. At minimum, the design should identify whether the slice touches `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-008`, and `SESSION-ID-009`, and should say why other nearby named-session rows are in or out of scope.
- A caller-surface inventory naming files and tests for API, CLI, worker, mail, and any reconciler path that currently resolves or classifies session targets.
- Exact test commands or package targets that prove no behavior drift after moving one caller to the session-owned classifier.
- A concrete checkpoint that prevents the first slice from absorbing work assignment, mail delivery policy, provider execution, pool scaling, or generic lifecycle orchestration.

**Required changes:**
- Replace this reviewed artifact, for this workflow, with a schema-conforming Mayor `implementation-plan.md`, or change the review workflow/schema if the intended target is actually a module-local `internal/session/DESIGN.md`. Under the current schema, the correct outcome is `blocked:wrong-artifact`.
- Add the required YAML front matter and exact top-level sections from `gc.mayor.implementation-plan.v1`; do not append design-review notes to `internal/session/DESIGN.md`.
- In `Current System`, cite concrete current files and tests, including `internal/session/resolve.go`, `internal/session/resolve_test.go`, `internal/session/REQUIREMENTS.md`, `internal/session/AGENTS.md`, and the API/mail tests that enforce rejection of `template:<name>` and bare ordinary config names.
- In `Proposed Implementation`, define the first narrow API/adapter shape precisely enough for task decomposition. It should preserve resolver identity classification while leaving caller-specific allow/deny policy outside the core classifier.
- In `Testing`, list exact scenario rows and proof commands. The minimum test plan should include the existing resolver tests plus caller-facing tests for API/mail/session operations and any CLI or worker caller moved in the first slice.
- In `Rollout And Recovery`, state whether durable session metadata changes are in scope. If the first slice is pure classification with no durable-state migration, say so explicitly and explain rollback as reverting the caller move.

**Questions:**
- Is this design review supposed to normalize a Mayor `implementation-plan.md`, or was the output schema accidentally attached to a module-local `internal/session/DESIGN.md` review?
- Which caller is the first implementation slice: API session operations, API/mail target resolution, CLI session commands, `internal/worker.Factory.HandleForTarget`, or another path?
- Should mail be part of the first target-classification slice, or is it intentionally out of scope because mail has named-session recipient behavior distinct from live session operations?
