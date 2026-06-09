# Camille Sato - DeepSeek

**Verdict:** block

Lane: required Core and provider pack loading, typed participation provenance, deny-by-default loaders, bypass containment, fail-closed behavior. Reviewed against `plans/core-gastown-pack-migration/implementation-plan.md` (`updated_at` 2026-06-09T13:20:59Z; §Required System Pack Loader lines 221-311, §Pack Registry/Cache lines 313-356, §Data And State lines 541-545) plus the live tree.

---

## Top Strengths

- **Fail-closed live-reload and degraded-mode contract:** The design correctly answers the reload integrity problem. Setting `Mode ∈ {ready, read_only_degraded, blocked}` ensures only `ready` drives active dispatch, formulas, hooks, prompts, and session starts. Having no-refresh reloads refuse behavior-changing operations and surface identical diagnostics via API and CLI is a massive step forward.
- **Two bound fatal validation gates:** Separating pre-resolution file-set validation (Gate 1) from post-resolution participation validation (Gate 2) is conceptual excellence. This prevents simple path/name/digest coincidence attacks where a same-named user-provided or copied pack masquerades as the host system pack.
- **AST/Type-aware syntactic scanner:** Requiring a scanner modeled on `TestGCNonTestFilesStayOnWorkerBoundary` that explicitly flags aliases, selector values, wrappers, and package-local helpers addresses the bypass risks that a naive text grep would miss.

## Critical Risks

### [Blocker] Required system-pack participation is not yet an implementable resolver-owned provenance contract

The design relies on `RequiredSystemPackParticipation` containing "embedded source identity, resolved config layer identity, and an import edge proving participation in final config resolution."
However, there is a fundamental design-dependency and package-layering issue here:
1. **Import Cycle Risk:** If `internal/config` is responsible for producing participation records, but those records consume or reference types owned/defined by `internal/systempacks`, we will introduce an upward dependency from `internal/config` to `internal/systempacks` (or a package layering violation). Today, `internal/config` has zero upward dependencies.
2. **Missing Provenance API:** The current `config.Provenance` (`internal/config/compose.go:72-91`) carries no per-layer embedded source identity, pack digest, or content-manifest digest—only `Imports[name]=sourceFile`. An import edge derived from today's provenance is exactly the name+path coincidence the plan says must NOT prove participation.
The design never specifies how the `internal/config` provenance API must be updated to mint these trusted records, nor where the descriptor/participation types are housed.

**Required change:** Define descriptor and participation types in `internal/config` or a lower leaf package (with no upward dependency to `internal/systempacks`). Specify that `internal/systempacks` supplies trusted required-pack descriptors during config resolution, and the config resolver mints participation records.

### [Blocker] Bootstrap, repair, and partial-read paths are not concretely carved out from the fail-closed loader

By mandating that any loader call must fail closed on missing, corrupt, or partially materialized Core, the design risks introducing a "bricked-city" failure mode. If Core is corrupt or missing:
1. `gc init` cannot run to initialize a city.
2. `gc doctor --fix` cannot run to repair the corrupt pack or configuration because it would try to load config and fail closed.
We must have a concrete escape hatch or allowlisted partial-read inventory for bootstrap/repair, but the plan does not define this.

**Required change:** Provide an explicit bootstrap/repair protocol. Either define allowlisted partial reads for `gc init`, doctor diagnostics, and doctor fix flows, or a dedicated bootstrap mode whose allowed behavior is limited to initialization/materialization/repair.

### [Major] Gate 1 and Gate 2 validation need a precise proof source

The validation gates must prevent user-supplied or copied packs from satisfying required system pack loading. Gate 2 must validate the exact Gate 1 descriptor target: resolved import edge, resolved layer id, effective contribution, materialized directory, and digests must all match the same trusted descriptor, with no fallback to path, name, or digest coincidence. The design is too vague on how Gate 2 verifies this participation.

**Required change:** Explicitly require Gate 2 to reject any Core layer whose edge is `"(implicit)"` or path/name-only. The implementation must include test fixtures and cases for user or imported packs named `core`, copied Core trees, manually imported Core paths, materialized-but-not-imported Core, and provider pack participation for `bd` and `dolt`.

### [Major] Fail-closed runtime readiness is voluntary and too easy to bypass

Under the current plan, `LoadRuntimeCity` produces a live `Config` regardless of `Mode`. Caller entrypoints merely receive a `RuntimeSnapshot` or `RuntimeGuard` and are expected to remember to call `RequireReady(op)`. This is a convention, not a structural enforcement. Any caller that forgets to call `RequireReady` can use stale/unvalidated config after Core integrity has failed, silently bypassing the loud failure guarantee.

**Required change:** Make the readiness gate structural in the API shape. Either the snapshot must withhold a behavior-usable `Config` unless `ready` (e.g., behavior-changing accessors return `(Config, error)` that errors in `read_only_degraded`/`blocked`), or add a scanner/test proving every behavior-changing entrypoint calls `RequireReady` before config fields are used.

### [Major] Scanner scope "behavior-driving `internal/` packages" is undefined and bypassable

The proposed scanner scope "`cmd/gc` and behavior-driving `internal/` packages" is not a deny-by-default scope. This relies on an ad-hoc pre-filter. If a package is newly added or misclassified, its direct `config.Load*` calls will escape the scanner. Furthermore, the plan is silent on whether generated code (such as `internal/api/genclient` and generated server handlers) is in scope.

**Required change:** Define the scanner scope as deny-by-default over all production Go (`cmd/gc` + every non-test `internal/` file), with `config.Load*` failing CI unless an allowlist row proves a partial read. The allowlist must map each exception to an approved schema with file, function, call kind, fields consumed, reason, and a focused non-behavior test. Generated API/client code must be explicitly addressed.

### [Major] Live reload and strictness boundaries are underspecified

Required Core/provider file-set integrity and participation gates must be completely non-disableable by `--no-strict`. No-refresh reloads must perform no repair/write and keep the last-known-good configuration only for status reporting. Under degraded mode, the system must refuse behavior-changing operations with the same diagnostic ID and expose one stable diagnostic/event to prevent lock contention or inconsistent state.

**Required change:** State explicitly that required-pack integrity and participation gates cannot be disabled by `--no-strict`. Specify that `tryReloadConfig` replaces manual include assembly with `LoadRuntimeCityNoRefresh`.

### [Major] Materialization and locking boundaries are not pinned

Normal, routine read-only commands and status/reporting routes should validate required packs without taking exclusive repair locks or mutating the filesystem. Repair/self-healing must be explicit, bounded, non-interactive, and safe around operator-edited system-pack files. Otherwise, concurrent read-only queries could contend with reload locks or corrupt operator edits.

**Required change:** Specify locking and mutation rules. Normal materialization and no-refresh validation must not write or take exclusive repair locks, while explicit repair must be bounded, non-interactive, and safe around local operator edits.

### [Minor] Scanner target list references non-existent symbol

The scanner target list includes `config.LoadCity`, which does not exist in the codebase.

**Required change:** Replace `config.LoadCity` in the scanner target list with the real loader symbols, including `LoadWithIncludesOptions` if that path remains reachable.

---

## Missing Evidence

- **Provenance API design:** A concrete `internal/config` provenance or resolved-layer API that records trusted required-pack descriptor ID, embedded source identity, materialized directory, digests (`pack.toml`, manifest, file-set), resolved layer ID, import edge, and allowed-use mode.
- **Package ownership contract:** Defined package ownership for `RequiredDescriptor` and `RequiredSystemPackParticipation` that avoids package layering violations (no upward dependency from `internal/config` to `internal/systempacks`).
- **Loader bypass inventory:** A checked-in allowlist mapping current production `config.Load*`, `LoadWithIncludesOptions`, `loadCityConfig*`, etc. to allowlist rows with justification.
- **Bootstrap/repair rules:** Explicit diagnostic and repair behavior rules when required packs are missing or corrupt, including how `gc init` and `doctor` run.
- **Test coverage definition:** Test names and fixtures for `core` collisions (user/imported), copied Core trees, and required-provider changes during reload.
- **Readiness Snapshot/Accessor API:** A named ready/degraded snapshot API or accessor model to structurally enforce readiness check.
- **Event definitions:** The event type, registered payload, and diagnostic ID for reload failures.
- **Concurrency & locking rules:** Detailed locking and mutation rules for concurrent CLI invocations and unclassifiable system-pack files.

---

## Required Changes (Priority Order)

1. **Resolver Provenance API & Package Layering:** Define `RequiredDescriptor` and `RequiredSystemPackParticipation` in `internal/config` or a lower leaf package to prevent upward dependency cycles. Have `internal/systempacks` supply descriptors and `internal/config` mint the participation records.
2. **Validation Gate Alignment:** Force Gate 2 to validate the exact Gate 1 descriptor target. Reject a Core layer whose edge is implicit or path/name-only.
3. **Bootstrap and Repair Escape Hatches:** Add an explicit bootstrap/repair protocol (either allowlisted partial reads for `gc init` and `doctor`, or a dedicated bootstrap mode).
4. **Deny-by-default Scanner Scope:** Expand scanner scope to all production Go. Require a generated allowlist with fields consumed, reason, and focused non-behavior tests. Correct the scanner target list (remove `config.LoadCity` and add `LoadWithIncludesOptions`).
5. **Structural Readiness Guard:** Structuralize `RequireReady(op)` by withholding a usable config in `read_only_degraded` or `blocked` modes, or verifying coverage via AST scanners.
6. **Strictness reload rules:** Assert Core integrity gates are not disableable by `--no-strict`.
7. **Locking & Materialization Boundaries:** Explicitly bound materialization/locking so read-only validation performs no writes and holds no exclusive locks.

---

## Questions

1. Is `LoadRuntimeCityNoRefresh` used for live controller reloads to prevent silent Core regeneration from masking operator-edited files?
2. Does the live reload failure publish an event registered under `events.KnownEventTypes` before controller/API callers rely on it?
3. How are concurrent `gc` invocations protected against materialization/integrity check races? Does read-only validation avoid taking exclusive repair locks?
4. In `blocked` mode, what API state reads are permitted, and how are we guaranteed they aren't served from unvalidated or corrupt configuration data?
