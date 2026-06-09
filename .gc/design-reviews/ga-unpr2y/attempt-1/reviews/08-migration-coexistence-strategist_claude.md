# Ravi Krishnamurthy - Claude

**Verdict:** block

Lane: migration sequencing, legacy-new coexistence, rollback slices,
worker-boundary collision. Reviewed the current attempt-15 `iterate` revision of
`internal/session/DESIGN.md` against `REQUIREMENTS.md`, `internal/session/AGENTS.md`,
the root `AGENTS.md` worker-boundary migration notes, and the live code in
`internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`,
`cmd/gc/session_beads.go`, and `cmd/gc/session_lifecycle_parallel.go`. All
findings ground-checked in this checkout.

This revision is materially stronger for my lane than prior ones: the Migration
Coexistence section now exists, fences are required, and the
`RepairEmptyType`-as-silent-read-write hole is closed by the `repair-needed`
result kind (DESIGN.md:246–251). I am blocking on one code-grounded fact the
Slice 1 separability premise gets wrong, and on the close split-brain that the
worker-boundary scoping still leaves controller-side. The block is surgical:
Slice 0 (non-mutating) is unaffected and safe; the required changes gate the
first behavior-moving slices.

**Top strengths:**
- Migration Coexistence And Rollback is the right shape: it names the three
  overlapping files, and requires ordered predecessor/successor slices, per-field
  ownership before/during/after, raw-writer retirement conditions, fences, and
  rollback data-direction. Split-brain-on-value is closed — "an exception that
  keeps two unfenced writers for the same session-owned field is not allowed,"
  proven with raced-writer tests.
- The `RepairEmptyType` quarantine (DESIGN.md:246–251) is a real fix: the
  read-only classifier returns `repair-needed` instead of silently repairing.
  This directly retires the prior "read slice is not write-free" hole, even
  though the 10 live call sites (`handler_sessions.go:455/513/725`,
  `huma_handlers_sessions_command.go:426/483/876/931`, `handler_beads.go:89`,
  `cmd_session_wake.go:72`, `cmd_session_pin.go:114`) still repair today.
- Slice 0 is correctly non-mutating, and `WORKER_BOUNDARY_EXCEPTIONS.yaml`
  inventories the API session-resolution direct-create as an exception with owner
  and expiry, so the worst collision site is at least catalogued before behavior
  moves.

**Critical risks:**

- **[Blocker] Slice 1's "side-effect free, separable" premise is wrong in the
  code: the read-only classifier and the worker-boundary create path are one
  function with a mode flag.** `resolveSessionTargetIDWithContext`
  (`internal/api/session_resolution.go:429`) is the single resolver for all three
  modes — read-only `{}` (`:484`), read-only-closed `{allowClosed:true}` (`:488`),
  and **mutating** `{materialize:true}` (`:492`, `:496`, `handler_beads.go:74`).
  The `materialize:true` mode reaches `materializeNamedSessionWithContext`
  (`:259`) → `CreateAliasedNamedWithTransportAndMetadata` (`:328`) — the exact
  live worker-boundary exception in root `AGENTS.md`. Precedence steps 1–7 are
  shared code; the modes diverge only at the terminal action. So Slice 1 cannot
  extract the read-only precedence without carving it out of the same function
  the worker-boundary migration must relocate. Red flag #1 ("two half-finished
  boundaries collide on the same call sites") is confirmed live, on the *first*
  executable slice — and the design's per-surface matrix and "documented
  exceptions, not precedent" framing assume a disjointness the code does not have.

- **[Blocker] Freezing the mutating surface "characterization only" forks the
  resolution precedence with no anti-drift control.** The per-surface matrix
  (DESIGN.md:289) freezes the `materialize:true` path until its own contract
  exists. Mechanically that means Slice 1 publishes a new precedence
  implementation in `internal/session` while the frozen materialize path keeps
  its inline copy of identical steps 1–7 — and the worker-boundary migration is
  simultaneously relocating that inline copy. Two precedence copies, one validated
  by the new typed classifier contract and one not, drifting across two
  migrations: red flag #2 at the resolution layer. The generic "add a migration
  row" requirement never names this fork or mandates single-source-of-precedence,
  so a Slice 1 author following the current text creates it unknowingly.

- **[Major] Close is split-brain today, and the worker-boundary scoping leaves
  the controller-side writers out — so Slice 4 adds a close writer instead of
  converging them.** Close-alignment is scoped to "production **API and CLI**
  callers" (DESIGN.md:410), but the reconciler drained-close
  (`session_reconciler.go:206`, `ClosePatch(..., "drained")`), failed-create/sweep
  closes (`session_beads.go:1737/1744`, `:2161/:2164`), and parallel-start
  failed-create close (`session_lifecycle_parallel.go:1823`) all write `state` +
  terminal timestamps via raw `ClosePatch` + `store.Close`, bypassing
  `worker.Handle` — and none are API or CLI callers. When Slice 4 introduces a
  session-owned close command, the same close fields gain a third writer with
  different validation (red flag #2). The Atomic Command Contract lists
  "reconciler lifecycle transitions" as needing a contract (DESIGN.md:391), which
  contradicts the API/CLI-only close-alignment scope; the design must say whether
  the reconciler close converges with the Slice 4 command or stays a raw writer.

- **[Major] The fence requirement closes split-brain on value but not on
  validation.** Fence option (c) — "repair-converged blind write with crash/race
  tests" — guarantees convergence under concurrency but not validation parity.
  Raced-writer tests catch timing divergence, not steady-state semantic
  divergence (legacy raw writer accepts a value the new command rejects). For a
  state-machine-constrained field like `state`, blind-write convergence can
  durably commit a value the new validation forbids. Red flag #2's "different
  validation" clause survives wherever a slice picks option (c) for a
  validation-critical field.

- **[Major] No backlog slice is proven independently shippable-and-revertible,
  and slices 4 and 6 collide on one file.** Lane question 3 is unanswered: the
  design never asserts which slice ships+reverts alone or proves a slice does not
  silently require the next. `cmd/gc/session_reconciler.go` carries the
  drained-close (Slice 4 territory) and the reconciler lifecycle facts (Slice 6
  territory) and is governed by the worker-boundary migration — a three-way
  overlap. The "old readers tolerate new fields / new readers tolerate old fields
  during rollback" tests are also scoped only to the three named files, so a slice
  that adds a session-owned field family elsewhere (e.g., Slice 3 wake metadata)
  can ship new metadata with no reader-tolerance proof — a hidden-flag-day path on
  revert.

- **[Minor] Slice 0's freshness gates silently constrain the concurrent
  worker-boundary migration.** `TestSessionBoundaryInventoryFresh`, the
  shrink-only `SetMetadata*` guard, and `API_CLI_ROUTE_INVENTORY.yaml` pin
  path+symbol. The in-flight worker-boundary migration moves session symbols
  (e.g., `materializeNamedSessionWithContext` behind `worker.Handle`), tripping
  these gates on the *other* migration's PRs. The design does not say who keeps
  the inventories green during concurrent migration.

**Missing evidence:**
- No stated worker-boundary end-state for `session_resolution.go` (does it move
  `materializeNamedSessionWithContext` while leaving the resolver?), so the
  reviewer cannot confirm Slice 1 and the worker-boundary change are
  symbol-disjoint. The code says they are not.
- No declared expectation of whether the worker-boundary migration completes
  before, during, or after this refactor's slices, despite a shared entry point.
- No anti-drift artifact (shared precedence table, or classifier-vs-inline parity
  test on `match_vectors`/`result_kind`) for the Slice 1 coexistence window.
- No per-field "window closes" rule: the shrink-only guard freezes the writer set
  but does not force a coexistence window shut, so a validated session-command
  writer and a legacy raw writer of the same field can coexist indefinitely as
  long as no new one is added.
- No per-slice code-level revert proof (revert slice N with N+1 absent → builds,
  parity green, no dangling provisional type).

**Required changes:**
- Add a "Shared-Resolver Sequencing" subsection treating
  `resolveSessionTargetIDWithContext` as one entry point shared by read-only and
  `materialize:true` modes. State the ordering rule relative to the worker-boundary
  migration in *this* document (not deferred to a per-slice row), and require a
  diff-level test asserting Slice 1's change does not alter the materialize/create
  path (`session_resolution.go:259`/`:328`).
- Require Slice 1 to make the extracted classifier the **single source of
  resolution precedence**: the frozen materialize/create path must consume the
  same classifier for steps 1–7 and own only the terminal create, OR mandate a
  parity test that the inline materialize precedence and the new classifier
  produce identical `match_vectors`/`result_kind` until that surface is migrated.
- Replace the "production API and CLI callers" scope of close-alignment with "all
  production close writers," and enumerate the reconciler/sweep/parallel raw close
  sites (`session_reconciler.go:206`, `session_beads.go:1737/2161`,
  `session_lifecycle_parallel.go:1823`) as callers the Slice 4 Atomic Command
  Contract must characterize and converge — so Slice 4 retires the split-brain
  instead of adding to it.
- Constrain fence option (c) to field families where legacy and new writers are
  proven to share validation (or the legacy writer is demoted to run the new
  validation); otherwise require CAS/token. Add a validation-parity proof to the
  migration row whenever a legacy raw writer coexists with the new command.
- Extend the rollback reader-tolerance test requirement to **every** slice that
  adds or changes a session-owned field family, not only slices touching the three
  named files. Add a per-field "window closes" rule (once a field moves behind a
  command, the guard forbids any remaining production raw writer of that field).
- Declare, per backlog slice, whether it ships and reverts independently and what
  proves it does not require the next slice; explicitly resolve the Slice 4 vs
  Slice 6 overlap on `session_reconciler.go`. Add a Slice 0 ownership note for
  keeping the boundary/route inventories green during the concurrent
  worker-boundary migration.

**Questions:**
- Does the worker-boundary migration retire the `session_resolution.go`
  direct-create (`:328`) before, concurrently with, or after Slice 1's read-only
  extraction? Who owns sequencing the two efforts on this shared function?
- When the `materialize:true` surface is frozen "characterization only," does it
  keep its own inline precedence copy (the fork) or immediately consume the new
  classifier for resolution steps (which contradicts "characterization only")?
- Does close-alignment to `worker.Handle` extend to the reconciler/sweep close
  paths, or do they stay on raw `ClosePatch` after Slice 4? If they stay, what
  reconciles the two validation regimes on the same close fields?
- For a field with a coexisting legacy blind-writer, is validation divergence
  from the new command acceptable during the window, and what bounds the damage
  before the raw-writer retirement condition fires?
