# Camille Sato

**Verdict:** approve-with-risks

**Top strengths:**
- The plan now names `internal/systempacks` as the single production boundary for required host packs and gives it concrete loader APIs, descriptor construction, materialization, validation, runtime include assembly, and refresh/no-refresh modes.
- The required-pack proof is no longer path-only. The descriptor-id flow requires both pre-resolution file-set validation and post-resolution `RequiredSystemPackParticipation`, which directly addresses copied-Core, same-name import, and stale materialization risks.
- The fail-closed behavior is explicit: only `ready` snapshots may drive dispatch, formula expansion, order evaluation, hooks, prompts, session start, worker creation, API mutation, or controller scheduling; degraded snapshots are read-only and must surface diagnostics.

**Critical risks:**
- [Major] Provider selection before strict Core validation is still underdefined. The API list includes `RequiredPackNames(cityPath string)` and later says callers pass city path and selected beads provider, but provider selection currently depends on city/config reads. If callers use direct `config.Load*` or ad hoc TOML reads to discover `bd` versus `dolt` before `internal/systempacks`, they can recreate the bypass the plan is trying to remove. The plan needs a typed bootstrap-read surface owned by `internal/systempacks`, with a narrow field contract for provider selection and source attribution.
- [Major] The bypass scanner contract is strong but still leaves "behavior-driving internal packages" as a judgment call. The current tree has direct config loads in areas that may become bootstrap diagnostics, report-only doctor paths, import commands, or runtime behavior. The plan should require the first systempacks slice to generate a complete initial bypass inventory from `cmd/gc` and `internal/`, classify every direct load as routed, partial-read, test-only, or illegal, and fail if new packages are omitted from scanner roots.
- [Major] `RuntimeGuard.RequireReady(op)` is described as an entry-point obligation, but the enforcement mechanism is still partly conventional. API/controller/CLI callers "must check" the mode and may not reinterpret diagnostics locally. That is correct, but the implementation plan should say whether behavior-changing APIs receive only a ready-checked handle/capability, or whether every call site manually calls `RequireReady`. Manual checks across dispatch, orders, hooks, prompts, sessions, workers, API mutations, and controller scheduling are easy to miss.
- [Minor] Last-known-good degraded mode needs a clearer ownership model. The plan says the controller keeps the last-known-good runtime config for read-only status/reporting when Core is invalid, while CLI and API paths surface the same diagnostic. It should name whether that snapshot is controller-memory only, persisted, or recomputed, and how stale prompt/formula/order content is prevented from leaking into read-only reporting beyond diagnostics.

**Missing evidence:**
- A concrete list of current direct loader sites and their expected disposition, especially `cmd/gc/init_provider_readiness.go`, `cmd/gc/cmd_import.go`, `cmd/gc/cmd_start_drift.go`, `internal/doctor/checks.go`, and command context helpers.
- The exact bootstrap-read contract for provider selection before required pack materialization, including allowed fields and why it cannot drive behavior.
- A call-site or type-level proof that behavior-changing operations cannot access config, includes, prompts, formulas, hooks, or worker/session start paths without a ready `RuntimeSnapshot` or equivalent capability.
- A degraded-mode test that corrupts Core after a successful load, then proves status/reporting still works while dispatch, formula expansion, order evaluation, hook rendering, prompt resolution, session start, worker creation, API mutation, and controller scheduling all fail closed with the same diagnostic.

**Required changes:**
- Add a `systempacks`-owned bootstrap provider-selection/read API, or explicitly fold provider selection into `LoadRuntimeCity`, so no caller needs a pre-systempacks `config.Load*` to choose required provider packs.
- Require the first loader slice to generate and validate a complete config-loader bypass inventory covering all production `cmd/gc` and `internal/` direct or wrapped config loads, with scanner roots that fail closed when new production packages are added outside classification.
- Strengthen the runtime guard design so behavior-changing code receives an already-ready capability or typed handle where practical, and reserve manual `RequireReady(op)` checks for audited edges with tests.
- Add degraded-mode tests that verify same-diagnostic read-only reporting and fail-closed behavior for every listed behavior-changing operation after Core is missing, corrupt, stale, shadowed, or missing participation.

**Questions:**
- What exact fields may be read before required Core validation to choose provider-conditioned packs, and which package owns that read?
- Should `RuntimeGuard` be a central capability passed into behavior-changing services, or a per-call check at each CLI/API/controller entry point?
