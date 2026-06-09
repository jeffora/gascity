# Priya Menon - Claude

**Verdict:** approve-with-risks

Reviewed `plans/core-gastown-pack-migration/requirements.md` (current
`updated_at: 2026-06-09T17:23:58Z`) against `requirements.schema.md`. Lane:
required Core loading, pack registry, import resolution, legacy import
retirement. This is a *requirements* artifact, so I judge whether the
pack-resolution product contract is pinned tightly enough that an
implementation plan cannot drift — not how it is coded. Findings are grounded in
the current tree (`internal/builtinpacks/registry.go`,
`internal/hooks/hooks.go`, `internal/config/compose.go`, `cmd/gc/cmd_prime.go`).

**Top strengths:**
- **One canonical Core identity, all consumers must close over it.** The
  Problem Statement pack-resolution contract (lines 56–68) plus AC2 and AC3
  establish exactly one Gas City-owned Core identity — the release-bundled
  payload sourced from `internal/packs/core`, with
  `internal/bootstrap/packs/core` demoted to migration-input-only — as a
  *required system layer* that no user/city/rig/root/public/cache/overlay/
  transitive import "may replace, alias, shadow, or silently satisfy." AC2
  enumerates every consumer class (embed, builtin registry, materialization,
  hook/import, generated/synthetic layouts, tests) that must close over that
  authority or be classified non-runtime. This maps exactly onto the real
  consumer set: `registry.go:53` (the bundled `core` entry, today still
  pointing at `internal/bootstrap/packs/core`), `hooks.go:21` (`core.PackFS`
  consumer), the `.gc/system/packs/core` materialization at `cmd_prime.go:330`,
  and the synthetic cache in `registry.go`. This directly kills the "multiple
  Core source-of-truth" and "shadow Core" red flags.
- **Legacy retirement is explicit and fallback-closed.** AC4 + AC5 + the
  retired-import negative path + the offline/stale edge cases retire implicit
  Maintenance and in-tree `examples/gastown`, force fresh `gc init --template
  gastown` to import public `gascity-packs/gastown` at an immutable `sha:` pin,
  and forbid any fallback to in-tree examples, system packs, or synthetic
  aliases even offline. Retirement is bounded in reality — only
  `registry.go:18-19,56-57` consume the in-tree gastown/maintenance FS — so
  AC5's source-consumer-closure is tractable, and the public repo URL it names
  matches `builtinpacks.PublicRepository`.
- **Runtime and doctor are structurally coupled per scenario row.** AC3 binds
  *both* "normal runtime behavior" *and* "bootstrap diagnostic behavior" to a
  shared stable condition code in each matrix row, so within a row runtime
  resolution and doctor cannot disagree about Core, and a bootstrap-only
  diagnostic mode (AC11) can run with no packs resolved.

**Critical risks:**
- **[Major] bd/dolt cardinality and the no-conflict baseline layer order are
  never stated, leaving "Core + Gastown + bd + dolt all participate"
  under-defined.** The doc calls bd/dolt "provider-conditioned support packs"
  (W6H "How", line 79; AC3, line 109) and AC3 requires the matrix to define
  "provider-pack cardinality and mutual-exclusion or co-activation policy for
  each `bd`/`dolt` condition," but the *conditions themselves* are never given:
  which store provider activates bd vs dolt, whether they are mutually exclusive
  (one per beads provider) or co-active, and what a default non-Gastown city
  loads. In the tree today bd/dolt are bundled side-by-side in
  `registry.go:54-55` and I found no auto-import conditioning in
  `internal/config/compose.go`, so this is unresolved product policy, not a
  documented mechanism. This migration's own support artifact,
  `support/maintenance-asset-classification.md` ("Working Decisions Needed" #4),
  goes further and proposes folding Dolt under bd as the owning provider pack
  ("stop treating Dolt as a separately selected required pack") — a different
  model than the requirements' two-pack framing. Q3 (deterministic order when
  Core, Gastown, bd, dolt participate) cannot be made deterministic without
  settling this, AC17 explicitly forbids implementation inferring product
  policy, yet the matrix author would have to invent exactly this. Compounding
  it, Open Questions is "None" (line 148), asserting readiness while a material
  resolution rule — and a decision the support evidence flags as open — is
  unresolved.
- **[Major] Cross-surface consistency of Core determination is implied, not a
  testable criterion.** AC3 requires init, doctor, import-state, CLI load, and
  runtime resolution to "use the same condition-code registry and
  source-attribution model … for the same broken city state," and couples
  runtime↔doctor *within* a row — but nothing *requires* that, for the *same*
  broken city, a runtime load failure surfaces the same condition code and
  identity attribution an operator gets from `gc doctor`, nor that init writes
  config the matrix resolves to that same canonical identity. A shared registry
  is a shared vocabulary, not a tested agreement on the same decision. The red
  flag "required Core behavior is inconsistent between init, doctor, and runtime
  load" is therefore only partially defended.
- **[Minor] bd/dolt source authority/provenance is unspecified.** Core receives
  a named canonical source root, trusted provenance, release identity, and
  content digest (AC2/AC3); bd/dolt get cardinality but no equivalent
  source-authority/provenance contract — even though they are bundled from
  `examples/bd` and `examples/dolt` (`registry.go:54-55`), the same in-tree
  `examples/` location the Problem Statement indicts for ownership ambiguity.
  For deterministic resolved-config provenance of every matrix participant,
  bundled support packs need their source authority pinned (even if the answer
  is "optional bundled packs, no required-layer identity").
- **[Minor] AC2's "dev/test escape hatch" boundary is asserted but worth
  hardening.** AC2 permits "a bounded dev/test escape hatch if tests need to
  construct partial configs" and says it "is available only to native test
  paths and cannot drive production CLI, doctor, controller, runtime, session,
  dispatch, formula expansion, or city-state mutation." That is good — the
  enumeration is concrete. The residual risk is that "native test paths" is a
  build/identity claim with no stated enforcement witness; without a test that
  proves the hatch is unreachable from a production binary, it remains a
  potential bypass of the "Core is required" guarantee. Recommend an explicit
  proof obligation, not just the prose bound.
- **[Minor] Retiring the synthetic public alias removes the current offline
  fresh-init fallback without an explicit replacement scenario.**
  `registry.go:106-133` (`syntheticPackLayouts`/`publicSubpathForPack`)
  deliberately serves a synthetic `gascity-packs/gastown` alias from embedded
  content "when the network is unavailable during init or doctor repair."
  AC4/AC16 retire that path and require fail-closed on cache miss, but the
  Example Mapping covers only an offline *existing* Gastown city (line 96) — not
  fresh offline `gc init --template gastown` with no seeded cache, a real
  operator-facing behavior change this lane should pin with its own scenario.

**Missing evidence:**
- Whether bd and dolt can be simultaneously active or are strictly
  one-per-provider, and what a default non-Gastown city loads (needed for
  deterministic ordering), and whether the bd-owns-dolt consolidation from the
  support classification doc is in or out of scope.
- The baseline (all-healthy) layer precedence among required Core, the provider
  pack(s), and the explicit Gastown import — AC3 enumerates participants and
  *conflict* rows but never states the no-conflict total order, and the AC3
  participant list reads ambiguously as either a precedence ranking or a
  coverage checklist.
- Whether a failed runtime load emits the SAME stable condition code as `gc
  doctor` for the same missing/duplicate-Core city.
- Whether bd/dolt keep `examples/` as their source root or receive Core-style
  provenance treatment in resolved config.
- An enforcement witness proving the AC2 dev/test escape hatch is unreachable
  from a production binary.
- The intended offline UX for fresh `gc init --template gastown` once the
  synthetic alias is gone (fail-closed + cache-seed vs. any sanctioned path).

**Required changes:**
- State the bd/dolt activation condition, mutual-exclusivity, and default-city
  cardinality, plus the no-conflict baseline layer precedence, in AC3 — or move
  the decision to Open Questions (so readiness is not falsely asserted) and
  reconcile with `maintenance-asset-classification.md` "Working Decisions
  Needed" #4 so the support evidence and the requirements do not disagree.
- Add an explicit, testable cross-surface consistency criterion: init, doctor,
  import-state, CLI load, and runtime resolution agree on Core identity and
  condition code for the same concrete city.
- State the source-authority/provenance expectation for bundled support packs
  bd/dolt, even if it is "remain optional bundled packs from `examples/`, no
  required-layer identity."
- Clarify whether the AC3 participant enumeration is a normative precedence
  order or a coverage checklist; if normative, state where explicit user Gastown
  imports rank relative to provider-conditioned support packs for same-named
  assets.
- Add an Example Mapping edge case (or extend AC4/AC16) for fresh *offline*
  Gastown init with no seeded cache: fail closed with a missing-cache
  diagnostic plus seeding guidance, explicitly not the retired synthetic alias.

**Questions:**
- Are bd and dolt ever active together, or strictly one per beads provider, and
  what does a default non-Gastown city load? Is the bd-owns-dolt consolidation
  in scope?
- For a healthy Gastown city (Core + provider pack + public Gastown), what is
  the authoritative layer precedence for same-named keys?
- Must a failed runtime load and `gc doctor` emit identical condition codes and
  source attribution for the same broken city?
- Do bd/dolt move out of `examples/`, or keep that root while only
  `examples/gastown/*` is retired, and do they get a provenance contract?
- Is offline fresh `gc init --template gastown` expected to fail closed
  (requiring cache seeding), confirming full removal of the embedded
  synthetic-alias init/doctor fallback in `registry.go:106-133`?

**Schema conformance (bead-requested):** Conforms to `requirements.schema.md`.
Front matter is correct (`phase: requirements`, `status: draft`, no
`requirements_file`/`design_file`, no bead IDs); all six top-level sections
appear in the required order; W6H covers every dimension; Example Mapping has
happy/negative/edge rows with an evidence column; ACs are testable with a
Verification column; Open Questions is `None`. Two [Minor] hygiene nits: (1) the
Problem-Statement pack-resolution-contract paragraph (56–68) leans on
implementation/architecture vocabulary ("resolved config," "materialized path,"
"lock/cache provenance," "matrix rows") that edges past the schema's "product
language"; (2) naming exact `support/*.yaml` artifact paths in several ACs is
acceptable as binding acceptance evidence (AC17 makes them binding contracts),
but it is the closest the doc comes to prescribing structure and an AC1 reviewer
should confirm it does not trip the schema's no-implementation-files rule. The
substantive readiness concern is the "Open Questions: None" claim while the
bd/dolt provider-conditioning rule (Major #1) is unresolved.
