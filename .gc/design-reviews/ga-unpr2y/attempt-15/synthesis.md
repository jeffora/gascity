# Design Review Synthesis

## Overall Verdict: block

Nine of ten persona syntheses return `block`; the remaining persona approves only a narrowed, non-mutating evidence preflight and blocks behavior-moving slices until coexistence and ownership rules are resolved. The design has strong direction, but the current artifact still lets implementers proceed with stale proof, broad future-facing contracts, porous mutation boundaries, and underspecified runtime/event recovery semantics.

## Consensus Strengths
- Multiple personas praised the reset toward a small first adopter, side-effect-free classification/deciders, and one-caller-at-a-time extraction.
- Reviewers consistently agreed that session-owned commands, explicit fact inputs, durable recovery scans, and generated inventory artifacts are the right direction.
- The Vocabulary Lifecycle, anti-`SessionService` posture, event-sourcing deferral, and controller-owned infrastructure model are aligned with Gas City architecture principles.
- The migration strategy is viable as a non-mutating Slice 0 evidence preflight if Slice 0 is narrowed, made mechanically enforceable, and prevented from becoming an all-future-slices contract.

## Critical Findings

### [Blocker] Slice 0 Is Not Yet A Reliable Gate
**Sources:** Natasha Volkov; Kwame Asante; Ravi Krishnamurthy; Sarah Chen; Liam Okonkwo
**Issue:** Slice 0 mixes universal discovery with detailed contracts for future wake, close, drain, runtime-start, repair, diagnostics, and provider-side behavior. At the same time, its proof command can pass while named validators are absent or stale, and active requirements rows such as `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` cite missing or stale executable proof.
**Required change:** Split Slice 0 into source-complete universal evidence and per-slice preflight contracts. Make the Slice 0 proof command self-validating, fail on absent/skipped validators, repair or owner-retire stale requirement evidence, and define exact artifact schemas, owners, negative fixtures, and proof selectors before dependent slices can cite them.

### [Blocker] First-Adopter Target Classification Would Regress Current Resolution Semantics
**Sources:** Amara Diallo; Sarah Chen; Natasha Volkov; Ingrid Holm; Kwame Asante
**Issue:** The API query first-adopter contract does not match the current resolver order. It omits configured named-session canonical/conflict handling, config-orphan rejection, path-alias by `Title`, closed lookup rules, and repairable empty-type behavior. The proposed candidate taxonomy is also too flat: one `kind` cannot preserve overlapping alias/session-name matches, configured identity state, liveness, diagnostics, and result policy without caller reinterpretation.
**Required change:** Rewrite the first-adopter row around the exact `resolveSessionTargetIDWithContext` behavior: reject `template:<name>`, direct bead ID, configured named-session handling before live aliases, config-orphan rejection, live `session_name`/`alias`, API path-alias with state filter and newest-created tiebreaker, then closed lookup where currently allowed. Use a result schema that separates match vectors, bead/config state, diagnostics, and terminal result kinds; preserve `writeResolveError` and `humaResolveError` wire behavior with parity tests.

### [Blocker] Session Mutation Ownership Still Has Escape Hatches
**Sources:** Elena Marchetti; Sarah Chen; Ravi Krishnamurthy; Takeshi Yamamoto
**Issue:** Generic bead mutation bridges, `beads.UpdateOpts.Metadata`, create-time metadata, direct close/reopen/status/type writes, exported repair helpers, patch constructors, and helper-returned metadata maps can still mutate session-owned fields outside session-owned command APIs. Path-level exceptions for API/CLI bead update surfaces cannot constrain dynamic metadata keys or target bead IDs, and shrink-only exception text does not by itself prevent indefinite multi-writer coexistence.
**Required change:** Add runtime bridge fencing or session-command routing for session-owned metadata key families and session bead `Type`/`Status`. Expand the mutation ledger and guard contract to all bead-store mutation forms, enforce exception expiry and retirement in CI, and add negative fixtures for dynamic-key, bridge, create-time metadata, status/type, repair, and patch-map bypasses.

### [Blocker] Atomic Command Semantics Are Not Enforceable
**Sources:** Takeshi Yamamoto; Liam Okonkwo; Elena Marchetti; Ingrid Holm
**Issue:** The design lists immutable facts, stale-fact rules, fence tokens, recovery fields, and runtime intents, but does not require commits to revalidate the exact snapshot a decider consumed. Store-level compare-and-swap support is not defined for current persistence surfaces, partial metadata writes are not mapped, and wake/start/close/drain/stop/repair ordering is still a checklist rather than operation-specific contracts.
**Required change:** Define immutable operation fact snapshots, including mandatory `now` and relevant config facts, plus the exact token/revision/value preconditions checked immediately before commit. Add command rows for wake/start, close, drain, stop/interrupt, identity retirement, runtime-missing cleanup, and repair/backfill that name store semantics, stale outcome, post-write verifier, runtime/event ordering, repair owner, and failure-injection tests.

### [Blocker] Event Recovery And Diagnostics Are Not Concrete Enough
**Sources:** Amara Osei; Ingrid Holm; Takeshi Yamamoto
**Issue:** Existing `session.*` events are not inventoried with stable canonical session identity, typed payload fields, critical-vs-best-effort classification, durable scan owners, idempotency keys, and public wire impact. Close/retire work-release convergence lacks release-identity snapshots, scanner triggers, completion markers, duplicate/stale handling, and partial-query behavior. `DIAGNOSTICS_MANIFEST.yaml` is also not specified enough to preserve `gc trace` site/reason/outcome visibility after side-effect-free classifiers and deciders move logic out of call sites.
**Required change:** Add machine-readable event and diagnostics ledgers for current session events and first-slice decisions. Require stable canonical identity, typed payloads where envelope fields are insufficient, controller-owned durable scan rows for safety-critical recovery, shared idempotency keys, diagnostic-to-trace mappings, renderer proof, and negative tests for skipped events, duplicates, stale events, partial scans, and message-only machine data.

### [Blocker] Reconciler, Runtime, And Session Fact Boundaries Are Still Too Implicit
**Sources:** Liam Okonkwo; Takeshi Yamamoto; Ravi Krishnamurthy; Ingrid Holm
**Issue:** `BOUNDARY_MATRIX.yaml` is not defined enough to keep policy and demand outside `internal/session`. Provider-health and progress behavior are not anchored to active proof on this branch, wake-cause production is unnamed, and unknown/stale/partial/provider-error runtime facts do not say when destructive actions fail open, fail closed, or remain advisory.
**Required change:** Define the boundary matrix row schema before behavior-moving slices: source owner, policy owner, allowed session inputs/outputs, freshness rule, unknown/stale/partial/provider-error behavior, destructive-action safety, current code/test selectors, and forbidden session imports/fields. Add rows for provider health, progress, runtime observations, wake-cause production, partial snapshots, drain, runtime-missing cleanup, and adapter provider actions.

### [Blocker] Operability And Cost Gates Miss Hot Paths
**Sources:** Ingrid Holm; Amara Diallo; Kwame Asante; Sarah Chen
**Issue:** The first-adopter path can include `resolveLiveSessionByPathAlias`, an all-session scan by `Title`, but no cost row names, indexes, budgets, or removes it. Diagnostic and route inventories risk becoming speculative registries while still failing to provide concrete first-slice trace, renderer, source-fact, and budget guarantees.
**Required change:** Add per-hot-path budget rows for target resolution and reconciler fact compilation, including query shape, maximum store calls, maximum scanned rows or index proof, subprocess count, fixture size, proof command, threshold, and owner. Either index, remove, or explicitly budget path-alias resolution before the classifier delegates to that path.

### [Major] Worker Boundary And API/CLI Compatibility Need Hard Choices
**Sources:** Sarah Chen; Ravi Krishnamurthy; Elena Marchetti
**Issue:** Wake and drain are still allowed to become ordinary store-level lifecycle operations by local slice decision, and API close already diverges between legacy worker-routed close and Huma direct manager close. Route inventory requirements do not yet force no-delta API, CLI, generated-client, dashboard, stdout/stderr, exit-code, and JSON compatibility.
**Required change:** State that production CLI/API mutating lifecycle operations route through `worker.Handle` unless a root-approved, exact, expiring exception row exists with parity and retirement proof. Route Huma close through the worker boundary or add a temporary exception with tests for events, cleanup, response shape, OpenAPI/generated clients, and dashboard compatibility.

### [Major] Migration Coexistence And Rollback Are Under-Specified
**Sources:** Ravi Krishnamurthy; Sarah Chen; Elena Marchetti; Takeshi Yamamoto
**Issue:** Worker-boundary and session-mutation-boundary migrations overlap in `internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and `cmd/gc/session_beads.go` without an authoritative sequence. Once new validated writers exist, legacy raw writers can still mutate the same field families unless ownership transfer is atomic or physically fenced.
**Required change:** Add cross-migration ordering for shared files, per-field ownership transfer rules, raw-writer retirement conditions, and per-slice rollback data-direction rules. If old and new writers intentionally coexist for a field family, require a concrete cross-process conditional-write fence.

### [Major] RepairEmptyType Is Still An Under-Owned Side Effect
**Sources:** Elena Marchetti; Amara Diallo; Ravi Krishnamurthy; Ingrid Holm
**Issue:** Read-like resolution paths can still call `session.RepairEmptyType`, and reviewers flagged both hidden mutation during read-only classifier adoption and unsafe behavior if persistence errors are swallowed while in-memory state is patched.
**Required change:** Quarantine `RepairEmptyType` from read-only classifier paths, route it through an audited repair owner, or return `repair-needed` from read paths. Persisted repair failures must propagate, and before/after trace evidence plus parity tests must prove current successful target selection is preserved or intentionally amended.

### [Minor] Historical Notes And Message Text Need Clear Non-Normative Boundaries
**Sources:** Kwame Asante; Amara Osei
**Issue:** `DESIGN_REVIEW_NOTES.md` is described as historical, but validators or implementation beads could still treat uncopied notes as acceptance criteria. Separately, some current event messages may embed data in human text while using `NoPayload`.
**Required change:** State that historical notes are rationale only unless copied into active approved artifacts. Promote machine-readable event data into typed payloads, or make clear that message text is operator-only and not subscriber input.

## Disagreements
- Verdict severity: nine persona lanes block. Ravi Krishnamurthy returns `approve-with-risks` only for a narrowed, non-mutating Slice 0 evidence preflight; that does not outweigh the global blockers because the design currently authorizes or prepares behavior-moving slices without the required gates.
- Store strategy: reviewers disagree on whether native conditional writes, value-embedded token fencing plus repair, or stronger cross-process synchronization is required. The synthesis assessment is that any approach can be acceptable only when chosen per operation, tied to current store semantics, and proven with stale/partial/failure tests.
- Target classifier shape: some reviewers would accept the broad candidate table as provisional documentation, while others require tagged per-kind production structures now. The required invariant is narrower: active production and wire vocabulary for Slice 1 must be limited to the first API-query adopter, and production shared types must not be broad optional envelopes.
- Worker boundary route: reviewers differ on whether wake/drain may remain store-level commands under exceptions. The synthesis assessment is that the design must choose the route per operation before implementation beads are created and must make exceptions exact, expiring, owner-approved, and parity-proven.
- Event recovery severity: some reviewers see event guidance as mostly future-facing, while others block on existing `session.*` contracts. The synthesis assessment is that current public events and safety-critical work-release recovery are active design surface and need inventory plus durable convergence proof now.

## Missing Evidence
- Self-validating Slice 0 validator definitions, meta-checks, positive/stale fixtures, and proof commands.
- Live executable proof or owner-approved rewrite/retirement for stale `SESSION-*` rows, especially the cited reconciler scenarios.
- Exact first-adopter API query route list, precedence matrix, route inventory rows, and no-delta HTTP/Huma/CLI/generated-client/dashboard tests.
- Candidate/result schema that preserves multiple match vectors, configured identity state, liveness, repair-needed status, diagnostics, and terminal result kinds without caller reinterpretation.
- Machine-readable session-owned metadata key-family list and top-level session field list shared by static guards and runtime bridge denial.
- Concrete `SESSION_BOUNDARY_SYMBOLS.yaml`, `SCENARIO_PARITY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, `COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, `EVENT_RECOVERY.yaml`, and `DIAGNOSTICS_MANIFEST.yaml` schemas, owners, validators, fixtures, and proof selectors.
- Store capability matrix for atomic, partial, conditional, and blind update behavior across current session lifecycle command paths.
- Failure-injection tests for stale snapshots, raced lifecycle operations, partial writes, duplicate commands, provider success followed by commit failure, event failure after commit, and durable scan crash recovery.
- Complete current `events.Session*` inventory with stable identity, payload, idempotency, recovery owner, SSE/OpenAPI impact, and criticality.
- Large-city cost baseline for target resolution and reconciler fact compilation, including an explicit row for `resolveLiveSessionByPathAlias`.
- Branch-state proof for provider-health, progress, and scale-from-zero behavior: what exists on this branch, what exists only on another ref, and what must be restored, replaced, or retired.
- Artifact provenance note: the ten persona syntheses for `gc.attempt=15` were stamped by child beads with output paths under `attempt-1`; this synthesis normalized copies into `attempt-15/persona-syntheses` before writing the global report.

## Recommended Changes
1. Narrow Slice 0 to universal evidence and a self-validating proof harness; move detailed command, runtime, event, diagnostics, and mutation contracts to the slice that first adopts each behavior.
2. Repair or owner-retire stale requirement evidence and require row-to-slice, row-to-surface, row-to-proof parity metadata before any behavior-moving slice.
3. Rewrite the first target-classifier adopter around the exact API query resolver behavior, typed result states, match vectors, configured identity semantics, path-alias rules, repair-needed handling, and wire-error parity.
4. Fence all session-owned mutation surfaces: generic bead bridges, direct metadata updates, create-time metadata, status/type/close/reopen writes, repair helpers, patch constructors, and dynamic-key paths.
5. Define operation-specific atomic command rows with immutable fact snapshots, concrete store semantics, pre-commit revalidation, runtime/event ordering, repair ownership, and failure-injection tests.
6. Add event and diagnostics ledgers that preserve stable identity, durable recovery, idempotency, public wire impact, trace mappings, renderer proof, and negative tests for current `session.*` events and first-slice diagnostics.
7. Define `BOUNDARY_MATRIX.yaml` and add rows for provider health, progress, wake-cause production, runtime observations, partial snapshots, drain, runtime-missing cleanup, and adapter provider actions.
8. Make production CLI/API mutating lifecycle operations route through `worker.Handle`, or require exact expiring exceptions with owner, parity proof, retirement conditions, and same-change updates.
9. Add cross-migration sequencing, per-field ownership-transfer rules, raw-writer retirement conditions, and rollback proof for shared `cmd/gc` and `internal/api` files.
10. Fix or quarantine `RepairEmptyType` so read-only classifier paths do not silently mutate; route repair through an audited owner or return `repair-needed` with explicit parity and trace proof.
11. Add per-hot-path cost budgets for target resolution and reconciler fact compilation, including an index/remove/budget decision for `resolveLiveSessionByPathAlias`.
12. Keep non-first-adopter surfaces, diagnostic codes, event fields, and classifier vocabulary provisional until a concrete caller delegates and renderer/parity tests exist.
