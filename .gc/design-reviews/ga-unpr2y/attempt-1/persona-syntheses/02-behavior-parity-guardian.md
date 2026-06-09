# Natasha Volkov

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Slice 0's parity gate is not yet mechanically enforceable. The proof command must fail when expected validators are absent, skipped, build-tagged out, or matched by zero tests; otherwise the parity apparatus can go green without running the checks it claims to require.
- [Blocker] Active `SESSION-*` rows already cite stale or missing proof. The reviewers all call out `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`, including missing paths such as `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`.
- [Blocker] Evidence repair is not tied tightly enough to same-behavior proof. Repointing a stale row to a passing selector is insufficient unless the selector is executable, current, and asserts the same product behavior the requirement row records, or an owner-approved amendment/retirement explicitly changes that behavior.
- [Major] Scenario parity is not yet mapped to implementation slices. Later slices need a reviewable row-to-slice, row-to-surface, and row-to-proof map so a migration bead cannot cite too few scenario rows or move behavior without the right oracle.
- [Major] `RepairEmptyType` and `repair-needed` handling are behavior-sensitive and under-specified. Query-side adopters must preserve or intentionally amend current successful lookup behavior for repairable empty-type sessions without smuggling writes back into the pure classifier.
- [Major] First-adopter classification work is too broad unless it names exact routes, operation IDs, legacy/Huma variants, generated-client implications, and caller-level parity tests for status, body shape, aliases, closed lookup, configured-name rejection, path alias handling, transcript, pending, and stream paths.
- [Major] Characterization requirements need to be hard gates for both mutating and read-side extractions. Tests must prove user-visible or system-level behavior before delegation and run unchanged against the refactored path before duplicated logic is removed.

**Disagreements:**
- Claude and DeepSeek return `block`; Codex returns `approve-with-risks` for design shape only, with explicit limits to non-mutating Slice 0 decomposition. I choose `block` because the parity lane cannot approve behavior-moving decomposition until stale proof, validator self-checks, and scenario allocation are enforceable.
- DeepSeek argues that reconciler evidence repair cannot live entirely in non-mutating Slice 0 and proposes a transition proof allowlist backed by historical commits. Claude and Codex require repair, replacement, or owner-approved retirement before stale rows can satisfy parity. My assessment: a transition allowlist may explain why Slice 0 starts, but it must be time-bounded and cannot authorize any dependent behavior-moving slice without live executable proof or an owner-approved amendment.
- DeepSeek wants a Scenario Allocation Matrix directly in `DESIGN.md`; Claude and Codex are comfortable with machine-readable Slice 0 artifacts if they are validated. My assessment: the location is flexible, but the row-to-slice and row-to-surface mapping must be reviewable before any behavior-moving work depends on it.
- DeepSeek requires a concrete adapter behavior for `repair-needed` now; Claude and Codex require hard endpoint-specific parity gates before adoption. My assessment: the design must at least define route-specific adapter contracts before the classifier reaches a caller; any chosen behavior must have positive and negative parity fixtures.
- DeepSeek demands black-box characterization rules; Claude and Codex frame the same risk through same-behavior proof and caller-level tests. I assess the stricter black-box/system-level wording as the safer acceptance bar.

**Missing evidence:**
- Definitions for the named Slice 0 validator tests, plus a meta-check proving they are present, selected by the proof command, not skipped, and not hidden behind build tags.
- Positive, negative, stale-path, zero-match, and skipped-test fixtures for `TestScenarioParityFreshness`, `TestSlice0Contract`, and related validators.
- Live executable proof or owner-approved retirement/amendment for every active stale row, not only the four named `SESSION-RECON-*` rows.
- Confirmation that `SESSION-WORK-001` through `SESSION-WORK-004` selectors execute today rather than merely pointing at an existing file.
- A single authoritative relationship between `REQUIREMENTS.md` Evidence and `SCENARIO_PARITY.yaml` current-oracle fields: equality, explicit supersession, or validator-enforced precedence.
- Durable owner-approval artifacts for behavior-changing requirement edits, including changed row IDs, approval authority, artifact IDs, amendment state, expiry if temporary, and row content hashes or equivalent drift detection.
- A complete row-to-slice, row-to-surface, row-to-route, and row-to-proof mapping for every active `SESSION-*` requirement before behavior-moving work begins.
- Exact first-adopter route inventory and adapter wrapper contracts for target classification, including how `repair-needed` is interpreted without mutating inside the classifier.
- Characterization test selectors for read-side reconciler fact extraction, runtime start, and any other non-mutating extraction that still changes behavior routing.

**Required changes:**
- Make the Slice 0 proof command self-validating: fail when any named validator symbol is absent, skipped, build-tagged out, or matched by no tests; require each validator to include stale and negative fixtures.
- Require every active `SESSION-*` row to execute a current test/symbol selector or carry an owner-approved retirement/amendment; repaired evidence must assert the same product behavior unless the owner-approved record says otherwise.
- Add assertion-level freshness requirements: selectors should be test-name or symbol based, must resolve to compiled and executed tests, and must fail on zero matches or unconditional skips.
- Create a reviewable scenario allocation artifact mapping each active row to backlog slice, caller surfaces, exact route or operation where relevant, current oracle, characterization proof, amendment state, and proof command.
- Define the precedence between `REQUIREMENTS.md` Evidence and `SCENARIO_PARITY.yaml` oracle fields so the two sources cannot silently diverge.
- Specify route-specific `repair-needed` adapter behavior before the first classifier adopter lands, with fixtures showing whether successful empty-type lookup is preserved through raw metadata plus separate repair or intentionally changed by owner-approved amendment.
- Make the first target-classification adopter exact: name the handler or route group, legacy and Huma surfaces, generated-client effects, consumed scenario rows, and unchanged caller-level parity tests.
- Extend characterization gates to read-side reconciler fact extraction and runtime start, and require black-box or system-level assertions for exit codes, stdout/stderr, API status/problem shape, generated-client shape, and committed bead/session state.
- If transition allowlists are used for missing reconciler proof, make them explicit, time-bounded, visible in requirements, and insufficient by themselves to authorize any behavior-moving slice.
