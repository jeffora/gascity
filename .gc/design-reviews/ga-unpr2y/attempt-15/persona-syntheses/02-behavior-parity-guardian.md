# Natasha Volkov

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Slice 0's parity gate is not yet mechanically enforceable. Codex verified that the design's published proof command can return success while most named validators are absent or report no matching tests; Claude says the first freshness run against the current ledger has no defined hard outcome; DeepSeek says the same missing evidence makes the gate circular once validators exist.
- [Blocker] Active requirements rows already cite stale or missing proof. All three reviewers call out `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`, including missing `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`. A behavior-preserving refactor cannot treat these rows as fresh oracles.
- [Blocker] Scenario parity is still not bound to executable, row-scoped behavior proof. Codex rejects static selectors and commit citations as sufficient parity proof; DeepSeek requires exact test-function and non-skipped assertion freshness; Claude finds a dual-source-of-truth risk between `REQUIREMENTS.md` Evidence and the future `SCENARIO_PARITY.yaml` oracle fields.
- [Blocker] The owner-approval rule is policy text, not an enforceable mechanism. Claude requires durable approval artifact IDs and row hashing to prevent silent behavior edits; Codex requires enumerated amendment states with closure semantics; DeepSeek's stale-proof transition proposal still needs formal approval and visibility to avoid normalizing missing evidence.
- [Major] Target classification adoption needs surface-specific parity, canonical vocabulary, and adapter contracts before behavior moves. Claude flags the new `kind` taxonomy as a parallel vocabulary that must map to canonical identity projections and agree with `ProjectLifecycle`; Codex says the first adopter is too broad and must name exact routes; DeepSeek says pure-classifier integration needs wrappers so caller-owned materialization, repair, wake, and policy side effects are not dropped.
- [Major] Characterization-test requirements are not strong enough for regression prevention. Claude asks for explicit per-surface close conditions, Codex requires runnable characterization proof for rows touched by behavior-moving slices, and DeepSeek requires black-box or system-level old-vs-new checks rather than white-box helper assertions.

**Disagreements:**
- There is no verdict disagreement: Claude, Codex, and DeepSeek all return `block`.
- The reviewers disagree on the stale-evidence repair path. Claude and Codex require repair, replacement, or owner-approved retirement before stale rows can close parity; DeepSeek proposes a temporary transition allowlist that can fall back to historical commits. My assessment: historical commits can remain supporting context, but they must not satisfy current behavior parity for a behavior-moving slice unless an owner-approved amendment explicitly retires or rewrites the row.
- The reviewers frame the Slice 0 failure mode differently. Codex observes a fail-open proof command today because named validators are absent; DeepSeek warns the future validator will fail closed immediately because required evidence paths are missing; Claude focuses on the design not saying which outcome is mandatory. My assessment: both fail-open and circular-fail risks must be designed out with a self-validating proof command plus an explicit stale-ledger repair outcome.
- DeepSeek requires a design-time Scenario Allocation Matrix in `DESIGN.md`, while Claude treats the planned `SCENARIO_PARITY.yaml` as the right freshness guard if it is pinned and validated. My assessment: the exact file can be flexible, but approval requires a reviewable row-to-slice, row-to-surface, and row-to-proof mapping before implementation relies on Slice 0.
- DeepSeek calls adapter wrappers a blocker; Claude and Codex express the same risk through vocabulary, route-specific behavior, and surface parity gaps. My assessment: wrapper contracts are required before the classifier is exposed to any migrated caller.

**Missing evidence:**
- Definitions for the named Slice 0 validator tests, plus a meta-check proving they are present, selected by the proof command, not skipped, and not hidden behind unsatisfied build tags.
- Positive and negative/stale fixtures for each Slice 0 validator, especially `TestScenarioParityFreshness`.
- Live executable proof or owner-approved retirement/rewrite for `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`.
- A complete row-to-slice, row-to-surface, and row-to-proof mapping for every active `SESSION-*` requirement before behavior-moving work begins.
- A single authoritative relationship between `REQUIREMENTS.md` Evidence and `SCENARIO_PARITY.yaml` current-oracle fields: equality, explicit supersession, or a validator-enforced precedence rule.
- Durable approval artifacts for behavior-changing requirements edits, including changed rows, approval authority, artifact IDs, and validator-checkable coverage.
- A `kind` to canonical identity-projection mapping for the target classifier, with parity proof against `ProjectLifecycle` for closed, historical, conflict, ambiguous, configured, and reserved-unmaterialized states.
- Route- and surface-specific characterization inventories for the first classifier adopter, including API status/problem shape, transcript and pending behavior, aliases, configured names, closed lookup, materialization, repair, wake, and caller policy side effects.
- Adapter wrapper specifications that identify which layer interprets classifier output and which layer performs any allowed mutation after classification.

**Required changes:**
- Make the Slice 0 proof command self-validating: fail if any named validator symbol is absent, skipped, build-tagged out, or matched by no tests; require each validator to include positive and stale/negative fixtures.
- Repair the stale reconciler evidence as Slice 0 work before any dependent behavior-moving slice: restore tests, repoint rows to live executable proof, or retire/rewrite rows through owner-approved requirements amendments. If a temporary transition allowlist is used, make it explicit, time-bounded, and insufficient by itself to authorize behavior-moving parity.
- Collapse or pin the dual oracle sources. Require every active `SESSION-*` row to have parity metadata, and require the parity oracle to match or explicitly supersede the row's `REQUIREMENTS.md` Evidence path under validator control.
- Make owner approval mechanically enforceable with durable artifact IDs, row-level coverage, enumerated amendment states, and row content hashing or equivalent change detection for required-behavior edits.
- Add a reviewable scenario allocation artifact mapping every active `SESSION-*` row to backlog slice, caller surfaces, current executable oracle, characterization proof to add or keep, and exact proof command.
- Make the first target-classification adopter route-specific. Name the first handler or route group, the consumed `SESSION-*` rows, the required route-inventory rows, and the unchanged caller-level tests that must remain green in the migration commit.
- Add the target-classifier `kind` to canonical projection mapping, and require parity against `ProjectLifecycle` for identity states before Slice 1 adopts the classifier.
- Define adapter wrapper contracts for all classifier callers so the pure classifier never performs mutations and is never exposed directly where caller-specific materialization, repair, wake, or error-policy side effects are required.
- Tighten characterization-test rules: tests must assert user-visible or system-level outputs, run against the legacy baseline before delegation, and run unchanged against the refactored path before duplicated caller logic is deleted.
