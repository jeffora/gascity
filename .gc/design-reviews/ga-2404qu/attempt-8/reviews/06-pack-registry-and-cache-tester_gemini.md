# Marcus Lindqvist — Gemini (Independent Review, Iteration 8)

**Persona:** Pack Registry and Cache Tester
**Mandate:** Builtin registry identity, synthetic cache pruning, system pack materialization, provider-dependent pack continuity
**Verdict:** approve

---

## Lane Context & Alignment

This independent review evaluates the revised migration design document (updated 2026-06-07T06:45:00Z, located at [design-before.md](file:///data/projects/gascity/.gc/design-reviews/ga-2404qu/attempt-8/design-before.md)) against the `gascity` Go codebase, specifically the built-in pack registry (`internal/builtinpacks/registry.go`), the synthetic cache/pruning system (`internal/packman/cache.go`), and provider-dependent pack matrices.

The transition from a **block** in Iteration 7 to an **approve** in Iteration 8 is fully justified. The updated design document has undergone a complete paradigm shift: abstract, un-gated intentions have been replaced by concrete, machine-readable contracts, explicit active-pack filtering, and an exceptionally rigorous **"Retired-Source Classifier and Runtime Containment"** contract. Every critical, high-risk gap in my lane has been resolved.

---

## Top Strengths of the Revised Design

1. **Stale Prompt/Template Zombie Prevention (Registry/Cache Containment):**
   The introduction of the explicit instruction that **"Prompt and template discovery must walk only active resolved packs and required host packs. It must not glob preserved stale system-pack directories"** (lines 767–768) is a major victory. This prevents old templates (e.g., `mayor`, `deacon`, `polecat`) inside preserved stale `.gc/system/packs/` subdirectories from being loaded as active prompt baselines on startup, securing the decoupling boundary completely.
2. **Actionable Retired-Source Diagnostics (No Silent Git Fallbacks):**
   Rather than silently reclassifying retired in-tree sources (e.g., `internal/bootstrap/packs/core` or `examples/gastown/packs/maintenance`) as ordinary Git remotes and triggering opaque network errors, the design mandates a central **"Retired-Source Classifier"** (lines 749–766). The classifier converts retired sources into typed states, yielding precise, actionable migration diagnostics telling operators exactly how to update their manifests and lockfiles.
3. **Strict Core Validation (Full File-Set Integrity):**
   The required Core pre-resolution integrity check (lines 617–621) now enforces a strict manifest/file-set policy where **unexpected/injected files are fatal** unless explicitly classified as non-influential. This solves the "expected-files-only" vulnerability of `packContainsEmbeddedManifest` and ensures required system packs cannot be silently tampered with.
4. **Decoupled Bundled Cache & Offline Provider-Pack Continuity:**
   The design successfully acknowledges that the global `SyntheticContentHash` digest changes when Maintenance/Gastown are retired. It mitigates this invalidation by requiring offline old-cache-to-new-binary tests to prove `bd` and `dolt` caches self-heal and re-materialize byte-identically without needing network access (lines 838–840, 1133–1136).
5. **Database Enumerator Safeguard Relocation:**
   Moving the `TestBuiltinDatabaseEnumeratorsSkipManagedProbeDatabase` witness along with the Core copies of `jsonl-export.sh` and `reaper.sh` (lines 1179–1182) ensures that Dolt's managed-probe-DB protection needles are actively preserved. The generic Core reaper will not accidentally sweep provider probe databases.

---

## Critical Risks

*None.* The updated design successfully designs out all major and minor risks identified in previous iterations.

---

## Verification & Implementation Requirements

To ensure the implementation matches the high quality of this design, the following gates must be strictly enforced during development:

1. **Exact-Equality Assertions:**
   The rewritten `TestAllAndSourceAreDeterministic` and `TestBuiltinPackIncludes_*` tests must keep their exact-equality/exact-count assertions (lines 1149–1154). They must not be weakened to generic membership checks, ensuring that any accidental re-introduction of retired aliases triggers an immediate build failure.
2. **Negative Source Tests:**
   The negative test suite must explicitly assert that `IsSource` and `NameForSource` return false and trigger the retired diagnostic for:
   - `internal/bootstrap/packs/core`
   - `examples/gastown/packs/maintenance`
   - `examples/gastown/packs/gastown`
3. **Stale Directory Preservation Verification:**
   Tests for `MaterializeBuiltinPacks` must actively verify that if a user has custom files in `.gc/system/packs/maintenance`, running the materialization or `gc doctor --fix` does *not* delete or prune those files, safeguarding potential operator custom configurations.

---

## Questions for the Implementation Phase

1. **Retired-Source Classifier Placement:**
   Will the new retired-source classifier be implemented as a dedicated package or nested inside `internal/builtinpacks/`? Keeping it as a self-contained helper will make it easier to import across the loader and doctor packages.
2. **Unexpected File Classifier Policy:**
   For the pre-resolution integrity check, what is the default behavior if an unexpected file is found in Core and the central classifier does not recognize it? It should fail closed by default to prevent silent security bypasses.
