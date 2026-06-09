# Elena Marchetti - Claude

**Verdict:** block

Lane: session lifecycle ownership, external metadata write audit, patch-map
escape hatches, CI guard coverage. This reviews the current `DESIGN.md` (the
attempt-16 `iterate`-response revision), `REQUIREMENTS.md`, and
`internal/session/AGENTS.md`. Every code claim re-verified against this checkout;
citations inline. Scope: the active global verdict is `block` and this document
authorizes only a non-mutating Slice 0, so my disposition governs whether the
*mutation-boundary apparatus Slice 0 must build* is sound, not whether any writer
moves yet. Slice 0 *evidence-gathering* may proceed; Slice 0 *close* and every
mutation slice that depends on it cannot, for the reasons below.

I block rather than approve-with-risks (the prior iteration's call) for one
reason: the apparatus' single structural forcing function is still missing after
being raised against attempt 15 and not addressed by attempt 16. The design keeps
accreting descriptive schema while the one mechanism that converts "shrink-only"
into an actual ratchet remains absent. All three of my lane's red flags are now
live in current code.

**Top strengths:**
- The build-then-fence ordering is still correct for my dimension. Slice 0 is
  explicitly non-mutating (DESIGN.md:161-164), the Mutation Ownership Ledger is a
  hard precondition for every mutating slice with a thorough per-row schema
  (DESIGN.md:361-376: exact path+function, store/helper method, session-targeting
  proof, key family + literal key *or* dynamic-key source, top-level field, owner,
  intended owning slice, exception reason+expiry, persistence-error handling), and
  read-only classification (slices 1-2) precedes the write-heavy slices.
- The second enforcement plane is the right idea for the part a static scanner
  structurally cannot reach: the session-owned key-family list must be
  machine-readable and "shared between static guards and runtime bridge denial"
  (DESIGN.md:390-391), and the generic API/CLI bead-mutation bridge is named as a
  denial point (DESIGN.md:376,403). Runtime denial keyed on the family list is the
  correct answer for dynamic-key batch writes a static guard cannot resolve.
- The repair rule is stated firmly: a read path "may not silently repair
  session-owned keys unless it is the named repair owner for that key family and
  propagates persistence errors" (DESIGN.md:412-416), and the read-only classifier
  must return `repair-needed` instead of mutating (DESIGN.md:283-295). The
  principle directly governs the live `RepairEmptyType` leak.

**Critical risks:**
- [Blocker] **Shrink-only still has no forcing function; recorded expiry/retirement
  is cosmetic, so the existing baseline becomes a permanent bypass.** I grepped the
  whole document for any build-failing expiry rule. The validator-failure lists —
  the attempt-16 `SLICE0_CONTRACT` close gate (DESIGN.md:66), the Slice 0 validator
  list (DESIGN.md:210-216), and the `SLICE0_CONTRACT.yaml` meta-check
  (DESIGN.md:177) — fail on missing/skipped/build-tagged/zero-match/stale-path/
  absent-negative-fixture/unproven-`SESSION-*` conditions, but **none** fails when a
  ledger/exception row passes its stated expiry, or when a retirement condition is
  met while the writer still exists. The shrink-only guard itself only "must fail on
  new external `SetMetadata*` ... unless an exact expiring row exists"
  (DESIGN.md:377-381): it blocks the *next* writer and freezes today's baseline. The
  expiry/retirement words appear ~15 times (lines 66, 68, 314-equiv 372, 381, 485,
  528, 545, …) exclusively as *row fields*, never as a *build assertion*. Against
  today's baseline — 21 exported patch constructors, 12+ external patch-apply sites,
  14 `RepairEmptyType` sites, plus raw lifecycle/identity `SetMetadata*` writers
  (below) — a guard frozen here guarantees the bypass is permanent. This is
  red-flag #3 arriving through omission of any ratchet, and it is the same Blocker
  raised against attempt 15, still open verbatim after a full iteration.
- [Major] **The patch surface is wide open in current code and the design still
  refuses to commit to type-level closure.** `type MetadataPatch map[string]string`
  (`internal/session/lifecycle_transition.go:19`) is a transparent alias, so any
  caller can forge a lifecycle patch with a raw map literal and bypass every
  constructor. There are 21 exported `*Patch` constructors
  (`lifecycle_transition.go`: `ClosePatch:397`, `RetireNamedSessionPatch:444`,
  `SleepPatch:283`, `ArchivePatch:372`, `QuarantinePatch:462`, `BeginDrainPatch:259`,
  …). External application is current, not hypothetical, and broader than the prior
  iteration captured: `cmd/gc/session_reconciler.go:206` and
  `cmd/gc/session_lifecycle_parallel.go:1823` do `for k,v := range
  sessionpkg.ClosePatch(...)`; `cmd/gc/session_beads.go:1502` ranges over
  `RetireNamedSessionPatch`; and the **API layer** retires a configured identity at
  `internal/api/session_resolution.go:171`. Direct construct/apply sites total 12+
  (`session_beads.go:444,514,1502,1737,2161`; `session_reconciler.go:206,1858,1917`;
  `session_sleep.go:319`; `session_lifecycle_parallel.go:1823,2511`;
  `session_resolution.go:171`). The design requires the guard to cover
  "patch-constructor use" and "helpers that return patch maps" (DESIGN.md:378-380,
  400) but never commits that slices 3-5 **unexport** the constructors / make
  `MetadataPatch` opaque (applied only by a session-owned method) as their
  retirement proof — the only durable, type-checked closure, and the one
  `internal/session/AGENTS.md` ("call session-owned command APIs instead of applying
  patch maps", AGENTS.md:28-30) already implies. `TestSessionBoundaryGuard` resolves
  to **0 files** (only `TestGCNonTestFilesStayOnWorkerBoundary`,
  `cmd/gc/worker_boundary_import_test.go:11`, exists), so today only prose sits
  between a new file and a raw lifecycle write. Red-flag #2, realized in current
  code.
- [Major] **`RepairEmptyType` write-on-touch is an unbounded repair exception
  firing on 14 ordinary read/command/mail paths, with no named owner and no
  `REQUIREMENTS.md` sanction.** `session.RepairEmptyType` (`internal/session/resolve.go:222-228`)
  calls `store.Update(b.ID, beads.UpdateOpts{Type: &t})` and **swallows the
  persistence error** (`_ = store.Update(...)`), the exact opposite of the design's
  "propagates persistence errors" rule (DESIGN.md:416). It fires from 14 external
  sites including mail resolution (`internal/mail/beadmail/beadmail.go:640,676`),
  API list/command paths (`internal/api/handler_beads.go:89`,
  `internal/api/handler_sessions.go:455,513,725`,
  `internal/api/huma_handlers_sessions_command.go:426,483,876,931`), and CLI
  wake/pin/resolve/name-lookup (`cmd/gc/cmd_session_wake.go:72`,
  `cmd/gc/cmd_session_pin.go:114`, `cmd/gc/session_resolve.go:196`,
  `cmd/gc/session_name_lookup.go:443`). The design governs only the Slice 1
  classifier path (DESIGN.md:283-295); it never names `RepairEmptyType`, these 14
  callers, or the "audited repair owner" it invokes four times (DESIGN.md:295,353,
  412-416) as ledger rows, and `REQUIREMENTS.md` has **no** scenario row authorizing
  auto-heal-on-touch (I read the full ledger; nearest is `SESSION-LIFE-*`, none of
  which sanction repair-on-read). A mail read silently mutating a bead's `Type` and
  discarding the error is precisely the under-owned side effect the design forbids
  in principle but leaves unenforced in fact.
- [Major] **The static guard's matcher is still unspecified, and the only feasible
  in-tree pattern cannot separate session writes from the many non-session
  `SetMetadata*` writes in the same files.** The proven static pattern is the
  symbol/site string-contains denylist in `cmd/gc/worker_boundary_import_test.go`.
  Applied to `SetMetadataBatch` it would over-match the abundant non-session writers
  in the same packages (`gc.routed_to`, `workflow_id`, generic `close_reason`/
  labels) because `store.SetMetadataBatch(session.ID, batch)` carries a
  dynamically-built `batch` whose keys a static guard cannot resolve. The design
  conflates a shrink-only *site inventory* with a *key-family-semantic* guard
  without stating which `TestSessionBoundaryGuard` is. Runtime bridge denial
  (DESIGN.md:390-391) covers the generic bridge but not direct `SetMetadataBatch`
  calls in `cmd/gc` (e.g. the inline wake rollback batch below). Without committing
  the guard to "forbid the exported session-patch symbols + an inventory of named
  raw-writer sites; key-family column is descriptive, not the matcher," the guard is
  under-specified to build.

**Missing evidence:**
- The backlog (DESIGN.md:788-812) names zero concrete call sites and zero inventory
  IDs (grep-confirmed). The slice->writer map lives only in the not-yet-created
  ledger's "intended owning slice" field, so there is no reviewable mapping showing
  which slice retires the external `ClosePatch`/`RetireNamedSessionPatch` apply-sites
  (incl. `session_resolution.go:171`), the 14 `RepairEmptyType` sites, the inline
  wake-rollback batch (`cmd/gc/cmd_session_wake.go` ~84: raw `SetMetadataBatch`
  writing `state`, `state_reason`, `pending_create_claim`,
  `pending_create_started_at`, `wake_request`, `wake_requested_at` straight to the
  store *after* the session-owned `WakeSession`), the raw `sleep_reason` write
  (`cmd/gc/cmd_stop.go:329`), or the two `session_name` identity-assignment writers
  (`cmd/gc/session_name_lookup.go:227`, `cmd/gc/session_beads.go:1119`).
- No build-enforced consequence for `now > expiry`, no maximum-row count, no
  per-release ratchet — nothing makes the inventory monotonically shrink.
- The "audited repair owner" for empty-type repair / runtime-identity backfill is
  invoked but never named to a package/function.
- No owning slice for identity *assignment*: slices 1-2 are read-only and slice 4 is
  identity *retirement*, so the two `session_name` writers above are unhomed.

**Required changes:**
1. Add "ledger/exception row past its stated expiry" and "row whose retirement
   condition is met while its writer still exists" to the Slice 0 validator-failure
   list (DESIGN.md:210-216) and the `SLICE0_CONTRACT` close gate (DESIGN.md:66),
   turning recorded expiry into an enforced forcing function. Without it,
   shrink-only guarantees a permanent bypass for the baselined sites. (Prior-
   iteration RC#1, still open.)
2. State explicitly what `TestSessionBoundaryGuard` matches: a symbol/site guard
   (following `worker_boundary_import_test.go`) that forbids new external references
   to the session-patch surface — `MetadataPatch`, the exported `*Patch`
   constructors, and the runtime-identity/alias helpers — plus an inventory of named
   raw `SetMetadata*` session-key writer sites; clarify the key-family column is
   descriptive ownership metadata, not the matcher, and that dynamic-key batch
   writes are fenced at runtime via the bridge denial list (DESIGN.md:390-391).
3. Commit in the backlog / Atomic Command text that slices 3-5 retire the external
   patch-apply sites and then **unexport** the `*Patch` constructors / make
   `MetadataPatch` opaque (applied only by a session-owned method) as their
   retirement proof — or justify why a static scanner must remain the sole guard,
   including how it catches raw `map[string]string` literals.
4. Name `session.RepairEmptyType` and its 14 external callers (including the two
   `internal/mail/beadmail` sites) as Mutation Ownership Ledger rows with an intended
   owning slice and a named repair owner that propagates the `store.Update` error,
   and record in `REQUIREMENTS.md` whether write-on-touch repair is intended product
   behavior (scenario row + owner) or a migration crutch (retiring slice).
5. Give each mutation-touching backlog slice a concrete completion criterion naming
   the write sites it deletes or routes (file-level suffices): the wake-rollback
   batch under "Explicit wake," the `ClosePatch`/`RetireNamedSessionPatch`
   apply-sites (incl. `session_resolution.go:171`) under "Close and identity
   retirement," and add/extend a slice that owns the two `session_name`
   identity-assignment writers.

**Questions:**
- When a ledger/exception row passes its expiry with the writer still present, does
  the build fail, or is expiry advisory? Unspecified across three iterations, and it
  alone decides whether the bypass is temporary or permanent.
- Is the end-state to unexport the `*Patch` constructors and make `MetadataPatch`
  opaque once external apply-sites move behind session-owned commands, or to bless
  "external caller ranges over an exported patch" as permanent? `AGENTS.md` implies
  the former; the code and backlog imply neither.
- Is `RepairEmptyType` firing (and error-swallowing) on ordinary
  wake/pin/resolve/mail/API-list paths intended auto-heal-on-touch or a migration
  crutch, and which named repair owner / backlog slice owns retiring the 14 sites?
