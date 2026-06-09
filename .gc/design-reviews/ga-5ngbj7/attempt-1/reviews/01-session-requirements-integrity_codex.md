# Session Requirements Integrity Reviewer - Codex

**Verdict:** block

**Top strengths:**
- The document is valuable as a session behavior ledger: it gives a clear purpose for the Session primitive, a reconcile rule that keeps code/tests/docs aligned, and durable invariants around bead-backed state, exact targeting, work release, and idempotent reconciliation (`internal/session/REQUIREMENTS.md:9`, `internal/session/REQUIREMENTS.md:15`, `internal/session/REQUIREMENTS.md:56`).
- The canonical vocabulary section correctly anchors lifecycle language to `ProjectLifecycle` and names base states, desired states, runtime projections, identity projections, blockers, and wake causes (`internal/session/REQUIREMENTS.md:35`, `internal/session/REQUIREMENTS.md:60`).
- The scenario ledger covers important high-risk surfaces: lifecycle projection, reducer transitions, identity resolution, create/wake/close behavior, reconciler behavior, work release, drain safety, runtime submission, and observation (`internal/session/REQUIREMENTS.md:73`, `internal/session/REQUIREMENTS.md:136`, `internal/session/REQUIREMENTS.md:145`).

**Critical risks:**
- [Blocker] The artifact does not satisfy the requested Mayor requirements schema. The schema requires YAML front matter and exact top-level sections `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`, `Out Of Scope`, and `Open Questions`; this file starts with a Markdown heading/table and uses a module-local ledger shape (`requirements.schema.md:15`, `requirements.schema.md:41`, `internal/session/REQUIREMENTS.md:1`, `internal/session/REQUIREMENTS.md:3`, `internal/session/REQUIREMENTS.md:13`).
- [Blocker] Several cited evidence paths are stale in this checkout. `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go` are cited as proof but do not exist, leaving `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` without live test evidence (`internal/session/REQUIREMENTS.md:129`, `internal/session/REQUIREMENTS.md:130`, `internal/session/REQUIREMENTS.md:133`, `internal/session/REQUIREMENTS.md:134`).
- [Major] The status value is not schema-compatible and is ambiguous for downstream workflow gates. The schema allows `draft`, `questions`, `approved`, or `blocked`; this file has a table field `Status | Seed draft`, outside YAML front matter (`requirements.schema.md:32`, `requirements.schema.md:35`, `internal/session/REQUIREMENTS.md:3`, `internal/session/REQUIREMENTS.md:5`).
- [Major] The document has many evidence cells that name broad files or commits rather than specific enforcing tests, commands, or cases. For example, rows cite `internal/session/manager_test.go` or `cmd/gc/session_reconcile_test.go` for multi-part behavior, which makes it hard to verify happy, negative, and edge coverage without re-deriving intent from the test suite (`internal/session/REQUIREMENTS.md:83`, `internal/session/REQUIREMENTS.md:115`, `internal/session/REQUIREMENTS.md:149`).
- [Major] Durable user-visible outcomes exist as ledger rows, but they are not expressed as schema-ready examples or acceptance criteria. Drain-ack/recovered work, confirmed-dead workers, invalid targets, stale snapshots, and repeated controller passes are present in prose, but not as concrete input/expected-output examples with verification methods (`internal/session/REQUIREMENTS.md:69`, `internal/session/REQUIREMENTS.md:107`, `internal/session/REQUIREMENTS.md:141`, `internal/session/REQUIREMENTS.md:142`, `internal/session/REQUIREMENTS.md:143`).
- [Minor] The file explicitly says it is a reconciliation ledger and not a scratchpad for implementation plans, which is useful locally but also confirms it is not the Mayor `requirements.md` artifact shape requested by the output schema (`internal/session/REQUIREMENTS.md:9`, `internal/session/REQUIREMENTS.md:160`).

**Missing evidence:**
- YAML front matter with `plan_slug`, `phase: requirements`, `rig`, `rig_root`, `artifact_root`, schema-compatible `status`, and timestamps.
- Required sections for W6H, Example Mapping, top-level Acceptance Criteria, Out Of Scope, and Open Questions.
- Live replacement evidence for the missing scale-from-zero, provider-health, and progress-aware health test paths.
- Per-row proof granularity: exact test names, commands, or source assertions that enforce each scenario rather than broad file references.
- Concrete examples for drain-ack/recovered work, confirmed-dead session close, invalid target rejection, stale snapshot avoidance, and repeated controller idempotence.

**Required changes:**
- Decide whether this file is meant to be a Mayor requirements artifact or a module-local session ledger. If it is a Mayor requirements artifact, rewrite it to `gc.mayor.requirements.v1`; if it is intentionally a ledger, use a different output schema or review contract.
- Add schema-compliant front matter and required top-level sections before treating the artifact as implementation-plan ready.
- Replace stale evidence paths with live tests or restore the missing tests. At minimum, repair the evidence for `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`.
- Tighten evidence cells so each row cites the specific test function, command, source guard, issue, or commit that proves the behavior.
- Add Example Mapping for the high-risk session outcomes named in the persona: drain-ack, confirmed-dead sessions, invalid targets, recovered work after drain cancellation, stale snapshots, and repeated controller passes.
- Convert scenario outcomes into top-level acceptance criteria with explicit proof types if this document will remain under the Mayor requirements workflow.

**Questions:**
- Is `internal/session/REQUIREMENTS.md` supposed to become a schema-compliant Mayor requirements artifact, or should it remain a scoped module ledger with its own schema?
- What replaced `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`, if anything?
- Should evidence cells require exact test function names to prevent broad file references from going stale silently?
