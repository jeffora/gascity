# Marta Hidalgo

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] Host capability provenance is not end-to-end pinned strongly enough for the diagnostic contract. Codex finds that the required constructor still accepts only `enabled bool` plus a source string, which cannot deterministically carry source kind, path, raw value, position, or config generation. Claude treats normalized requirement handling as a strength, but still identifies missing omitted-default and downgraded-host provenance fixtures. The operator surface cannot be considered stable until typed provenance is supplied by the edge and asserted through CLI, Huma JSON, generated TypeScript, dashboard state, and Event Bus payloads.
- [Major] Cross-surface parity is proven for the main `disabled-v2-host` launch failure, but not for the full operator diagnostic surface. Claude requires a same-identity artifact downgrade fixture and API-routed CLI warning rows. Codex requires conflict-status fixtures that split validation/preview 400 from launch 409. These gaps leave CLI, API-routed CLI, dashboard, and event projections free to diverge in cases operators will actually troubleshoot.
- [Major] Background producer cadence and warning persistence are underspecified. Claude flags undefined summary emission cadence, ambiguous scan-series wording, config-generation resets that can re-emit deprecation warnings on reload churn, and undefined LRU eviction behavior. Codex separately flags the absence of an accepted-warning write path and fixtures for polling producers. Without a precise grouping and suppression lifecycle, the design can either flood operator logs or lose durable warning evidence.
- [Major] Host downgrade with a valid same-identity artifact lacks an operator-facing projection contract. Claude finds that the design permits reuse after host downgrade but only pins the new-launch-fails remediation. That leaves the CLI and dashboard free to contradict each other about whether the existing root may continue under its persisted artifact.
- [Major] `formula.compiler_requirement_conflict` status handling is internally inconsistent. Codex finds one table assigning HTTP 400 while the final operator-surface decision says launch-time unsatisfied or conflicting requirements use HTTP 409. This must be split by operation, as with unsatisfied requirements, or API-routed CLI and dashboard launch flows can disagree.
- [Minor] Interactive CLI deprecation warning behavior remains noisy by policy but has no operator control. Claude notes that repeated separate commands will re-emit the same warnings, with no quiet or once-style option documented for CI loops that only need the exit code.

**Disagreements:**
- Claude verdict is `approve-with-risks`; Codex verdict is `block`. My assessment is `block` because the provenance constructor gap can make typed operator diagnostics non-deterministic across surfaces, which directly violates this persona lane.
- Claude views `SourceKind`, the compatibility matrix, and default-capability equivalence fixtures as strong evidence for host-capability handling. Codex argues that the required API still cannot carry the structured source fields the diagnostics contract promises. My assessment: requirement normalization may be correct, but the edge-to-wire provenance path still needs a structured input and golden fixtures.
- Claude emphasizes warning flood risks from summary cadence, config reload, cross-process CLI invocations, and LRU eviction. Codex emphasizes durable grouped-warning evidence for background producers and alias-removal gates. My assessment: both are required; the design needs one explicit key shape, lifecycle, write order, and projection contract for accepted warnings.
- Claude requires multi-producer dashboard grouping into one operator card for the same disabled-host condition. Codex does not call out that grouping shape. My assessment: the design should keep one canonical remediation where the operator action is identical, but preserve producer and subject details as typed subfields.
- Kimi 2.6 was absent. That is allowed by the bead contract and does not affect the synthesis verdict.

**Missing evidence:**
- A structured host capability source type or constructor that carries enabled state, source kind, source path, source key, raw source value, source position where available, and config generation.
- Golden fixtures for disabled-host diagnostics across omitted `formula_v2`, explicit `formula_v2 = false`, explicit true, deprecated `graph_workflows`, test override, and config reload boundaries.
- A `disabled-v2-host-with-artifact` fixture covering CLI, API/Huma JSON, API-routed CLI, dashboard state, and remediation text for same-identity artifact reuse after host downgrade.
- Projection fixtures proving `formula.compiler_requirement_conflict` returns 400 for validation/preview and 409 for launch, with direct CLI and API-routed CLI preserving the same typed body.
- Background warning fixtures for accepted legacy aliases across repeated scans, process restart, config reload, host capability toggle, LRU eviction at the 4097th distinct key, dashboard projection, and zero warning Event Bus events.
- A fixture or matrix row for API-routed CLI rendering of `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse`.
- A multi-producer disabled-host fixture showing order dispatch, convergence, and controller validation grouped under one canonical operator remediation while retaining producer-tagged details and stable occurrence fields.
- Same-branch updates to stale reference and context docs that still teach `version` and omit `[requires]`.

**Required changes:**
- Replace or extend `HostCapabilitiesFromFormulaV2(enabled bool, source string)` with a structured edge input carrying source kind, path, key, raw value, source position, and config generation. Make disabled-host diagnostics consume only that structured data, then fixture-lock the fields across all operator-visible surfaces.
- Add disabled-host fixture coverage for omitted default, explicit false, explicit true, deprecated alias promotion, test override, config reload, and same-identity artifact reuse after host downgrade.
- Define one background producer cadence: first occurrence appends a typed event or accepted warning group as appropriate; later matching occurrences update grouped state without appending until the documented key changes. Pin how scan ticks, config generation, host capability toggles, producer policy, and LRU eviction affect that key.
- Specify the accepted-warning persistence path: whether warnings reuse `FormulaDiagnosticGroupState` or a separate typed state, the write order relative to creating accepted work, and the fields dashboard/reporting use for alias-removal gates. Add polling-producer fixtures for the full lifecycle.
- Align `formula.compiler_requirement_conflict` status semantics so validation and preview use 400, launch uses 409, and direct/API-routed CLI paths preserve the same diagnostic body.
- Pin remediation text for host-downgrade-with-artifact and for omitted-default versus explicit-false provenance, even if the remediation action remains the same.
- Add the API-routed CLI warning projection row and fixtures for `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse`.
- Update the reference docs and architecture/proposal context in the same stack before shipping any user-visible diagnostic surface.
