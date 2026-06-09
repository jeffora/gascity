# Kwame Asante - Claude

**Verdict:** approve-with-risks

*(Lane: minimal vocabulary, facade creep, event-log deferral, backlog scope
control. This review is against the current 692-line `DESIGN.md` — the
Attempt-15 `iterate` response, not the 472-line Attempt-14 text a prior wave
reviewed; line refs below are to the current file. From my lane the direction is
sound: the design has genuinely absorbed anti-facade and anti-event-sourcing
rules, and Attempt 15 added a universal-vs-per-slice-preflight split (L143-148)
that begins to fix the inflation a prior wave predicted. The hard `block` in the
global synthesis comes from the parity, mutation, and atomic-command lanes, not
mine. Two of the risks below are the design violating its **own** stated rules,
which I weight as Major; they are tightening, not a redesign.)*

**Top strengths:**
- Anti-facade discipline that matches my mandate end to end: "Prefer small
  operation-specific APIs over a broad `SessionService`" (L603), per-operation
  shape — target classifier / wake decider / close decider (L607-611), the
  `Avoid: one large SessionFacts struct` / generic command bus list (L613-618),
  the Non-Goal "Do not introduce a large facade before one small operation
  proves value" (L684), and `AGENTS.md` reinforcement (L35-36). No `SessionFacts`
  mega-struct exists. Lane question 3 is answered by the document.
- Event sourcing is explicitly deferred — the core of my event-log mandate:
  events are "post-commit facts ... not commands, locks, durable truth, or the
  only recovery mechanism" (L499-500), durable scans own critical convergence
  (L514, L519-520), Non-Goal "Do not make event sourcing the first implementation
  step" (L687), and Slice 0 only *inventories* current `session.*` events
  (L521-528) instead of building new payloads. Lane question 2: well-handled by
  deferral; no current API is shaped for a future durable log.
- A real, enforceable vocabulary lifecycle (`documented`/`private`/`provisional`/
  `delegating`, L587-599) that keeps speculative shared types out of public API,
  generated clients, and event payloads "until a production caller proves the
  exact field set" (L594), plus `VOCABULARY_CHECKPOINTS.yaml` (L165) and the new
  per-slice-preflight mechanism (L143-148). The right control mechanisms exist;
  the problem is the document does not consistently point them at its own new
  contracts.

**Critical risks:**

- **[Major] The Slice 1 Target Classification result schema carries sub-fields
  the documented first-adopter precedence never exercises — speculative
  vocabulary landing before a caller proves it (red flag #1).** The eight-step
  query-side precedence (L222-244) resolves by id / session_name / alias / title
  / open-vs-closed status plus the configured-name and path-alias rules. It never
  consumes `bead_state.labels` or `bead_state.lifecycle state` (L261); it does
  not need a `config_state.materialization allowed flag` separate from the
  `reserved-unmaterialized`/`config-orphan` flags it already carries (L262); and
  a pure in-memory classifier over already-fetched beads has no
  `diagnostics.stale or partial fact marker` (L263) — staleness/partiality is
  runtime/reconciler vocabulary for later slices. This contradicts the design's
  own provisional-until-proven rule (L594). The `result_kind` enum (6 values) and
  `vector_kind` enum (8 values) are *not* the problem — each maps to a precedence
  row the first adopter actually hits — so the fix is surgical: classify the
  sub-fields, not rebuild the taxonomy.

- **[Major] The result schema reads as a flat optional envelope — the exact shape
  the document's own rule bans.** L599: "Flat optional envelopes are not
  acceptable for new shared types; use tagged result kinds or per-kind structs
  when only some fields are meaningful." Yet the typed result (L253-264) lists
  `result_kind` beside `match_vectors[]`, `bead_state`, `config_state`,
  `diagnostics`, and `terminal_error` as co-resident fields. For
  `result_kind=not-found`, what are `bead_state`/`config_state`? For `selected`,
  what is `terminal_error`? For `store-error`, what is `match_vectors`? That is a
  flat struct with kind-dependent nil fields — precisely what L599 forbids. This
  is the first concrete artifact the refactor builds and every downstream surface
  inherits its shape, so the self-contradiction is load-bearing. The contract
  must declare per-kind structs / a tagged union, or justify the flat shape
  against its own rule.

- **[Major] Universal Slice 0 front-loads mutation-only inventories that only
  slices 3-6 consume — premature shared infrastructure for a read-only first
  slice (backlog scope control).** Slice 1 is read-only and mutates nothing
  (L210-218). Yet universal Slice 0 (L150-165) still requires
  `COMMAND_APPLIERS.yaml` (inventory of command-like *writers*), the
  mutation/destructive rows of `BOUNDARY_MATRIX.yaml` (drain, repair,
  destructive-action safety, L161/L477-494), `WORKER_BOUNDARY_EXCEPTIONS.yaml`
  (mutating-lifecycle exceptions, L163), and a `DIAGNOSTICS_MANIFEST.yaml`
  spanning deciders/commands (L164) — none of which the read-only Slice 1 reads.
  Attempt 15 deserves credit for deferring the *operation-specific* COMMAND_APPLIERS
  rows to per-slice preflight (L160) and inventing the split (L148); the gap is
  that the split is applied to *rows inside* artifacts but not to the *artifacts
  themselves*. As written this is the YAGNI anti-pattern — shared inventory before
  two consumers exist, the `rule-of-two` the design itself cites (L165) — and it
  cuts against "This is not a new architecture program" (L12), "No premature
  abstraction" (settled decision), and L684. Existential risk for a *small*
  refactor: Slice 0 becomes a long speculative-contract program whose only
  delivered output is inventory, and "prove value with one thin caller first"
  never lands.

- **[Minor] `SCENARIO_PARITY.yaml` requires a row for *every* active `SESSION-*`
  scenario in Slice 0 (L158), broader than Slice 1's touched rows** (the
  identity/targeting cluster: SESSION-ID-003/004/007/008/009). A universal parity
  ledger is a defensible baseline, so I keep this Minor and note parity
  *completeness* is partly the behavior-parity lane's concern — but from pure
  scope control, Slice 1 only needs parity for the rows it touches.

**Missing evidence:**
- `TR-007` is named in my lane brief ("Does TR-007 future durable-event
  compatibility shape current APIs ...?") but appears nowhere in `DESIGN.md` or
  `REQUIREMENTS.md`, which use `SESSION-*` IDs exclusively. I cannot evaluate a
  durable-event-compatibility requirement that is absent from the behavior source
  of truth. Either it is stale brief vocabulary, or a real requirement is missing
  an ID and a row.
- The exact result-schema field subset the Slice 1 API-query adapter actually
  produces is not shown in `DESIGN.md`; it is deferred to the
  `TARGET_CLASSIFICATION_CONTRACT.yaml` "first-adopter rows" (L159) which are not
  visible here. The prose presents the full schema as though all of it is the
  Slice 1 commitment.
- No statement classifying the Target Classification Contract itself under the
  Vocabulary Lifecycle states (`provisional` upper-bound vs `delegating`
  production commitment). The whole inflation question turns on this.

**Required changes:**
- Classify the Target Classification Contract under the Vocabulary Lifecycle:
  mark the schema (and its kind-dependent sub-fields) `provisional` upper-bound,
  and state that Slice 1 implements only the `result_kind`s, `vector_kind`s, and
  fields its read-only API-query adopter actually produces (drop/defer
  `bead_state.labels`, `bead_state.lifecycle state`, the redundant
  `materialization allowed flag`, and the `stale/partial` marker until a slice
  proves them). This converts a "spec to build now" into a "sketch to draw from"
  without deleting the design thought.
- Resolve the flat-optional-envelope contradiction (L599 vs L253-264): declare
  per-kind result structs / a tagged union, or explicitly note the table is a
  provisional field census and production uses per-kind structs.
- Apply the universal-vs-per-slice split to the *artifacts*, not just their rows:
  move `COMMAND_APPLIERS` writer inventory, the mutation/destructive
  `BOUNDARY_MATRIX` rows, `WORKER_BOUNDARY_EXCEPTIONS`, and the decider/command
  `DIAGNOSTICS_MANIFEST` rows into the per-slice preflights of the mutating slices
  (3-6) that read them; reduce universal Slice 0 to what Slice 1 consumes.
- Resolve or remove `TR-007`. If durable-event compatibility is a real product
  rule, give it a row in `REQUIREMENTS.md`; otherwise confirm the target stays
  in-process best-effort events plus durable-scan recovery, with no API shaped for
  a future durable log.

**Questions:**
- Is the classification result one struct or per-kind structs / a tagged union?
  (Determines whether L599 is satisfied.)
- For Slice 1 (read-only API query), which result sub-fields does the adapter
  actually consume to preserve current `writeResolveError`/`humaResolveError`
  wire output?
- What is `TR-007`, and does any current requirement demand durable-event
  compatibility — or is in-process best-effort + durable-scan recovery the only
  target shape?
- Can Slice 0 close with *only* the Slice-1-consumed artifacts, deferring the
  mutation/command/boundary inventories to the slices that move those operations?
  If not, what makes those inventories load-bearing for a read-only first slice?
