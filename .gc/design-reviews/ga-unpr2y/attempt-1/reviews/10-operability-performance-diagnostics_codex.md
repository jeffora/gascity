# Ingrid Holm - Operability, Performance, Diagnostics (Codex)

**Persona verdict:** approve-with-risks

Approve the current design for Slice 0 only. The revision is directionally
sound: it requires structured diagnostics, trace mappings, renderer proof, and
hot-path budgets before delegation. It should not approve any behavior-moving
slice until those requirements exist as machine-readable artifacts with tests.

## What is now strong enough

- Slice 0 now owns the correct operability artifacts. The required artifact set
  includes `DIAGNOSTICS_MANIFEST.yaml` for operation/check IDs, reason/outcome
  codes, renderer surfaces, trace mappings, redaction, event relationships,
  query budgets, subprocess budgets, and hot-loop constraints
  (`internal/session/DESIGN.md:150-165`).
- The design correctly bans local-string-only decisions. Classifiers, deciders,
  and commands must return structured diagnostics with operation ID, result
  kind, reason code, retryability, selected identity, source facts, stale/missing
  fact markers, renderer surfaces, and redaction keys
  (`internal/session/DESIGN.md:541-554`).
- Diagnostics must be observable through the existing incident tooling. Every
  diagnostic row must map to a `gc trace` site/reason/outcome rendering or
  explicitly state that it is intentionally not trace-rendered, with source
  facts, redaction, renderer tests, and negative tests
  (`internal/session/DESIGN.md:556-564`). That aligns with the reconciler
  debugging runbook, which makes `gc trace status`, `gc trace reasons`,
  `gc trace show --type cycle_result --json`, full cycle dumps, event logs, and
  controller logs the standard incident evidence
  (`engdocs/contributors/reconciler-debugging.md:24-81`).
- The design now names the hot-path performance risks. It requires bounded
  indexed lookups or counting-store proof, forbids all-session hot-path scans
  without a measured budget and large-city baseline, forbids subprocess loops in
  classification and reconciler hot loops, and requires event/subscriber scan
  caps or benchmark-relative budgets (`internal/session/DESIGN.md:566-574`).
- The known expensive API path is explicitly covered. Today
  `resolveLiveSessionByPathAlias` calls `session.ListAllSessionBeads` and scans
  session beads in process (`internal/api/session_resolution.go:392-426`).
  The design requires a specific decision to index, remove, or keep that scan
  with a maximum-row budget and large fixture proof
  (`internal/session/DESIGN.md:580-581`).
- The design is careful about existing scan/subprocess tradeoffs. The named
  session resolver documents why it uses one label-scoped scan rather than four
  metadata subprocess calls under reconciler/wake load, including measured
  improvement from 5.2s to 1.3s
  (`internal/session/named_config.go:350-363`). The new budget language gives
  future changes a way to preserve that kind of evidence instead of replacing it
  with unmeasured per-session fan-out.

## Required Slice 0 close gates

1. **`DIAGNOSTICS_MANIFEST.yaml` must be validated by runnable tests, not just
   checked into the repo.** The design's minimum proof command names
   `TestSessionDiagnosticsManifest`, but that validator is not present today.
   Slice 0 must fail if any diagnostic row lacks centralized operation/reason
   constants or manifest IDs, trace rendering, doctor/session-inspect rendering
   where applicable, API/CLI renderer proof, redaction metadata, or negative
   tests proving machine-readable data is not hidden only in messages.
2. **Trace coverage must include negative and recovery cases.** Accepted
   decisions are not enough. The manifest must cover rejected, blocked,
   deferred, stale-fact, partial-fact, repair-needed, event-emission-failed,
   event-skipped, duplicate-event, scan-recovered, and scan-failed outcomes for
   any slice that touches wake, drain, close, runtime start, provider health, or
   work release. The runbook expects operators to reason from `gc trace` and
   `.gc/events.jsonl`, not from stdout/stderr alone
   (`engdocs/contributors/reconciler-debugging.md:77-81`).
3. **Budget rows need numeric or baseline-relative thresholds.** The current
   design correctly names required rows for API target resolution,
   `resolveLiveSessionByPathAlias`, reconciler fact compilation, and event
   recovery scans (`internal/session/DESIGN.md:576-583`). Slice 0 must make
   those rows concrete: max store calls, max scanned rows, max subprocesses,
   fixture size, benchmark command or counting-store test, threshold, owner, and
   allowed delta.
4. **Large-city proof must be tied to the route or loop that pays the cost.**
   A generic benchmark is not enough. For API target resolution, prove Huma and
   legacy query endpoints. For reconciler fact compilation, prove the controller
   loop with representative session/work counts and partial snapshots. For
   event recovery scans, prove retry/cap behavior and duplicate idempotency.
5. **Repair/doctor paths must emit before/after evidence and propagate
   persistence failures.** This matters for `RepairEmptyType` and future
   backfills. If a repair is operator-visible, the manifest row should name
   trace, doctor, or proof evidence and redaction behavior; if it is not
   trace-rendered, that must be an explicit decision.
6. **Subscriber fan-out and durable scans need backpressure tests.** The design
   requires caps or benchmark-relative budgets, but the close gate should prove
   event emission cannot synchronously amplify work by session count and that
   missed-event convergence does not rely on unbounded scans.

## Residual risks

- The design still has no actual Slice 0 artifacts in this checkout. That is
  acceptable for a design pass only because Slice 0 is non-mutating. It is not
  acceptable for any later slice to cite these sections as if they are already
  enforced.
- The path-alias lookup remains an all-session scan today. That is an explicit
  first-adopter risk, not a blocker, because the design now requires a decision
  and measured budget before delegation.
- Some existing operational behavior is still message-string based. The design
  can tolerate that as legacy inventory, but new decision outputs should not add
  more local strings without manifest IDs and renderer tests.

## Required changes before operability-sensitive implementation

- Add `DIAGNOSTICS_MANIFEST.yaml` with stable operation IDs, reason/outcome
  codes, trace site mappings, doctor/session-inspect/API/CLI renderer rows,
  event relationships, redaction keys, and negative fixtures.
- Add budget rows and validators for API target resolution, path-alias lookup,
  reconciler fact compilation, event recovery scans, subscriber fan-out, and
  large-city baselines.
- Add proof commands that fail on missing diagnostics renderers, zero-match
  budget tests, absent large fixtures, unbounded scans, subprocess loops in hot
  paths, and event recovery without idempotency or caps.
- Require every behavior-moving slice to reference the specific diagnostics and
  budget row IDs it consumes, rather than citing the design section generally.

## Bottom line

This design can move to Slice 0. It must not move session behavior until
diagnostics and performance budgets are executable, trace-visible, and tested
against large-city and missed-event cases.
