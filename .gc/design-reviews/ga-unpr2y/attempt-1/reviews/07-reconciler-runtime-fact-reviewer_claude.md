# Liam Okonkwo - Claude

**Verdict:** approve-with-risks

Reviewed revision: `internal/session/DESIGN.md` (692 lines — the attempt-15
`iterate` response; global verdict still `block`), against
`internal/session/REQUIREMENTS.md`, `internal/session/AGENTS.md`, and direct
source verification on this branch (`fix/required-artifact-store-errors-ga-ksno8`).
Lane: reconciler boundary, runtime intent adapter ownership, fact isolation,
health gate split.

This revision resolves both `[Major]` blocks my prior (attempt-14) review raised
in this lane, which is why the verdict moves from `block` to
`approve-with-risks`:

- The health-gate-anchoring block is answered. Slice 0 now must "repair or
  owner-retire the evidence for `SESSION-RECON-002`, `SESSION-RECON-003`,
  `SESSION-RECON-006`, and `SESSION-RECON-007` before a later slice cites those
  rows" (DESIGN.md:175-178), and every `BOUNDARY_MATRIX.yaml` row "must
  distinguish what exists on the current branch from behavior that exists only on
  another ref or in historical notes" (DESIGN.md:493-494). I verified these are
  the correct rows to name: on this branch `provider_health_gate_test.go` and
  `session_progress_test.go` are absent and no non-test `cmd/gc` source carries a
  `ProviderHealth`/red-episode gate, yet REQUIREMENTS.md:133-134 still cite those
  files as proof. The design now forces that staleness to be reconciled instead
  of asserting an ownership it cannot anchor.
- The wake-cause-production-seam block is answered. "wake-cause production" is now
  a required `BOUNDARY_MATRIX` row (DESIGN.md:490) with an explicit "wake-cause
  production owner" field (DESIGN.md:486).

The remaining risks are `[Minor]` row-level refinements that should land in the
relevant `BOUNDARY_MATRIX.yaml`/Atomic Command Contract rows before slices 3, 5,
and 6 move. None is an existential flaw in the split; the boundary table itself
is correct.

**Top strengths:**

- The fact-isolation pattern this lane wants is already real, which makes the
  split credible rather than aspirational. `ProjectLifecycle(input
  LifecycleInput) LifecycleView` (`lifecycle_projection.go:374`) imports only
  `strings`/`time` — no bead store, no reconciler helpers — and consumes
  precomputed facts. The design's "caller gathers facts -> internal/session
  decides" shape (DESIGN.md:30) and the Atomic Command Contract's "immutable
  facts read by the decider, including mandatory `now` ... runtime observation
  timestamp when runtime facts are used" (DESIGN.md:364-368) codify exactly this
  existing shape. Lane Q2 (facts precomputed by adapters, not queried by
  deciders) is answered by both the doc and the code.
- The "Reconciler, Runtime, And Session Split" ownership table (DESIGN.md:459-465)
  is a correct, explicit answer to lane Q1: lifecycle projection / terminal
  states / wake blockers / identity conflicts -> `internal/session`; work demand
  / dispatch scheduling / pool desired size / cold-start demand / restart budgets
  / alert dedupe / progress policy / idle-sleep -> controller/reconciler; runtime
  liveness / provider errors / process existence -> runtime provider or
  worker/runtime adapter. Backlog slice 6 restates the "keep" side
  (DESIGN.md:677-679), and the cost section bounds reconciler fact compilation
  (probe/query counts, "No subprocess loop is allowed in ... reconciler hot
  loops", DESIGN.md:572,582), which keeps fact materialization adapter-side and
  cheap.
- RuntimeIntent is bounded the way lane Q3 requires: provider-neutral fields only
  (stable bead ID, generation/instance token, provider family, runtime session
  key, work dir, config hash) with an explicit exclusion of "provider-specific
  scheduling, health, progress, budget, or alert decisions" (DESIGN.md:467-471),
  held `provisional` until a real caller proves the field set (DESIGN.md:596).
  Paired with "Destructive actions with unknown, stale, partial, or
  provider-error runtime facts are rejected unless ... a safe convergence rule"
  (DESIGN.md:473-475) and events being post-commit facts that converge from
  durable scans (DESIGN.md:498-499,517-519), the "runtime missing observations
  become durable session truth" red flag is well defended on destructive paths.

**Critical risks:**

- [Minor] **Runtime-missing cleanup is itself a durable write driven by a missing
  observation, and its convergence-safety rule is not anchored to the existing
  false-negative protection.** The destructive-action-rejection rule
  (DESIGN.md:473-475) already covers "unknown/missing" facts, and
  "runtime-missing cleanup" is a required Atomic Command Contract operation
  (DESIGN.md:390) and `BOUNDARY_MATRIX` row (DESIGN.md:491). But the design never
  binds that cleanup's "safe convergence rule" to `SESSION-RUNTIME-001` (probe
  false-negative tolerance; "Providers without process names do not force a
  running session dead") or `SESSION-LIFE-005` (runtime-missing must preserve
  resume identity). Without that anchor the per-slice author can reinvent — or
  omit — the anti-flap rule, and a single transient missing probe can become
  durable closed/released truth. This is my red flag #2 narrowed to the one
  operation that legitimately acts on a missing fact.
- [Minor] **The wake-eligibility / health-gate "cannot move cleanly" risk has an
  inventory but no required disposition.** Slice 0 captures the "receiver/helper
  chain" in `SESSION_BOUNDARY_SYMBOLS.yaml` (DESIGN.md:157) and each
  `BOUNDARY_MATRIX` row carries "forbidden `internal/session` imports" plus a
  negative fixture (DESIGN.md:481,487), so a helper shared across the
  session-owned wake-eligibility / reconciler-owned health-gate boundary would
  surface. But nothing requires an explicit decision (split / duplicate / move to
  a neutral package) when that sharing is detected. Slice 3 (explicit wake,
  DESIGN.md:667-670) and slice 6 (reconciler facts, DESIGN.md:677-679) both touch
  `cmd/gc/session_reconciler.go` and could each claim or strand the same helper.
  This is my red flag #3; the migration-coexistence section (DESIGN.md:430-450)
  names that shared file but is writer/field-family focused, not read-side decision
  helpers.
- [Minor] **The decider read boundary is stated for commands but not for the
  read-only classifier, which leaves the "no bead-store queries in deciders" red
  flag half-open.** The Atomic Command Contract passes immutable fact snapshots
  (DESIGN.md:364-368) — clean. But the Target Classification Contract forbids only
  *writes* ("no store writes", DESIGN.md:211) and the cost section confirms
  classification performs indexed store lookups on the hot path
  (DESIGN.md:567-568), so the first "session decider" does read the store. Reading
  *session* beads via session-package resolvers is defensible; the gap is that the
  design never states the line: deciders may read session beads via
  `internal/session` resolvers, but must not import reconciler helpers or
  reconciler-side demand/schedule/health queries, and runtime/progress/demand
  facts must arrive as immutable snapshots. State it so the rule of "deciders
  consume precomputed facts" is not silently weakened to "deciders may query
  anything read-only."
- [Minor] **`hold` (and `wait`) are in the Slice 0 matrix schema but dropped from
  the required-before-moving rows, and hold production vs projection owners are not
  separated.** The `BOUNDARY_MATRIX` universal inventory lists "wake, hold, wait,
  drain, ..." (DESIGN.md:161), but the "Required rows before behavior-moving
  slices" list (DESIGN.md:489-492) omits hold and wait. `SESSION-RECON-004` shows
  the interaction is non-trivial: holds suppress config/attached/work wake reasons
  except wait-only wake — and `work`/`scale-demand` are reconciler-owned causes.
  Hold *projection* is session-owned (a blocker) while hold *production* (e.g.,
  restart-budget backoff) is reconciler-owned; the matrix has a "wake-cause
  production owner" field but no equivalent for hold production. Slice 3 can
  entangle hold suppression with reconciler wake causes without a pinned row.
- [Minor] **`provider family` is the one RuntimeIntent field that can become a
  policy seam, and there is still no routing-only caveat.** It is allowed to cross
  into a session-owned command (DESIGN.md:470), which is fine for routing a start
  to the right provider. But `SESSION-RUNTIME-003` (ACP routing) is live
  provider-branching behavior; if any session decider branches lifecycle policy on
  provider family, that is provider policy inside `internal/session`. The field
  list restricts *which* fields cross but not *what session code may do with
  them*. Add: provider family may be carried for routing/identity only; session
  deciders must not branch lifecycle policy on it.

**Missing evidence:**

- No explicit disposition of the existing reconciler observation->durable
  conversion (`sleep_reason="runtime-missing"` written reconciler-side) against
  the Atomic Command Contract's "runtime-missing cleanup" operation: does that
  write stay reconciler policy, or become a session-owned command, and on what
  corroboration threshold?
- Slice 6's "extract only narrow lifecycle eligibility facts" (DESIGN.md:677-679)
  has no example inventory and no explicit negative. Which facts are in
  (hold/quarantine expiry, stale-creating, continuity eligibility) and which are
  permanently out as decider-required inputs (work counts, pool size, provider
  health, progress)? The ownership table implies it; slice 6 does not restate it.
- No stated anti-flap / corroboration threshold (consecutive misses or grace
  window) that gates runtime-missing from observation to durable cleanup, beyond
  the general "reject destructive actions on missing facts" rule.

**Required changes:**

- In the runtime-missing-cleanup Atomic Command Contract / `BOUNDARY_MATRIX` row,
  name `SESSION-RUNTIME-001` (false-negative tolerance) and `SESSION-LIFE-005`
  (resume-identity preservation) as the convergence-safety anchors, and require an
  explicit anti-flap rule (corroboration count or grace window) so a single
  transient missing probe cannot become durable closed/released truth.
- Add a required disposition to the boundary process: when the Slice 0 helper-chain
  inventory shows a helper shared across the session-owned wake-eligibility /
  reconciler-owned health-gate boundary, the matrix row must record the decision
  (split, duplicate, or move to a neutral package) before slices 3 and 6 touch
  `cmd/gc/session_reconciler.go`.
- State the decider read boundary explicitly: deciders may read session beads via
  `internal/session` resolvers, but must not import reconciler helpers or
  reconciler-side demand/schedule/health queries; runtime, progress, and demand
  facts arrive as immutable snapshots (matching the Atomic Command Contract).
- Promote `hold` (and `wait`) to required `BOUNDARY_MATRIX` rows before
  behavior-moving slices, and add a "hold production owner" distinction (reconciler
  produces backoff/budget holds; session projects them as blockers), citing
  `SESSION-RECON-004`.
- Add the routing-only caveat on `provider family`: carried for routing/identity
  only; session deciders must not branch lifecycle policy on it (cf.
  `SESSION-RUNTIME-003`).

**Questions:**

- Does the reconciler's `sleep_reason="runtime-missing"` conversion remain
  reconciler policy permanently, or is it the "runtime-missing cleanup" operation
  the Atomic Command Contract lists for future session ownership — and if the
  latter, what corroboration threshold gates it?
- Is the intended end-state that wake-cause production stays entirely in the
  reconciler/worker adapter (feeding `ProjectLifecycle` via `[]WakeCause`), with
  slice 3 moving only the explicit-wake operation and the other seven causes
  passed in as facts?
- When a session-owned wake/close/runtime-start command needs the caller to
  perform runtime work, does it return bounded RuntimeIntent that the caller
  executes, or does the caller keep its own runtime sequencing? DESIGN.md:33-35
  ("caller executes or renders") admits both readings.
