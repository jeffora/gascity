# Camille Sato - DeepSeek V4 Flash

**Verdict:** approve-with-risks

> Lane: required Core/provider pack loading, typed participation provenance,
> deny-by-default loaders, bypass containment, fail-closed behavior. Reviewed
> against the current `implementation-plan.md` (529 lines,
> `updated_at: 2026-06-09T01:20:00Z`) — §"Required System Pack Loader"
> (171–213), §"Pack Registry, Cache, And Retired Source Authority" (215–241),
> §"Bootstrap Fixture Isolation" (304–322), and the
> `RequiredSystemPackParticipation` record in §"Data And State" (353–357).
> This independent review is produced using the DeepSeek V4 Flash persona,
> focusing specifically on first-principles trust boundaries, cross-document
> state consistency, and unstated runtime assumptions.

**Schema conformance:** Conforms to `gc.mayor.implementation-plan.v1`. Front
matter carries the required keys with `phase: implementation-plan` and no
`design_file`; the eight required top-level sections appear once each in the
required order (`Summary` → `Current System` → `Proposed Implementation` →
`Data And State` → `Testing` → `Rollout And Recovery` → `Open Questions`), and
`Open Questions` is `None`. No appended attempt/review prose in the artifact.

**Top strengths:**
- **Explicit Loader Boundary:** Consolidating the required pack loading lifecycle under `internal/systempacks` (171–180) is an excellent separation of concerns. This ensures that `cmd/gc` and internal behavior drivers do not independently compose include lists or invoke lower-level loaders, closing a major source of loader-path drift.
- **Strict Two-Gate Validation:** Separating pre-resolution file-set integrity (187–188) from post-resolution typed participation (189–192) introduces a robust, deny-by-default defense. If Core or provider packs are modified on disk, or if they fail to participate in the resolved config, the load fails immediately.
- **Fail-Closed Live Reload (LKG Protection):** Explicitly requiring the controller to retain the last-known-good (LKG) config *only* for read-only status and diagnostics (194–200) while blocking state-mutating dispatches, hooks, and agent starts is a major improvement over silent-continue or ungraceful crashes.

**Critical risks:**

- **[Major] Last-Known-Good (LKG) In-Memory Volatility and Startup Deadlock (Cross-document inconsistency).**
  The plan states that on reload failure, the controller "keeps the last-known-good runtime config only for read-only status/reporting" (194–196). However, the plan assumes LKG is held purely in-memory. If the controller process crashes or is restarted by the operator/orchestrator while the Core pack is in an invalid/corrupted state, the LKG configuration is lost.
  Upon restarting, the controller will run `LoadRuntimeCity` (182), which will fail the pre-resolution or post-resolution gate. Because there is no persisted LKG state, the controller will fail to boot entirely, meaning even "read-only status/reporting" and CLI/API diagnostics will be completely unavailable.
  *Real-world threat:* A minor corruption in `.gc/system/packs/core` during an automated reload will escalate from a graceful read-only state to a total system deadlock and API blackout upon any controller restart.
  → **Resolution:** The plan must specify whether LKG config state is serialized/persisted securely (e.g., as a protected bead in `.gc/beads.db` or a signed cached config) to survive restarts, or explicitly define the "bootstrapping failure mode" when no LKG is available in memory.

- **[Major] Accidental Shadowing and Core Erasure (Config resolution provenance loophole).**
  The post-resolution gate validates a typed `RequiredSystemPackParticipation` record (189–191) which includes "resolved config layer id" and "import edge proving participation in final config resolution" (355–356). However, in Gas City's multi-layer override model, a malicious or buggy user-supplied config pack can configure overlays, agents, or formulas that completely shadow or override every asset provided by Core.
  If a user-supplied config shadows all Core behaviors, the Core layer still technically "participates" in the config resolution chain (i.e., its import edge is present), but its actual behavioral contributions are nullified/erased.
  *Real-world threat:* Attackers or misconfigured third-party packs can bypass Core protections and hijack system behavior while still presenting a valid participation stamp, because "participation" is checked structurally rather than behaviorally.
  → **Resolution:** The participation gate must verify that Core assets are not shadowed/erased at critical extension points, or require that any override of a Core-defined system binding must carry an explicit security approval stamp in the resolved configuration.

- **[Major] Bootstrap Fixture Leakage and Allowlist Bloat via Low-Level Loaders.**
  To enforce bypass containment, line 202 mandates a scanner test to reject direct calls to `config.Load*` in production. However, line 183 explicitly permits low-level `internal/config` tests to use lower-level loaders.
  If the bypass scanner only targets `cmd/gc` and "behavior-driving `internal/` packages" (205) but exempts `*_test.go` files, there is a risk that test helper files (which are technically test code but may reside in packages imported by tests across the repository) leak direct-loader calls.
  Furthermore, if every test that needs a custom config mock must use the low-level loader, the partial-read allowlist (206–208) will become bloated with test-related exceptions, diluting its utility as a security-critical audit log.
  → **Resolution:** The bypass scanner must target all non-test code unconditionally. Test-only bypasses must be confined to packages named `*_test` or explicitly designated test-fixtures; they must never be permitted in non-test packages or shared helper utilities. The plan must clarify the exact scope of the scanner relative to test packages.

- **[Minor] Ambiguous Provider-Change Matrix on No-Refresh Reloads.**
  Line 179 introduces `LoadRuntimeCityNoRefresh`. If an operator modifies `city.toml` to change the active provider (e.g., from `dolt` to `bd`) or modifies provider-scoped env vars, but a "no-refresh" load is executed, how does the loader handle the mismatch?
  Since no-refresh reloads do not repair (194), if the provider was changed but the new provider's pack was never materialized, the pre-resolution gate (187–188) will fail. Does the system fall back to the LKG config (which used the *old* provider), or does the loader fail and lock down the controller? If it falls back to the LKG config, the active provider in memory will diverge from the provider declared on disk in `city.toml`, causing massive inconsistency in behavior-dispatch and status reporting.
  → **Resolution:** Define a concrete transition matrix for provider changes during no-refresh reloads. Specify that any change to the provider or provider-related configuration in `city.toml` invalidates the LKG cache and forces a full refreshed load.

**Missing evidence:**
- **Persisted LKG Config Schema/Location:** No detail on where the last-known-good config is stored to survive controller process restarts and crashes.
- **API/CLI Error Code Mapping under LKG:** The exact HTTP status codes and CLI exit codes returned by behavior-changing routes when the system is in read-only LKG mode (e.g., does a POST to `/dispatches` return a 503 with the loader diagnostic?).
- **Shadowing / Override Protection Rules:** Evidence showing how the participation validator prevents user packs from shadowing essential Core security overlays or system bindings.
- **Test-Fixture Sandboxing:** Verification of how `GC_BOOTSTRAP=skip` tests can run in network-isolated CI environments when they are barred from skipping systempacks materialization or provider validation (318–322).

**Required changes:**
- **LKG Persistence:** Explicitly require that the LKG config state be persisted to disk in a secure, read-only location (e.g., `.gc/system/lkg.toml` or as a sealed bead) so that the controller can boot into read-only diagnostic/reporting mode even after a hard crash or process restart.
- **Provider-Change Invalidation:** Add a rule stating that any change to the provider declaration in `city.toml` immediately invalidates the LKG config and forces a refreshed load, preventing provider-configuration mismatch.
- **API Guard Specifications:** Explicitly map the API behavior-changing entry points that must reject requests with a `503 Service Unavailable` or similar typed error containing the loader diagnostic when running on LKG config.
- **Scanner Target Isolation:** Specify that the bypass scanner runs on all packages except packages ending with `_test` or files within test directories; restrict the partial-read allowlist strictly to production-necessitated exceptions (like low-level config parsing tools), completely excluding test helpers.

**Questions:**
- If the controller starts up and the Core fileset is invalid, and no LKG is available in memory, what is the expected CLI/API output and exit behavior of the controller?
- How does the post-resolution participation validator ensure that a Core asset has not been shadowed or rendered inactive by an aggressive user-pack override?
- In an air-gapped test environment with `GC_BOOTSTRAP=skip`, how do provider-dependent pack validations succeed if the provider pack assets are not embedded or accessible?
