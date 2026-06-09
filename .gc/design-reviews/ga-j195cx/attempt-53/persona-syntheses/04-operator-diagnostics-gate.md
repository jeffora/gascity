# Marta Hidalgo

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The operator-facing diagnostic contract is not yet implementable without drift. Codex finds that the core diagnostic model cannot carry both formula requirement attribution and host gate attribution, while later wire/event/dashboard contracts require distinct `requirement_source_*` and `host_source_*` fields. Claude independently flags that `host_source_path` and `host_source_key` are operator-critical and must be required for host-capability failures. Without one typed source-attribution model, CLI, API, dashboard, and event surfaces will reconstruct remedies from strings or side channels.
- [Major] Background diagnostic grouping is internally incomplete. Claude flags unbounded structured logs, reset on unrelated `config_generation` changes, in-memory occurrence counters that lose meaning on restart, and missing grouped dashboard rollups. Codex flags a contradiction between "one append-only event" and an event payload that somehow reaches occurrence count `3`, plus missing storage for grouped failure history. The design must pick a concrete producer-state and projection model.
- [Major] Warning suppression is specified with conflicting keys. Claude notes the CLI suppression key is coarser than the canonical warning `OnceKey`; Codex finds three different key definitions across policy and projection sections. This can hide distinct warnings or duplicate the same warning depending on which paragraph an implementer follows.
- [Major] Surface-parity evidence is too fragmented for disabled-host failures. Both reviews accept the intended parity goal, but the design does not yet require one golden row that pins direct CLI, API-routed CLI, HTTP API body/status, generated TypeScript/dashboard state, and applicable event payloads together.
- [Major] Disabled-host operator remediation is under-specified at shell and API boundaries. Claude flags that CLI exit codes make operator-fixable host/config failures indistinguishable from author-fixable parse failures and sometimes from internal/process failures. Claude also questions HTTP 409 retry semantics; the design needs a non-retryable operator-remediation signal that API clients can route without parsing prose.
- [Major] Fatal-plus-warning combinations need fixture coverage. Claude identifies the dual-declared plus disabled-host case where `formula.compiler_requirement_unsatisfied` and `formula.contract_deprecated` must both appear, in deterministic order, while background failure events project only the fatal diagnostic.
- [Minor] Dashboard grouping is intent, not contract. Claude asks for an operator-fixable versus author-fixable category or generated mapping, grouped host-capability rollups across many affected formulas, and defined clearance semantics after the host capability becomes satisfying. Codex similarly notes that grouped background failure history has no persistence location.
- [Minor] Several operator boundaries remain open: direct interactive `gc sling --formula` event emission, 4096-entry LRU sizing/configurability, canonical message/remediation bytes, and accepted-artifact warning persistence rate.

**Disagreements:**
- Claude returns `approve-with-risks`; Codex returns `block`. My assessment is `block` because Codex's source-attribution objection is foundational and Claude's host-source and surface-parity concerns reinforce it.
- Claude treats the disabled-host exit-code problem as a required design change; Codex does not raise exit codes. My assessment: exit-code parity is not the foundational blocker, but it is required before this persona can approve operator diagnostics.
- Claude proposes bounding background logs, changing reset keys, persisting/recomputing counters, and adding dashboard rollups. Codex focuses on the append-only event contradiction and undefined producer-state owner. My assessment: these are the same issue and should be solved as one durable grouping contract, not as separate surface patches.
- Claude wants `host_source_path` and `host_source_key` required on selected wire diagnostics. Codex wants the core diagnostic and host capability types extended first. My assessment: the core typed attribution model should be the source of truth, and the wire requirements should be generated/projected from it.
- Claude questions whether 409 is the right HTTP status for an unsatisfied host capability; Codex does not. My assessment: the exact status is less important than pinning non-retryable operator remediation semantics in the typed body or headers.
- Gemini was absent, which is allowed by the bead contract.

**Missing evidence:**
- No fixture proves `formula.compiler_requirement_unsatisfied` preserves formula requirement source and host source from one typed diagnostic object without parsing strings.
- No concrete registered event payload type is shown for formula diagnostics, so it is unclear whether events and API/dashboard payloads share a source of truth or duplicate fields by hand.
- No repeated-loop test proves grouped order/controller failures avoid Event Bus spam, avoid per-tick structured logs, expose occurrence state deterministically, and behave coherently after restart.
- No test covers a config reload that advances generation without changing the relevant formula compiler host capability.
- No single disabled-host golden fixture pins CLI stderr, CLI exit, HTTP status/body, API-routed CLI behavior, generated TypeScript/dashboard state, and applicable event payloads in one row.
- No checked-in canonical message/remediation fixture is identified for byte-level parity.
- No dual-declared plus disabled-host fixture proves fatal-plus-warning co-emission and ordering.
- No dashboard fixture pins operator-fixable versus author-fixable grouping, grouped host-capability rollup behavior, or clearance after remediation.
- No explicit boundary states whether direct interactive `gc sling --formula` emits a registered failure event or only synchronous stderr and exit status.
- No rationale or configurability contract is given for the 4096-entry suppression LRU.

**Required changes:**
- Add a reusable typed source-attribution model to the core diagnostic path. It must represent primary source, requirement source, and host source, including path, key, value, and source kind where relevant. `HostCapabilities` should carry structured source attribution rather than only a display string.
- Make CLI output, API/Huma payloads, registered event payloads, dashboard/generated TypeScript fixtures, and `OnceKey` generation derive from that same diagnostic object. Add a fixture asserting disabled-host diagnostics preserve both requirement and host sources without string parsing.
- Require `host_source_path` and `host_source_key` for `formula.compiler_requirement_unsatisfied` and `formula.host_capability_invalid`, and assert them in the disabled-host fixture.
- Define the durable background grouping contract: storage owner, grouping key, occurrence update semantics, append-only event behavior, structured-log suppression, dashboard/API projection, restart behavior, and reset behavior based on relevant host-capability changes rather than raw config-generation churn.
- Reconcile warning suppression on one key, preferably canonical `OnceKey` plus command invocation for CLI surfaces, and remove the shorter conflicting code/path/key rule.
- Extend the diagnostics fixture schema so each fatal row can carry CLI stderr, CLI exit, API status/body, generated TypeScript/dashboard representation, and applicable event payloads in the same golden row.
- Decide and fixture the CLI exit behavior for operator-fixable host/config diagnostics across validation/report commands and runtime launch commands.
- Pin HTTP retry/remediation semantics for `formula.compiler_requirement_unsatisfied`, either through a typed remediation class/header or an equivalent body field, and document the chosen status code.
- Add canonical message/remediation fixtures consumed by tests for every diagnostic code.
- Add dual-declared disabled-host coverage proving `formula.compiler_requirement_unsatisfied` and `formula.contract_deprecated` co-emit in the documented order, with only the fatal projected into background failure events.
- Specify dashboard diagnostic categories or a generated code-to-category mapping, grouped host-capability rollups across multiple affected formulas, and clearance semantics after the host capability becomes satisfying.
- State the direct interactive `gc sling --formula` event-emission boundary, the LRU sizing/configurability rationale, and whether accepted-artifact warning persistence is deduped by immutable formula content.
