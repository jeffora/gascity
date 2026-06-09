# Priya Menon - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The requirements establish a single Core source authority: release-bundled Core from `internal/packs/core`, represented in resolved config with source, version or digest, materialized path, lock/cache provenance, and collision state.
- AC3 is appropriately broad for the pack-resolution problem. It covers required Core, provider-conditioned `bd`/`dolt` packs, explicit Gastown imports, root/city/rig imports, locks, caches, overlays, duplicate names, stale materialized copies, same-named assets, synthetic aliases, missing Core, and transitive diamond conflicts.
- The legacy-retirement contract is fail-closed rather than fallback-based: in-tree Gastown and Maintenance roots cannot silently satisfy behavior, while doctor/import-state keep bootstrap diagnostics and source attribution available.

**Critical risks:**
- [Minor] AC3 is dense enough that the implementation plan must avoid splitting it into inconsistent resolver, doctor, and CLI behaviors. The requirement already demands the same condition-code registry and source-attribution model across init, doctor, import-state, CLI config load, and runtime resolution; that must remain a single acceptance gate.
- [Minor] Provider-conditioned `bd`/`dolt` support-pack cardinality and same-named asset precedence are named, but they are high-risk edge cases. The future matrix must include executable rows for co-activation, mutual exclusion, local override pins, and pack-qualified asset selection.

**Missing evidence:**
- No product unknowns are left open. The missing items are the expected future artifacts: `pack-resolution-matrix.yaml`, condition-code registry tests, source-attribution goldens, stale-cache/system-pack fixtures, and diamond-conflict tests.

**Required changes:**
- None before requirements approval.

**Questions:**
- None.
