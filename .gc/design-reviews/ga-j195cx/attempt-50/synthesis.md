# Design Review Synthesis

## Overall Verdict: block

Three persona syntheses returned `block`, so worst-verdict-wins requires a global `block`. The design is directionally strong around moving formula compiler authority into `internal/formula`, but it is not yet implementable as an external or operator-facing contract because diagnostics, convergence projection, pack migration, and rollback/recovery semantics remain underspecified.

## Consensus Strengths
- Multiple reviewers praised the central compiler-boundary direction: raw requirement interpretation, host satisfaction, diagnostics, accepted artifacts, and graph-workflow facts belong in `internal/formula`.
- Reviewers broadly agreed the v0 requirement grammar should stay closed, normalize behavior on capability values rather than raw source spelling, and fail closed for unknown future axes.
- The design correctly treats pack revision, refs, SHAs, and lockfile provenance as the ecosystem reproducibility boundary rather than introducing formula-level artifact semver.
- The split between preview results and accepted durable-write artifacts is the right invariant for keeping projections from reimplementing formula decisions.
- The terminology model is close: `[requires] formula_compiler` and `[pack] requires_gc` can work if docs and examples make the distinction explicit.

## Critical Findings

### [Blocker] Operator diagnostics are not a stable wire contract
**Sources:** 04-operator-diagnostics-gate / Claude, Codex; reinforced by 02-contract-migration-guardian, 03-parser-validation-matrix, 08-docs-dx-terminology

**Issue:** The diagnostic projection matrix contradicts the CLI exit-code mapping, and disabled-host failures are not distinguishable enough from author-fixable parse errors. There is no single golden fixture pinning CLI stderr, CLI exit, API/Huma response, API-routed CLI behavior, dashboard/generated TypeScript state, and order failure event payload for the same disabled-host case.

**Required change:** Reconcile command-class exit/status semantics, add byte-normalized golden rows spanning CLI, API, dashboard, generated types, and events, and make host-capability provenance mandatory for unsatisfied or invalid host diagnostics.

### [Blocker] External pack migration outputs are not precise enough
**Sources:** 07-pack-versioning-ecosystem / Claude, Codex; reinforced by 02-contract-migration-guardian, 06-rollout-sequencing-reviewer, 08-docs-dx-terminology

**Issue:** Stable JSON outputs and migration hints do not identify the author-editable import binding that must change. Duplicate bindings, shadowed formulas, transitive imports, and pinned external pack refs can therefore produce remediation that names a source/ref but not the specific binding path, lockfile key, or parent binding to edit.

**Required change:** Add binding identity to validation JSON, migration hints, requirement-diff output, persisted provenance, and any consumed root metadata/artifact; include duplicate-binding, aliased-transitive-import, and shadowed-winner fixtures.

### [Blocker] Convergence still lacks a compiler-owned projection
**Sources:** 09-convergence-subset-reviewer / Claude, Codex; reinforced by 01-compiler-boundary-invariant, 03-parser-validation-matrix, 05-caller-integration-inventory

**Issue:** The design does not define a canonical `internal/formula` representation for convergence fields such as `convergence`, `required_vars`, and `evaluate_prompt` after raw decode, inheritance, and acceptance. Without that, convergence must keep reopening raw formula files, preserve a legacy subset parser, or reconstruct required vars outside the compiler boundary.

**Required change:** Add a normative compiler-owned convergence projection, specify the accepted artifact passed through create/retry/reconcile/missing-child/speculative paths, and require zero durable writes before compile, accept, projection, and convergence validation all succeed.

### [Blocker] In-flight and legacy graph roots can fail closed with no recovery path
**Sources:** 05-caller-integration-inventory / Claude, Codex; reinforced by 01-compiler-boundary-invariant, 04-operator-diagnostics-gate, 06-rollout-sequencing-reviewer, 09-convergence-subset-reviewer

**Issue:** Existing graph roots accepted under capability 2, legacy-only roots, and active v2 wisps can become unrecoverable when `formula_v2` is disabled, source data is missing, a host is downgraded, or a retry/on-complete path needs a new compile. The current design names fail-closed behavior but not an operator repair, abandon, re-stamp, or documented trade-off path.

**Required change:** Define the controller/operator recovery contract for unrecoverable active roots, including downgrade behavior, missing-source behavior, idempotent artifact stamping, zero-write diagnostics, and cross-binary continuation fixtures.

### [Major] Durable-write boundaries still rely on convention
**Sources:** 01-compiler-boundary-invariant / Claude, Codex; 05-caller-integration-inventory / Claude, Codex; 09-convergence-subset-reviewer / Claude, Codex

**Issue:** Store-mutating APIs can still accept raw formula products, caller-built requirements, mutable graph-workflow facts, or local predicates unless CI/static guards make those paths impossible. `Recipe.GraphWorkflow` is load-bearing but not yet protected from caller synthesis.

**Required change:** Make graph-workflow truth compiler-owned, align durable artifact minting with resolver-owned source provenance, add accepted-artifact/write-intent APIs, and enforce static guards for raw metadata reads, durable-writer signatures, convergence legacy-parser use, and allowlist expiry.

### [Major] Parser and validation matrices are not closed enough
**Sources:** 03-parser-validation-matrix / Claude, Codex; 10-future-capability-architect / Claude, Codex; 02-contract-migration-guardian / Claude, Codex

**Issue:** The requirement grammar lacks byte-exact fixture coverage for signed values, zero/lower-than-supported values, leading zeros, overflow-sized integers, tabs/newlines, Unicode whitespace/operators/digits, NUL/control characters, and possibly JSON-only shapes. V2-only construct enforcement also conflicts with existing documentation around `check`.

**Required change:** Add grammar rows and count locks for every suite, define overflow/future-capability recognition byte-exactly, resolve `check` semantics, specify TOML raw-field extraction behavior, and enumerate first-party formulas/examples/fixtures that must be dual-declared or rewritten.

### [Major] Rollout, alias removal, and external-support gates are not executable
**Sources:** 02-contract-migration-guardian / Claude, Codex; 06-rollout-sequencing-reviewer / Claude, Codex; 07-pack-versioning-ecosystem / Claude, Codex

**Issue:** Alias removal depends on artifacts that do not yet define objective expiration criteria, old-reader sources of truth, external-pack status transitions, release-captain authority, or unknown/SHA-pinned pack policy. "Two minor releases" has no calendar floor, and first-party dual-declaration is not a named precondition before user-visible diagnostics or migrated producers ship.

**Required change:** Seed release artifacts in the first implementation PR, make compatibility YAML the executable active-reader source, add external-support status rules and removal guards, name first-party dual-declaration ownership, and add a calendar/release-cadence floor.

### [Major] Pack validation and `[pack] requires_gc` are underspecified
**Sources:** 07-pack-versioning-ecosystem / Claude, Codex; 08-docs-dx-terminology / Claude, Codex

**Issue:** `gc formula validate --pack-source <url> --ref <ref> --all --json` lacks acquisition semantics for fetching, cache mutation, temp checkout lifetime, credentials, offline mode, prior import requirements, and failure diagnostics. `[pack] requires_gc` is load-bearing but lacks a byte-exact parser, active-binary comparison contract, and fail-closed resolver behavior.

**Required change:** Specify acquisition behavior and diagnostics, define the `[pack] requires_gc` grammar and source attribution, add fixtures for built-in/local/git/SHA-pinned/transitive/shadowed packs, and require pack-floor remediation when formula compiler requirements exceed host capability.

### [Major] Documentation gates still allow stale user guidance
**Sources:** 08-docs-dx-terminology / Claude, Codex; reinforced by 03-parser-validation-matrix, 07-pack-versioning-ecosystem

**Issue:** Docs can still teach `version` as normal formula metadata, omit tutorials from stale-guidance scans, recommend `GC_NATIVE_FORMULA=false` as production rollback guidance, or present remediation snippets that are invalid TOML. External authors also lack clear dual-declaration versus requires-only guidance during the alias window.

**Required change:** Broaden stale-guidance scanning to tutorials or all docs, reject `version` examples/prose outside legacy context, rewrite `docs/reference/formula.md` with content-level acceptance criteria, supersede stale migration guidance, and include side-by-side `[requires] formula_compiler` plus `[pack] requires_gc` examples.

### [Minor] Future capability metadata needs a deterministic encoding contract
**Sources:** 10-future-capability-architect / Claude, Codex

**Issue:** Omitted requirements, empty `[requires]`, and explicit `formula_compiler = ">=1"` are intended to be behaviorally equivalent, but exact persisted root/artifact metadata and provenance differences are not pinned. Future-axis manifests also lack identifier grammar, ordering, separator, duplicate, and old-reader behavior.

**Required change:** Add a canonical metadata table for omitted, empty, `>=1`, and `>=2`; define `gc.formula_requirement_axes`, `gc.formula_min_reader_capability`, and `gc.formula_unsupported_axes`; add unknown-axis and future-capability-root fixtures across CLI/API/order/convergence/fanout.

## Disagreements
- Several personas had severity disagreements rather than directional disagreements. Operator diagnostics, external pack migration, and convergence projection each had one reviewer returning `block` or `request-changes` while the other returned `approve-with-risks`; my assessment follows the stricter verdict because each issue defines a public or durable contract that implementers will encode.
- The `requires` naming overlap was seen as acceptable by some reviewers and risky by others. I do not treat the field names themselves as blockers, but the design must add side-by-side examples and glossary wording so formula compiler requirements and pack compatibility floors cannot be confused.
- Reviewers differed on whether temporary compatibility shims are acceptable for convergence. The safer design default is deletion plus guards; any shim must be one-call-deep, expiry-bound, non-authoritative, and backed by a real caller/test.
- Claude put more weight on release cadence, warning reach, and external maintainer lead time; Codex put more weight on stable JSON, parser, and acquisition contracts. These are complementary and should be resolved in the same release-gate artifacts.
- Gemini was absent for all persona syntheses. The current run was configured with `skip_gemini=true`, and the persona contracts treat Gemini as optional when skipped.

## Missing Evidence
- No attempt-50 persona synthesis files were present under `.gc/design-reviews/ga-j195cx/attempt-50/persona-syntheses`; the closed persona beads stamped their output paths under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses`, and those stamped outputs were used as the source material.
- No disabled-host parity fixture spans CLI, API, dashboard/generated types, and events in one byte-normalized row.
- No external-pack migration example shows duplicate bindings, binding identity, lockfile key, manual hint application, pack-floor bump, republish, and consumer lockfile update.
- No compiler-owned convergence projection type or accepted-artifact field is specified for convergence fields.
- No executable release artifacts demonstrate active old-reader selection, external-support expiration criteria, unknown external pack policy, or alias-removal approval.
- No per-surface Phase 3 rollback table names rollback handles, durable-state compatibility, cleanup actions, and zero-write tests.
- No complete raw-consumer/durable-writer allowlist, ratchet, or static-guard fixture set proves all direct `gc.formula_contract`, `contract`, `[requires]`, graph-workflow, and legacy subset parser consumers are blocked outside owned packages.
- No docs stale-guidance proof covers tutorials, formula `version` prose/examples, or production rollback guidance for `GC_NATIVE_FORMULA=false`.

## Recommended Changes
1. Fix the three blocking public/durable contracts first: diagnostic projection/exit semantics, external pack migration JSON with binding identity, and compiler-owned convergence projection plus zero-write API shape.
2. Add an active-root recovery section covering host downgrade, missing source, accepted-artifact reuse, artifact re-stamping, abandon/repair paths, and cross-binary continuation.
3. Convert rollout prose into executable release artifacts: compatibility YAML as active-reader source, external-support status gates, first-party dual-declaration checkpoint, calendar floor, and per-PR legacy-contract reports.
4. Add static and CI guards for durable-writer APIs, raw formula metadata consumers, convergence legacy-parser imports, workflow-control metadata writers, and allowlist shrinkage.
5. Complete the parser, raw-shape, construct, caller-preflight, projection-parity, future-axis, and pack-floor fixture matrices with count locks.
6. Define `gc formula validate --pack-source` acquisition semantics and `[pack] requires_gc` grammar, diagnostics, and resolver ordering.
7. Rewrite docs/reference and stale-guidance rules before any user-visible diagnostic ships, including `version` deprecation, `[requires]` examples, pack compatibility examples, and supersession of stale migration docs.
8. Pin default and future capability metadata semantics for roots/artifacts, including source attribution and old-reader behavior.
