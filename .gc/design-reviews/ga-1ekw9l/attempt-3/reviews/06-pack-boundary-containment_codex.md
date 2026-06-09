# Owen Gallagher - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan creates `internal/packsource` as the sole retired-source classification authority and routes load, install, cache, lockfile, materialization, discovery, doctor, docs-lint, generated-reference-lint, and public-source normalization through it (`design-before.md:222`-`design-before.md:228`).
- It explicitly removes Maintenance and Gastown from the embedded registry only after the activation gate, while stale generated directories remain on disk as ignored legacy state instead of active sources (`design-before.md:218`-`design-before.md:240`).
- The rollout keeps source deletion gated behind public Gastown proof, activation pin consumption, no-Maintenance production-loader packcompat, and duplicate-definition requirements (`design-before.md:481`-`design-before.md:489`).

**Critical risks:**
- [Major] The classifier adoption list is broad, but it does not explicitly name every behavior enumerator this migration must contain: prompt-template loading, formula discovery, order discovery, script resolution, and hook overlay enumeration. If any of those continue to use independent globs or direct path walks, stale `examples/gastown` or Maintenance directories can remain invisible to load/install tests while still entering live behavior.
- [Major] Compatibility-slice duplicate behavior needs an explicit active-source invariant. Slice 2 rewires examples to public imports while in-tree sources still exist, and Slice 5 stops if duplicate-definition requirements fail, but the plan should state the invariant directly: at any point, a given formula/order/prompt/script/hook id has exactly one active owner in resolved runtime behavior.
- [Minor] Retired custom/fork classification is named, but the operator outcome is not. A custom local fork that resembles retired Maintenance/Gastown should be reported and excluded from automatic rewrite; the plan should say whether it can remain as an explicit user import after classification or must be migrated manually.

**Missing evidence:**
- A behavior-enumerator inventory showing which loaders currently enumerate prompts, formulas, orders, scripts, hooks, overlays, docs/generated references, and pack manifests.
- Tests that stale in-tree Gastown/Maintenance directories remain on disk but do not appear in resolved prompts, formulas, orders, scripts, hooks, or rollback candidates.
- A duplicate-active-definition test for the compatibility window and rollback paths.
- Operator-facing diagnostics for retired custom/fork sources versus stale generated system packs.

**Required changes:**
- Extend the `internal/packsource` adoption contract to explicitly cover prompt, formula, order, script, hook, and overlay enumeration.
- Add an invariant that every active behavior id resolves from exactly one owner after classification, including compatibility and rollback states.
- Add tests that leave stale source directories and cache entries on disk and prove runtime discovery excludes them.
- Define how retired custom/fork sources are diagnosed and why automatic mutation is refused.

**Questions:**
- Which package owns behavior enumeration after this migration: `internal/systempacks`, `internal/packsource`, existing config loaders, or a combination with a strict call chain?
- During Slice 2, can a city import both public Gastown and the old in-tree Gastown path, and if so does that fail as duplicate active behavior or classify the old path as retired before load?
- Are retired custom/fork imports blocked at load time or allowed as explicit user-owned packs with warnings?
