# Marta Hidalgo - Claude

**Verdict:** approve-with-risks

**Top strengths:**
- Structured `DiagnosticAttribution` with separate `Primary`, `Requirement`,
  `Host`, and `Pack` source kinds (and the `DiagnosticSourceKind` enum that
  distinguishes `host_capability` from `formula_requirement`) cleanly separates
  operator-fixable host failures from author-fixable parse errors. The
  `disabled-v2-host` golden fixture is named as a single source of truth that
  drives direct CLI stderr, CLI exit code, API-routed CLI, Huma JSON, generated
  TypeScript, dashboard state, and `order.formula_compile_failed` payload —
  so cross-surface remedy drift is fixture-locked, not aspirational.
- Cadence controls are spelled out per producer with explicit
  `FormulaDiagnosticBurstBudget` fields (`Window`, `InitialEvents`,
  `MaxEventsPerWindow`, `SummaryEvery`, `ResetOnConfigGeneration`, etc.).
  Warning codes never publish Event Bus events; order dispatch caps at one event
  per `(order id, formula, OnceKey, host capability, config generation)` per
  scan series with a 15-minute summary; CLI suppression is per process
  invocation; warnings persist in the accepted artifact rather than being
  recomputed on every resolve. This directly answers the "every resolve cycle"
  concern.
- The compatibility matrix, normalization rules, and `RequirementSource`
  provenance guard make the host gate operate on the normalized capability,
  not the source spelling. Both legacy `contract = "graph.v2"`, dual-declared,
  and `requires`-only formulas funnel through one
  `formula.compiler_requirement_unsatisfied` diagnostic when
  `[daemon] formula_v2 = false`, and `TestRequirementSourceProvenanceOnly`
  blocks production code from branching on the source enum.

**Critical risks:**
- [Major] **Canonical-string lock covers warnings but not fatal codes.** The
  "Warning messages and remediation are canonical strings across CLI,
  API-routed CLI, Huma JSON, generated TypeScript fixtures, dashboard state,
  reports, and accepted artifacts" table (Host Capability And Diagnostics
  section, lines ~3962–3970) enumerates only `formula.contract_deprecated`,
  `formula.version_deprecated`, and `formula.version_misuse`. The fatal
  `formula.compiler_requirement_unsatisfied` remediation ("enable
  `[daemon] formula_v2` or choose a v1 formula") is locked only by the
  `disabled-v2-host` fixture and the projection-matrix prose. There is no
  explicit "no surface may append local advice" sentence for the fatal codes
  as there is for warnings, so a future dashboard or CLI patch can quietly
  append a docs link, a "(click to enable)" affordance, or producer-local
  phrasing without breaking the fixture *and* without breaking the parity
  contract. This is the most likely path for dashboard/CLI to give
  contradictory remedies. The same "canonical strings, no surface may append"
  rule must extend to every fatal code in the projection matrix.
- [Major] **`RollupClass` assignment rule is implicit.** `RollupClass` has a
  closed enum (`host_config`, `formula_source`, `pack_source`, `internal`) and
  the dashboard grouping rule says "group by operator-fixable host/config
  diagnostics separately from author-fixable formula/pack diagnostics." But
  the mapping from `Diagnostic.SourceKind` (and from each diagnostic code) to
  `RollupClass` is never spelled out as a table or fixture row. Without an
  explicit `{code → RollupClass}` mapping locked in a generated fixture,
  `formula.compiler_requirement_unsatisfied` could end up in
  `formula_source` on one surface and `host_config` on another — which is
  exactly the red flag of "disabled-capability errors look like parse
  failures" recurring at the rollup layer rather than the diagnostic layer.
- [Major] **Conflict precedence hides the host gate.** Line ~1130 specifies
  "Host satisfaction runs only when normalization produced a usable
  requirement and no fatal requirement syntax, type, axis, or conflict
  diagnostic exists." That is correct for ordering, but it means an operator
  with both a dual-declaration conflict *and* a disabled host fixes the
  conflict, re-runs, and only then discovers the daemon flag is off — two
  remedy-discovery cycles for one underlying state. Diagnostics should at
  least surface a *non-blocking informational note* ("host capability is
  currently `1`; this formula will also need `[daemon] formula_v2 = true`
  after the conflict is resolved") when the gate is known disabled and a
  fatal author diagnostic is being returned. Otherwise the design is
  technically correct but operationally a remedy-chain trap.
- [Minor] **Unsupported-store diagnostic is unnamed.** The diagnostics design
  requires `EnsureSingletonByMetadata` / `SetMetadataCompareAndSwap` and says
  "an implementation that cannot provide the atomicity must report an
  unsupported store diagnostic and the producer must fail closed before
  protected writes." But no diagnostic code is reserved for this state. If an
  operator hits this on an older bead-store backend they will see an unnamed
  internal error rather than the actionable remedy ("upgrade the bead store
  or change `[beads].store`"). Add a `formula.diagnostic_store_unsupported`
  (or similar) code to the projection matrix with its own remediation and
  surface-parity fixture.
- [Minor] **Diagnostic-subject lifecycle when an order is deleted.** The
  order-dispatch diagnostic-subject singleton (`gc.diagnostic_kind =
  formula-diagnostic-group`, `gc.subject_id =
  order:<scope>:<scoped-name>`) is great, but the cleanup contract
  ("a group is cleared only when the subject is closed, the formula/config
  source changes, the diagnostic no longer occurs on the next successful
  scan, or an operator invokes the producer's normal cleanup path") does not
  cover the case where the *order itself is removed* from city config. In
  that case the next scan never re-touches the subject and the rollup row
  remains visible forever. Either the order-dispatch loop must garbage-collect
  diagnostic subjects whose order definition no longer exists, or
  `expired` retention must explicitly cover this. Without that, dashboards
  accumulate stale rows after order churn.
- [Minor] **`--warnings=once` is mentioned but underspecified.** Final
  operator-surface decisions row says "launch commands keep fatal diagnostics
  visible and may use `--warnings=once` for CI logs." There is no row
  defining what `once` means (per `OnceKey`, per code, per source key, or
  per command invocation), how it interacts with `--quiet`, or whether it
  applies to `gc order` and `gc formula cook` in addition to `gc sling`.
  Operators will tune this in CI logs and silently get different behavior
  across surfaces if it is not nailed down.

**Missing evidence:**
- No explicit `{diagnostic code → RollupClass}` mapping table or fixture
  reference. The dashboard `host_config` vs `formula_source` grouping
  depends on this and it is the load-bearing UX promise behind the lane.
- No surface-parity fixture row for the metadata-spillover branch
  (`gc.formula_compile_artifact` ref vs inline metadata). The 8 KiB
  spillover boundary is fixture-locked for the *decision* but not for the
  *operator-visible diagnostic projection*: when a host-disabled root has
  spilled-over metadata, does the dashboard rollup load the warning record
  from the artifact, or skip it? The design says it must remain visible
  ("Background producers that accept a legacy alias but suppress display
  still persist the diagnostic in the accepted artifact or producer state.
  After a restart, the next validation/report run must recover the same
  accepted-alias counts by scanning root metadata and accepted artifacts.")
  but I do not see this asserted in the `disabled-v2-host` golden parity set.
- No named diagnostic code for `[daemon].graph_workflows` promotion in the
  diagnostic projection matrix. The host capability provenance fixtures cover
  the `deprecated_graph_workflows` source kind and say a deprecation
  diagnostic is emitted on validation/display surfaces, but the diagnostic
  code and remediation string for that case are not in the projection matrix
  (lines ~3300-3315). An operator who has `graph_workflows = true` and wants
  the canonical config will see an unfamiliar code.
- No retention bound named for `expired` rollup rows.
  `FormulaDiagnosticRollup` has a lifecycle (`active`, `resolved`, `expired`,
  `abandoned`) but the "retention cleanup after no matching subject remains"
  rule has no owner, TTL, or cleanup command. This is a slow leak — bounded
  but unnamed.

**Required changes:**
- Add a "Canonical fatal-diagnostic strings" table (parallel to the existing
  warning canonical-strings table) covering at minimum
  `formula.compiler_requirement_unsatisfied`,
  `formula.compiler_requirement_invalid_syntax`,
  `formula.compiler_requirement_unsupported_future`,
  `formula.compiler_requirement_missing`,
  `formula.compiler_requirement_conflict`,
  `formula.requirement_unknown_axis`,
  `formula.requirement_invalid_type`, and
  `formula.host_capability_invalid`. State explicitly that "no surface may
  append local advice to these strings unless it uses a separate typed
  field." Tie the rule to the existing surface-parity fixtures.
- Add an explicit `{code → RollupClass}` mapping table, generated and
  CI-locked, with `formula.compiler_requirement_unsatisfied` and
  `formula.host_capability_invalid` mapping to `host_config`;
  parse/syntax/type/axis/missing/conflict mapping to `formula_source`; pack
  floor mapping to `pack_source`; the new unsupported-store code mapping to
  `internal`. Add a generator test
  (`TestRollupClassMatchesDiagnosticSourceKind`) that fails when a new
  diagnostic code is added without a `RollupClass` row.
- When the host gate is known disabled and the compiler is returning a fatal
  *author* diagnostic (conflict, missing requirement, unknown axis), append
  a typed informational `Attribution.Host` projection so the operator sees
  the second remedy in the same response. Add a fixture row to
  `disabled-v2-host` (or a parallel `disabled-v2-host-plus-conflict`
  fixture) that asserts both attributions are present.
- Reserve a diagnostic code for the "bead-store lacks CAS / singleton
  metadata" failure mode (e.g. `formula.diagnostic_store_unsupported`),
  add a remediation string, add the projection-matrix row
  (CLI/API/dashboard/events), and add a producer fixture that proves the
  failure leaves zero protected writes.
- Specify diagnostic-subject lifecycle when an order is removed: either
  garbage-collect the diagnostic-group singleton in the order-dispatch
  cleanup pass, or add an `expired` retention rule keyed on
  "subject id no longer references a configured order" with an owner and
  command.
- Specify `--warnings=once` semantics in the same surface-parity table as
  `--quiet`: which commands accept it, what scope (per `OnceKey`), and how
  it interacts with both `--quiet` and JSON output.

**Questions:**
- Should `formula.compiler_requirement_unsatisfied` on the launch path be a
  `409 Conflict` or a `412 Precondition Failed` for API consumers? The
  design picks 409 explicitly ("`412 Precondition Failed` is reserved for
  future conditional requests with explicit precondition headers") but a 409
  conventionally implies retry-after-update on the *same resource*; here
  the operator must update a *different* resource (city config). Worth
  documenting the rationale in `docs/reference/cli.md` and the API
  reference so client libraries do not silently retry the launch.
- Does the `gc formula diagnostics status --json` rollup surface page when
  `MaxChildrenPerRollup=100` is hit? `DroppedChildCount` /
  `ExampleDigest` cover the cap, but the CLI output for an operator running
  the command interactively (no `--json`) does not have a defined
  truncation marker.
- Does config reload while an order has an in-flight scan close the old
  diagnostic-group row to `resolved`, or just freeze it under the old
  `ConfigGeneration`? Lines ~3741–3743 lean toward "old row remains
  historical." For a flapping `formula_v2` flag, this could create a
  spiral of historical rows on the dashboard. Confirm the retention owner
  before Phase 4c.
- For dashboard-less operators consuming `order.formula_compile_failed`
  events directly, is there a documented mapping between the event
  payload's `Attribution.Host` field and the city config key/path? The
  Huma response has typed source attribution; the event payload is
  described as carrying the same `FormulaDiagnosticPayload`, but the
  example event payload schema is not shown — worth a fixture excerpt in
  the design for clarity.
