# Alistair Sterling - Claude

**Lane:** zfc-role-neutrality-guardian (wave 1) — zero hardcoded roles, Core role
neutrality, `dog` configurability, SDK self-sufficiency. Reviewed the current
`plans/core-gastown-pack-migration/requirements.md` (`updated_at`
2026-06-09T17:23:58Z) strictly through the role-neutrality lens, and checked it
against the `gc.mayor.requirements.v1` schema.

**Verdict:** approve-with-risks

The product direction is exactly my mandate, and all three lane questions are
answered with explicit, testable ACs rather than prose promises: controller-only
SDK operation (AC2, AC9, Happy-path 1), configurable `dog` with renamed/replaced/
omitted/disabled cases (AC9 + the executor edge case), and an absence scan over
Go + Core assets + generated/rendered output with positive/negative controls
(AC8 + the role-routing negative path). No requirement introduces role-conditional
Go logic; the document actively prohibits it (AC8, AC9, Out-of-Scope). The
remaining risks are precision/scoping gaps in the enforcement contract —
concentrated where a role-like token is legitimately allowed in Core (the default
`dog` binding), where the neutrality scan is scoped for a Gastown city, and where
two self-sufficiency/completeness claims are asserted without a named test
surface. These should be tightened before implementation but do not block
requirements approval from this lane. Schema conformance: front matter, section
order, W6H, example mapping (happy/negative/edge), testable ACs, Out-of-Scope, and
`Open Questions: None` all conform.

**Top strengths:**
- AC9 makes the maintenance executor (`dog`) explicit pack data and forces
  renamed/replaced/omitted/disabled cases through config + diagnostics, never
  hardcoded role logic ("Go code and SDK infrastructure must treat it as
  user-supplied config"). Directly satisfies the dog-configurability mandate.
- AC2 + Happy-path 1 require the controller to execute SDK infrastructure
  operations with *no* configured Gastown agent, backed by an explicit
  controller-only / no-executor test. Directly satisfies SDK self-sufficiency
  and the "remove a `[[agent]]` and nothing SDK-side breaks" test.
- AC8 + the role-routing negative path establish a denied-token absence scan over
  Go, Core assets, formulas/orders/prompts/overlays, and generated/materialized/
  rendered output, with retired role names permitted only as source-attribution
  text and never as routes, notification targets, formula bindings, prompt
  defaults, overlays, generated defaults, or Go fallbacks. Directly satisfies
  zero-hardcoded-roles, including the generated-output blind spot.

**Critical risks:**
- **[Major] The single permitted `dog` token representation is not reconciled
  across AC8 and AC9.** AC8 permits literal `dog` only at "exact inert
  configured-default pack data keys" and explicitly forbids `dog` as a "formula
  binding," route, prompt default, overlay, generated default, or Go fallback. AC9
  says Core "may ship a default configured executor named `dog` as inert pack
  data" via "symbolic configurable bindings whose required and optional keys are
  declared in pack data." But a default executor must be *wired* to maintenance
  work somehow, and the document never states whether that default wiring is (a)
  the allowed declared pack-data key or (b) a formula binding (prohibited). The
  word "binding" in AC9 collides with the "formula bindings" AC8 prohibits. This
  is the one appearance of a role-like literal allowed in Core, so an
  underspecified target means the scan allowlist either false-positives on the
  legitimate default or false-negatives on a real leak. The neutrality guarantee
  has a soft center until AC8 and AC9 name one mechanism.
- **[Major] SDK self-sufficiency is asserted but its test surface is never
  enumerated, so the guarantee is not falsifiable as written.** AC2 says "Normal
  SDK infrastructure operations must not depend on any Gastown role existing" and
  the evidence is a generic "controller-only operation"/"no-executor
  controller-only" check. "Normal SDK infrastructure operations" is unbounded; a
  single smoke test can pass while a specific operation (gate evaluation, health
  patrol restart, bead lifecycle, dispatch/order dispatch — the operations AGENTS
  itself names) secretly depends on a configured executor or agent. The AC needs
  to name the operation set proven controller-only, or "controller-only" cannot be
  audited.
- **[Major] The denied-token set is not required to derive from an authoritative
  role inventory, so absence-scan completeness is assumed, not proven.** AC8 has
  `role-neutrality-scan.yaml` "define denied tokens," and the role list in the
  negative path is illustrative ("such as Mayor, Deacon, Polecat, Refinery,
  Witness, Boot, Crew, or Gastown"). A real Gastown role omitted from a
  hand-maintained denied-token list would pass the scan and persist in Core
  undetected. Completeness must be tied to the frozen AC6 ledger / AC7 manifest
  role-and-identity inventory rather than an ad-hoc enumeration.
- **[Minor] Neutrality-scan scoping for materialized/rendered output does not
  distinguish Core-owned outputs from legitimately materialized external Gastown
  pack content.** AC8 covers "generated/materialized metadata" and "rendered
  templates" with "scan roots" and "path-scoped allowlists," but does not scope
  those roots to Gas-City-owned Core output. A Gastown city materializes the
  external gastown pack (which legitimately contains Mayor, Deacon, Polecat, etc.);
  if the scan's materialized/rendered roots aren't scoped to Core output, the scan
  either false-positives on legitimate external roles or, if loosened to avoid
  that, can miss a Core leak that shares a path. The negative-path example
  demonstrates only the *catch* direction ("a Core asset routes to a role → fail"),
  never the *allow* direction.

**Missing evidence:**
- No positive/negative control proving the scan *permits* Gastown role names in
  external/materialized Gastown pack content while still denying them in
  Core-owned output. AC8 lists "positive controls" and "negative controls"
  generically but the document only exhibits the deny direction.
- The concrete set of "SDK infrastructure operations" that must pass the
  controller-only / no-executor test is not listed in AC2 or AC9, and the
  provenance of the denied-token inventory (derived vs hand-authored) is unstated.
- No concrete Example Mapping row for a *fully substituted* (different identity,
  not merely renamed/disabled) Core maintenance executor, though AC9 names
  "replaced." The executor edge case only exercises "renamed or disabled."
- The AC9 split between controller-owned maintenance (deterministic/structural/
  safety-critical) and optional LLM-executed maintenance is asserted, but the
  criteria for which side a given behavior lands on are not pinned. The
  "no-executor controller-only" witness proves the system still *operates*, but
  not that the controller-owned partition is itself free of an implicit
  executor/role assumption.

**Required changes:**
- Reconcile AC8 and AC9 to name one representation for the default `dog` binding:
  the exact declared pack-data key (and shape) that is the *only* permitted
  appearance of `dog`/executor tokens in Core, with all other forms (formula
  bindings, routes, prompt defaults, generated defaults, Go fallbacks) explicitly
  denied. Require a positive control for that exact key and a negative control for
  a `dog` route/formula-binding leak, wired into the AC8 allowlist-rot guard.
- Enumerate, in AC2/AC9, the specific SDK infrastructure operations that must be
  proven to run controller-only with no configured executor/agent (at minimum:
  gate evaluation, health patrol, bead lifecycle, dispatch/order dispatch), so
  "controller-only operation check" names a defined, auditable test surface rather
  than a single smoke test.
- Require the AC8 denied-token set and its positive controls to be generated from
  or cross-checked against the authoritative Gastown role and identity inventory
  in the AC6 ledger and AC7 manifest, so absence-scan completeness is provable
  rather than assumed.
- Scope the AC8 scan roots and generated/rendered/materialized coverage to
  Core-owned + Go + Gas-City-generated output, explicitly excluding legitimately
  materialized external Gastown pack content, and add the matching control pair
  (Gastown role in external pack content = allowed; Gastown role in Core output =
  denied) so the scan is correct for the `gc init --template gastown` path.

**Questions:**
- Is the default `dog` executor bound to maintenance work via a declared
  pack-data key (AC8-permitted) or a formula binding (AC8-prohibited)? As written
  the two readings conflict for the one legitimate default.
- Is `dog` a Core-generic maintenance-executor concept or a Gastown role being
  retained in Core? If the latter, why is it exempt from the same migration that
  moves Mayor/Deacon/Polecat to the public pack — and should role-neutral Core
  ship any *named* executor default at all, versus an empty/symbolic binding that
  packs populate?
- For a Gastown city, are materialized external-pack role names in scope for the
  neutrality scan, and if so how are they distinguished from a Core leak?
- (lane-adjacent, non-blocking) Provider-conditioned `bd`/`dolt` support-pack
  selection (AC3): is the condition resolved purely from declarative matrix/config
  data, or could it introduce a Go-side judgment call? Flagging only because
  "infrastructure behavior conditioned on a configured X" is the nearest neighbor
  to my third lane question; it is not a role-neutrality violation as specified.
