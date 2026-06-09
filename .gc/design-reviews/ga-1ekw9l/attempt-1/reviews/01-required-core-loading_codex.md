# Camille Sato - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The live implementation plan now gives Core loading a concrete boundary: `internal/systempacks` owns materialization, validation, runtime includes, runtime loading, no-refresh loading, typed participation, and diagnostics.
- The descriptor-bound two-gate model is the right invariant. Gate 1 proves the materialized file set for Core and provider packs; Gate 2 proves the same descriptor participated in the resolved config, which avoids path-only or name-only false positives.
- The no-refresh reload contract answers the stale-behavior risk directly: invalid Core state can keep only read-only status/reporting from last-known-good config, while dispatch, formulas, orders, hooks, prompts, agent start, API mutation, and scheduling require `RequireReady`.

**Critical risks:**
- [Major] Provider-pack selection still has a boundary ambiguity. The plan says provider-conditioned required-pack selection lives only in `internal/systempacks`, but the API sketch includes `MaterializeRequiredPacks(ctx, cityPath, provider)` and says callers pass the selected beads provider. If each caller obtains that provider outside the boundary, `bd`/`dolt` required-pack participation can diverge from the same single trusted loading path the design is trying to enforce.
- [Major] The bypass scanner scope is still partly phrased as `behavior-driving internal/` packages. That is too judgment-based for this migration. The scanner should be deny-by-default over all non-test `cmd/gc` files and all production `internal/` files outside the loader implementation itself, with every lower-level config read either structurally owned by `internal/config` or present in the generated partial-read allowlist.
- [Minor] The degraded reload mode names the behavior but not the stable condition/event contract. The plan says an event and diagnostic are published, but the exact condition code, event payload, and API/CLI propagation proof should be tied to `migration-diagnostics.schema.json` so callers cannot invent local interpretations.

**Missing evidence:**
- The support artifact directory currently does not contain the required loader-adjacent evidence files named by the plan, including `pack-resolution-matrix.yaml`, `migration-diagnostics.schema.json`, and `acceptance-proof-matrix.yaml`. The plan correctly gates dependent slices on these, but they remain prerequisites rather than present evidence.
- The generated config-loader allowlist (`internal/systempacks/testdata/config_loader_allowlist.yaml` or final equivalent) does not exist yet, so scanner completeness is still an implementation promise.
- No public Gastown host-Core/no-Maintenance transcript or pin ledger is present yet. That is acceptable only because the rollout limits current decomposition to prerequisite-producing work.

**Required changes:**
- Specify how `internal/systempacks` discovers the beads provider used to select `bd` and `dolt`. Prefer a minimal typed pre-resolution read inside `internal/systempacks`; if any caller-supplied provider remains, restrict it to tests or explicitly allowlisted diagnostic paths and add mismatch tests.
- Rewrite the scanner requirement to be deny-by-default for all non-test `cmd/gc` and production `internal/` source, not only packages already identified as behavior-driving. Add positive controls for aliases, wrappers, function values, generated helpers, and package-local helper names.
- Add the degraded reload diagnostic contract to the plan: stable condition code, event type/payload, API response shape, CLI text/JSON behavior, and focused tests proving controller, API, and CLI surface the same condition.
- Keep the decomposition gate as written: only prerequisite-producing beads should be created until the support artifacts and public-pack transcripts exist and pass.

**Questions:**
- Is the beads provider selection source limited to city/root config fields that can be read by a minimal pre-resolution parser, or can imports/rig overrides affect it?
- In `read_only_degraded` mode, are API reads allowed to expose formula/order/prompt-derived state from last-known-good config, or should the mode be limited to health, diagnostics, and raw status only?
