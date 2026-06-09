# Sofia Khoury — Gemini 3.5 Flash (Independent Review, Iteration 8)

**Verdict:** pass

**Persona focus:** Doctor fix idempotency, legacy import rewrite safety, custom data preservation, operator-safe diagnostics, cross-file consistency, missed edge cases, pattern drift, and architectural coherence. This iteration re-evaluates the updated design document containing the **Attempt 7 Review Resolution Contracts** (lines 590-888) against Sofia Khoury's safety mandate and the active codebase.

---

## Executive Summary

The transition of my verdict from **block** in Attempt 7 to **pass** in Attempt 8 is grounded in the introduction of the explicit, highly robust **Preservation-Proven Doctor Transaction** contract (lines 716-748). This contract comprehensively mitigates the atomic rollback gap, solves the scoped TOML editor contradiction, establishes content-aware provenance, and secures loader bypasses across all runtime boundaries. The proposed architecture is exceptionally cohesive, operator-safe, and resilient against intermediate-state corruptions.

---

## Evaluation of Sofia's Critical Questions

### 1. Is the Core presence doctor fix a proven no-op on a healthy city, including repeated or concurrent runs with a controller active?

**Yes. The design completely satisfies this condition:**
*   **Idempotency & No-Op Nature:** In Step 1 and Step 7 of the composite coordinator (lines 729-730, 739-740), the doctor reads the current active configuration, verifies typed Core participation records, and re-validates the status quo. If a city is already healthy, the scoped TOML editor planner (lines 719-725) identifies zero editable spans, resulting in a clean, immediate, zero-mutation exit.
*   **Concurrent Active Controller Safety:** By incorporating the "directory swap/rename" pattern (lines 77-80, 835) and staging the Core materialization in a temporary directory (`.gc/system/packs/.tmp-core`, line 734) before atomically swapping it, an active controller reloading or loading config concurrently is completely shielded. The controller will either see the old state or the new state, but never a half-materialized Core pack.

### 2. When `gc doctor --fix` removes redundant Core or legacy Maintenance imports, what prevents it from deleting user-added imports or custom pack edits?

**The design establishes dual-layer protection:**
*   **Scoped Byte-Preserving Planner:** The design rejects whole-file TOML re-encoding (`toml.NewEncoder`) which historically dropped comments and custom metadata. Instead, the newly mandated planner (lines 719-725) identifies exact editable spans for the target `[[imports.*]]` tables and scalar lines, validates the resulting layout via the TOML parser, and performs surgical, byte-preserving edits. If a file cannot be edited without destroying comments, whitespace, or unknown tables, the auto-fix refuses and routes to a manual diagnostic.
*   **Content-Aware Provenance:** Suffix-based checking is retired. The design explicitly requires that the legacy directory is validated against a synthetic content hash of the pristine embedded pack (lines 105, 731). If the operator has introduced custom scripts, edited local templates, or added custom assets, the provenance check detects the drift, classifies it as a custom fork (line 762), bypasses the automatic fix, and outputs instructions for safe manual migration. Furthermore, legacy Maintenance state is *never* deleted (line 737) and is staged with reversible names.

### 3. If a local Gastown import is rewritten to a public remote, does the fix verify reachability and immutable provenance or fail with explicit operator guidance?

**Yes. The design introduces strict pre-flight reachability checks:**
*   **Preflight Validation:** Step 3 of the coordinator (lines 732-733) mandates that the doctor prove public Gastown reachability, exact commit installability, ordinary remote-cache identity, and lockfile generation *before* committing any TOML edits.
*   **Air-Gap & Offline Safety:** If the preflight reachability check fails (e.g., in air-gapped environments without a populated cache), the coordinator aborts. Instead of bare-failing or selecting an unsafe embedded fallback, the doctor prints a failure message detailing instructions on how to manually cache the pinned commit or run in a disconnected mode (lines 93-100, 745).

---

## Evaluation of Red Flags

*   **Red Flag 1: Fix mutates `city.toml` or `pack.toml` beyond its stated scope.**
    *   *Status:* **Resolved.** Handled by the scoped byte-preserving edit planner that leaves all other bytes/comments untouched (lines 719-725).
*   **Red Flag 2: Stale `.gc/system/packs` directories are deleted instead of ignored or reported.**
    *   *Status:* **Resolved.** The design explicitly mandates that legacy Maintenance directories are ignored, diagnosed, and *preserved, not deleted* (lines 55, 213, 737, 883, 1005).
*   **Red Flag 3: Doctor messaging treats unsupported custom Core or air-gapped remote migration as silently fixable.**
    *   *Status:* **Resolved.** Preflight provenance and reachability validations explicitly halt the auto-fix on drift or network absence, shifting cleanly to high-fidelity, operator-safe diagnostics (lines 731-733, 744).

---

## Additional Strengths & Architectural Coherence

1.  **State-Migration Rollback Atomicity (Lines 727-748):** Deferring runtime-state relocations and manifest writes until all validation steps pass, and then applying a `compare-before-rename` strategy on every target, guarantees that a mid-migration panic, crash, or permission failure leaves the city byte-identical and perfectly stable.
2.  **Extended Scanner Coverage (Lines 626-634):** Expanding required Core participation checks beyond `cmd/gc` to encompass `internal/api/` (API session managers) and `internal/session/` (session resolution) ensures that background state reloads, controller signals, and API-driven execution paths honor identical required system pack boundaries.
3.  **Strict Sequenced Dependency Gate (Lines 615-625, 734-738):** Verifying that Core is fully materialized and integrated before committing manifest edits prevents the catastrophic scenario where an operator ends up with a city that has had its legacy fallback removed but cannot load Core due to disk or permission constraints.

---

## Recommendations for the Implementation Phase

While the design is exceptionally solid and ready to proceed, I recommend the following minor implementation guidelines to prevent any runtime regressions:
1.  **Strict Whitespace-Insensitive Hashing for Provenance:** When calculating the synthetic content-hash of the pristine embedded pack to verify legacy directory provenance, ensure the hash is calculated over canonical content (e.g., stripping minor platform-specific line-ending differences like CRLF vs LF) to prevent false-positive "custom drift" detections on Windows or mixed environment checkouts.
2.  **Lockfile Key Order Preservation:** Ensure the scoped TOML editor is specifically tested against lockfiles containing duplicate table names or non-standard key orderings to guarantee that edit spans are matched with absolute lexical precision.
