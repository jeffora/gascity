# Session Requirements Integrity Reviewer - Claude

**Verdict:** block

Reviewed artifact: `internal/session/REQUIREMENTS.md` (untracked working-tree
"Seed draft"). Schema judged against:
`gc-plan-pack/.../skills/mayor/requirements.schema.md` (`gc.mayor.requirements.v1`).
Review checkout: `HEAD 8db5484c7`, which is **379 commits behind**
`origin/main` (`0dae71e3b`, merge-base `1632db4aa`). That divergence matters for
the evidence lane and is addressed below.

The block is **not** a judgment that the document is low quality — its content,
vocabulary, and evidence are strong (verified below). The block is on a single
load-bearing, in-lane contradiction: this artifact does not, and by the
schema's own text arguably *should not*, conform to `gc.mayor.requirements.v1`,
and that mismatch must be resolved before anyone can certify it "approved
against the schema."

**Top strengths:**
- **Canonical vocabulary is real, not invented.** Every token in the "Canonical
  Vocabulary" section resolves in `internal/session/lifecycle_projection.go`
  (all base states `creating`…`stopped`; desired `desired-running/-asleep/-blocked/undesired`;
  runtime `stale-creating/fresh-creating/start-requested`; identity
  `reserved-unmaterialized/canonical/historical`; blockers
  `missing-config/identity-conflict/duplicate-canonical`; wake causes
  `pending-create/named-always/scale-demand`). `ProjectLifecycle` is a real
  function (`lifecycle_projection.go:374`). This is a strong defense against
  vocabulary drift and satisfies the "no invented lifecycle terms" guard.
- **Evidence actually enforces the claims (against the main frame).** Spot-checked
  rows hold: SESSION-STATE-001/002 match the `state_machine.go` command table
  exactly (create→start-pending; `ready` only from creating; `sleep` only from
  active; `drain` only from active; archive set; close from any non-closed);
  SESSION-RECON-002 matches `scale_from_zero_test.go` (min=0 cold pool, city-store
  routed work, demand clamped to 1); SESSION-RECON-006 matches
  `provider_health_gate_test.go` (fail-open on absent/stale/unknown,
  no-respawn-while-red); SESSION-RECON-007 matches `session_progress_test.go`
  (threshold stall, attached/startup exempt, provider-unhealthy precedence,
  unknown-progress conservative). Rows encode happy + negative + edge behavior,
  as the mandate requires.
- **Clean product/implementation boundary.** No bead IDs, formula targets, or
  workflow-launch instructions leak in; the Maintenance Rules explicitly forbid
  using the file as an implementation-plan scratchpad. This satisfies the
  *separation* half of lane question 1.

**Critical risks:**
- **[Blocker] Wholesale non-conformance to `gc.mayor.requirements.v1`.** The
  document fails essentially every structural requirement of the schema it is
  being judged against: no YAML front matter (`plan_slug`, `phase: requirements`,
  `rig`, `rig_root`, `artifact_root`, `status`, timestamps); none of the six
  required body sections (`Problem Statement`, `W6H`, `Example Mapping`,
  `Acceptance Criteria`, `Out Of Scope`, `Open Questions`); title is
  `# Session Requirements`, not `# Requirements: <title>`; path is
  `internal/session/REQUIREMENTS.md`, not `<rig-root>/plans/<slug>/requirements.md`.
  This directly triggers two of my red flags — "missing required front matter or
  top-level requirements sections" and "a module-local ledger shape replacing the
  schema." A reviewer cannot honestly say this conforms.
- **[Blocker, same root cause] The schema and the artifact are different document
  types, and nothing resolves which one governs.** `gc.mayor.requirements.v1`
  explicitly disclaims this artifact: "It is not an implementation plan,
  **module-local requirements ledger**, task list, or bead-creation payload."
  The document explicitly *is* a module-local ledger: "This document is the
  reconciliation ledger for session behavior … stored beside `internal/session`."
  Per the schema's own rule ("If the current artifact is the wrong path or wrong
  schema, stop with `blocked:wrong-artifact` rather than iterating the
  document"), applying this schema here looks like a category error. Forcing the
  ledger into the Mayor plan-requirements shape (W6H / Example Mapping /
  Acceptance Criteria + plan front matter) would *destroy* its value — it is a
  code↔test↔behavior reconciliation ledger, not a product-planning artifact. The
  contradiction must be settled before approval, not papered over.
- **[Minor] `status: Seed draft` is not a schema-legal status.** The schema's
  status enum is `draft | questions | approved | blocked`. "Seed draft" is a
  symptom of the same schema mismatch (the file uses a free-form ledger status
  table, not the front-matter lifecycle).

**Missing evidence:**
- **Which schema governs this artifact is undefined.** The review premise (judge
  vs `gc.mayor.requirements.v1`) and the document's self-declared identity
  (module reconciliation ledger) are in direct conflict, and neither artifact
  resolves it. This is the central unknown.
- **No reference-frame anchor for the evidence.** The document is an uncommitted
  working-tree file on a branch 379 commits behind `origin/main`. In *this*
  checkout, three cited paths do not resolve — `cmd/gc/scale_from_zero_test.go`,
  `cmd/gc/provider_health_gate_test.go`, `cmd/gc/session_progress_test.go` (and
  the provider-health-gate source is absent; progress logic appears here as
  `cmd/gc/session_circuit_breaker.go`). I verified all three **do** exist on
  `origin/main` with the claimed behavior, so these are *not* fabricated or stale
  citations — they are a behind-checkout artifact. But the document pins no frame
  (no "as of main@<SHA>"), so a future reader on an older or divergent tree
  cannot distinguish "citation rotted" from "my checkout is behind." That is a
  durability gap for an evidence ledger whose whole job is checkable proof.
- **`awake` is used but unlisted.** SESSION-LIFE-001 relies on the legacy compat
  term `state=awake` (behaves as active), but the Canonical Vocabulary base-state
  list omits `awake`, and the projection emits no `"awake"` string. The row is
  correct (`awake` is a legacy *input*), yet the vocabulary section's own
  "avoid inventing parallel terms such as 'live', 'dead', 'enabled', or 'active'"
  guard is undermined by silently leaning on an unlisted legacy term.

**Required changes:**
1. **Resolve the schema-applicability contradiction (the blocker).** Recommended
   resolution: explicitly declare `internal/session/REQUIREMENTS.md` **out of
   scope** for `gc.mayor.requirements.v1` (it is a module behavior ledger, which
   the schema disclaims) and have the design-review route stop applying the Mayor
   schema to it (`blocked:wrong-artifact`). If the project genuinely wants a
   Mayor-style product-requirements doc for the session module, create it
   separately at `<rig-root>/plans/<slug>/requirements.md` and leave this ledger
   intact.
2. **Only if the Mayor schema must govern this file:** add the required front
   matter and restructure into the six mandated sections. Flag explicitly that
   this path discards the ledger's reconciliation value, so it is *not*
   recommended — option 1 is the right fix.
3. **Add a reference-frame anchor** to the ledger (e.g., a one-line "Evidence
   current as of `main@<SHA>`" plus, where useful, commit pins already present in
   the RECON/WORK rows) so every citation is checkable from any checkout and
   "stale vs behind" is unambiguous.
4. **List `awake`** in Canonical Vocabulary as a recognized legacy-compat *input*
   state mapping to active, so the document's own anti-drift guard is internally
   consistent.
5. **Minor wording:** SESSION-STATE-001 says "any non-none state -> closed on
   close"; the code says "any non-closed state → StateClosed." Align the row to
   the code's "non-closed" phrasing (the none→closed edge is the discrepancy).

**Questions:**
- Is `gc.mayor.requirements.v1` *intended* to govern module-local behavior
  ledgers? If yes, the schema text (which excludes them) must change; if no, the
  design-review routing that points the schema at this file must change.
- Should module behavior ledgers get their **own** schema, distinct from the
  Mayor plan-requirements schema? This document is a strong template for one
  (vocabulary pinning + per-scenario evidence rows), and standardizing it would
  remove the recurring "wrong-schema" ambiguity for module-local ledgers across
  the codebase.
