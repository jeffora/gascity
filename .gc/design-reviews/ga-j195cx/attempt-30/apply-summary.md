# Apply Summary

Source synthesis: `.gc/design-reviews/ga-j195cx/attempt-30/synthesis.md`

Global verdict: `block`

Applied verdict: `iterate`

## Changes Applied

- Made `AcceptedCompileArtifact` the only durable-write proof, with compiler-only minting through `AcceptCompileResult`, identity verification, and an unexported seal.
- Expanded the caller/raw-consumer migration contract with a grep-derived manifest covering `gc formula show`, `gc formula cook`, `cook --attach`, sling, orders, fanout, convergence, API predicates, convoy scans, and dashboard state.
- Reworked validation coverage into independent construct identity, construct location, contribution path, caller path, and requirement-shape dimensions, including unsupported/impossible rows and generator self-tests.
- Tightened convergence migration around accepted artifacts, canonical projection, retry artifact reuse, downgrade behavior, metadata keys, and retirement/static-guard coverage for the legacy subset parser.
- Added pack provenance and compatibility rules for all pack origins, `requires_gc` rejection before selection/staging, compiler requirement increases as pack compatibility events, and requirement-diff reporting.
- Added executable diagnostic parity fixtures, grouped dashboard failure state, and cross-surface preservation rules for CLI, API, generated TypeScript, dashboard, and Event Bus projections.
- Added docs rollout skeletons, required TOML examples, stale-guidance gates, legacy-version policy, alias-window reports, optional legacy probe rules, and future capability guardrails.

## Unfixable Items

None. All blocker and major synthesis findings were addressed in the design document.

## Verification

- Reviewed the generated diff against `design-before.md`.
- Ran a consistency search for durable `CompileResult` misuse in the edited design.
- No code tests were run because this attempt changed only the design document and workflow artifacts.
