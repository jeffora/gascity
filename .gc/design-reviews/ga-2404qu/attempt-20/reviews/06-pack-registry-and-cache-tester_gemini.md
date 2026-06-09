# Marcus Driscoll — DeepSeek V4 Flash Perspective Independent Review (Iteration 20 / Attempt 20)

**Verdict:** approve-with-risks

**Scope:** Builtin registry identity, synthetic cache pruning, system pack materialization, and provider-dependent pack continuity.

This review evaluates the Iteration 20 / Attempt 20 draft of the Core and Gastown Pack Split design (`design.md` / `design-before.md`) against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 20 design document represents a highly disciplined iteration that successfully integrates feedback from prior cycles. It establishes a robust, fail-closed model for system pack materialization, explicitly addresses the `IsSource` call-site behaviors, and cleans up the builtin registry's identity. 

From the perspective of **Pack Registry and Cache Testing**, the design is functionally complete and demonstrates a solid understanding of content-and-provenance validation. However, several subtle edge cases around offline cache-orphaning during upgrade, `SyntheticContentHash` invalidation side-effects, and provider script falling back are critical risks that need close attention before the registry and cache slice lands. 

Accordingly, I am recommending an **approve-with-risks** verdict, with specific actionable mitigations provided below.

---

## Top Strengths (Lane Verification)

1. **Explicit Registry Identity and Alias Retirement (§2937–2941):**
   The registry identity is now tightly pinned: `All()` returns only `core` (at `internal/packs/core`), `bd`, and `dolt`, with historical `gastown` and `maintenance` aliases explicitly retired. This addresses **Lane Q1** directly. The inclusion of negative tests for `IsSource`, `NameForSource`, and install-lock generation to reject the retired paths (§2968–2970) prevents silent reintroduction.

2. **Required-Pack Materialization Fail-Closed and Recovery (§3004–3010):**
   The shift of system-pack loading to `internal/systempacks` with fatal gates (integrity checks before resolution and `RequiredSystemPackParticipation` validation after resolution) is excellent. It ensures that behavior discovery never runs on a partially materialized or corrupt Core pack, satisfying **Lane Q2**.

3. **Content-and-Provenance Preserved (§2942–2946):**
   The design retains the strong existing validation models where integrity is driven by file manifests, content digests, and strict tamper tests, rather than path-count heuristics. This addresses the "tamper tests count paths" red flag.

---

## Critical Risks and Edge Cases

### [Major] Offline Operator Cache-Orphan Data Loss (Cache Namespace Switch)
The design specifies that the public Gastown cache key (`RepoCacheKey`) will transition from a namespaced synthetic key (`bundled-synthetic-v1\x00...`) to an ordinary remote-pack cache path keyed by repository source and the immutable `PublicGastownPackVersion` (§2945–2947). It also declares that stale synthetic cache entries for public Gastown or Maintenance are ignored/rejected (§2951–2953, §2994).
* **The Risk:** In an offline/air-gapped upgrade scenario, an existing city's lock file references public Gastown. On the new binary, the cache manager computes the new non-synthetic key. Since we are offline, the new binary cannot fetch the pack from the remote. Even though the correct Gastown files are already cached on disk (under the old namespaced synthetic directory), they will be ignored as stale, and the offline run will fail because it cannot access the network.
* **Mitigation:** The cache resolver should support a **one-time offline fallback/migration**. If lookup at the new remote-cache key fails due to being offline, and the old namespaced synthetic cache directory exists for that same source and commit, the cache manager should copy/promote those assets to the new cache path rather than immediately failing.

### [Major] `SyntheticContentHash` Invalidation Cascade
Removing `maintenance` and `gastown` from `All()` modifies the return value of `SyntheticContentHash()` (§2943–2944), which is global over all bundled packs.
* **The Risk:** This global hash mismatch will invalidate *all* synthetic caches, including those for `bd` and `dolt`, on the very first start of the new binary. While `MaterializeSyntheticRepo` is safe-by-construction and will self-heal by re-materializing them from the new embeds, this cascade creates a one-time startup I/O and CPU overhead.
* **Mitigation:** The design must explicitly document this one-time self-heal re-materialization as a known and accepted behavior, and verify it with a dedicated test proving that the `bd`/`dolt` caches successfully re-materialize offline without losing provider-specific custom data.

### [Minor] Position-Dependent Script Fallback in `dolt-target.sh`
When `dolt-target.sh` is moved from the retired Maintenance pack to Core, its directory depth changes.
* **The Risk:** `dolt-target.sh` has a relative fallback path (`$SCRIPT_DIR/../../../../../dolt/assets/scripts`) to locate the dolt provider's `port_resolve.sh`. Moving it to `internal/packs/core/...` breaks this relative depth. Additionally, `examples/dolt/port_resolve_test.go` contains a hardcoded reference to the old Maintenance script location.
* **Mitigation:** Ensure the moved script in Core resolves provider scripts purely via the environment variable `GC_SYSTEM_PACKS_DIR`, update the relative co-location fallback to match the new Core directory depth, and migrate the hardcoded path in `port_resolve_test.go`.

---

## Lane-Specific Questions & Answers

### 1. Do registry and embed tests assert that only the intended built-in packs remain, with Core sourced from the new path and no Gastown or Maintenance aliases?
* **Answer:** Yes, the design specifies updating `internal/builtinpacks/registry_test.go` to assert that the expected identities are exactly `core=internal/packs/core`, `bd=examples/bd`, and `dolt=examples/dolt` (§2962–2963), and adds negative tests to reject retired aliases (§2968–2970).

### 2. Does MaterializeBuiltinPacks repair missing or tampered Core while preserving provider-dependent bd and dolt behavior exactly as before?
* **Answer:** Yes, §3008–3010 specifies that required-pack materialization repairs missing/corrupt expected files and prunes/quarantines unexpected files before validation, ensuring that `bd` and `dolt` bytes and provenance are preserved.

### 3. Do synthetic cache tests reject modified manifests, unexpected files, and stale retired pack sources rather than checking file existence only?
* **Answer:** Yes, §2942 specifies keeping synthetic cache validation for the remaining bundled packs, and §2971–2974 requires stale-cache tests to verify rejection of old synthetic public Gastown caches.

---

## Actionable Requirements for downstream slices

1. **Offline Dual-Lookup Cache Fallback:** In `internal/config/cache.go`, implement a dual-read fallback for `RepoCacheKey` resolution when running offline, to copy/promote legacy synthetic Gastown caches to the new remote-cache path instead of failing.
2. **Self-Heal Test Integration:** Add an integration test in `registry_test.go` where an existing city with a stale global `SyntheticContentHash` is started offline, and verify that `bd` and `dolt` re-materialize successfully from the embeds.
3. **Resolve `dolt-target.sh` Path depth:** Update the relative path depth fallback inside Core's `dolt-target.sh` to match its new path in `internal/packs/core`, and rewire `examples/dolt/port_resolve_test.go` to use the Core path.
