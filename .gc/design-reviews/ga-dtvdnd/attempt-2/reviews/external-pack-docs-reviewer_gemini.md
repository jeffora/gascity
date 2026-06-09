# Simone Kaye — DeepSeek V4 Flash (External Pack & Docs Review)

**Verdict:** approve-with-risks

**Lane:** external Gastown pack authority, registry behavior, source-tree cleanliness, documentation consistency.

---

## Executive Summary

The Attempt 2 Requirements Document for the Core and Gastown Pack Split represents an exemplary, highly disciplined evolution from the Attempt 1 draft. In my prior review, I issued a **BLOCK** verdict due to a critical, existential testing paradox: the requirements mandated deleting in-tree packs (`examples/gastown/packs/{gastown,maintenance}`) without addressing the extensive Go test suite (`examples/gastown/gastown_test.go`, etc.) that directly imports and depends on these local paths. This would have instantly broken the repository's build/CI pipeline.

In Attempt 2, the authors have systematically resolved my primary blocker and significantly strengthened the entire framework:
1. **Testing Paradox Resolved (AC13):** The addition of AC13 explicitly mandates that all legacy tests asserting implicit Maintenance or in-tree Gastown behavior must either be removed with explicit replacement coverage or rewritten to prove required Core plus explicit external Gastown imports. This ensures that the testing boundary is refactored cleanly alongside the pack split.
2. **Explicit Authority Boundaries (AC4 & AC14):** AC4 explicitly binds fresh Gastown init to the external public pack `https://github.com/gastownhall/gascity-packs.git//gastown` with a pinned SHA, and AC14 implements a CI/release validation gate using the public checkout or a pinned cache to verify that local mock-ups do not mask a broken external pack.
3. **Robust Documentation Consistency (AC12):** AC12 enforces a strict, three-tier terminology contract (Core is required, Gastown is external-optional, Maintenance is retired) across all user-facing documentation, CLI help screens, examples, and diagnostics, verified by a comprehensive audit/grep scan.

Given these massive improvements, I am upgrading my verdict to **APPROVE-WITH-RISKS**. The requirements are now structurally and conceptually ready to transition to the design phase. Below, I outline my lane-specific findings and specify critical, high-fidelity edge cases that downstream design must enforce to preserve source-tree cleanliness and Nondeterministic Idempotence (NDI).

---

## Lane-Specific Detailed Responses

### Q1: Does the requirement make `gascity-packs/gastown` the authoritative source for Gastown behavior, templates, overlays, and workflow checks?

**Yes.** 
The requirements establish this authority cleanly and unambiguously through multiple layers of constraints:
- **Product Definition:** The Problem Statement establishes the desired product outcome: "Gastown behavior is loaded explicitly from the public `gascity-packs/gastown` pack."
- **W6H "How" and Example Mapping:** The happy-path mapping (Row 2) explicitly states that `gc init --template gastown` configures the root `pack.toml` and rig imports to declare `[imports.gastown]` from `https://github.com/gastownhall/gascity-packs.git//gastown` with the current pinned version (`sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` as of this update).
- **Validation Guardrails (AC14):** Crucially, AC14 mandates that public pack checkouts or pinned caches must be validated in CI/release gates to prove the roles, prompts, commands, formulas, orders, overlays, and checks needed by supported templates. This ensures the external repository is treated as the true, verified authoritative source, preventing a local copy from masking external brokenness.

*Design Constraint:* Downstream design must ensure that the `gc init` template generation engine fetches the pinned remote pack and validates its contents during compilation or release validation. The template loader must use standard registry resolution rather than side-channeling to local folders.

### Q2: Will no maintained pack source remain under `examples/gastown/packs`, and do docs and tutorials stop presenting retired locations as authoritative?

**Yes.**
- **Source Tree Cleanliness:** AC4 explicitly bans fresh initialization from relying on in-tree examples, `.gc/system/packs/gastown`, or implicit Maintenance. While AC5 retires Maintenance as a standalone active pack, AC6 establishes a **validated asset migration ledger** that explicitly tracks every file under the legacy pack roots and records its "target owner, target output path or retirement action." This ensures that any remaining directories are demoted to explicitly non-authoritative examples (or deleted entirely) in an audited, ledger-validated manner.
- **Documentation Cleansing (AC12):** AC12 explicitly forbids presenting `packs/maintenance`, `packs/gastown`, `examples/gastown/packs/*`, or `.gc/system/packs/*` as authoritative current sources in documentation, examples, CLI help, doctor messages, and import-state outputs, except where explicitly noted in historical migration guides. A grep/audit tool with allowed exceptions is mandated to verify this.

*Design Constraint:* Downstream design must implement the AC6 ledger validation as an automated test. If any unrepresented file under `examples/gastown/packs/` is detected, or if any active file retains a cross-import like `[imports.maintenance] source = "../maintenance"`, the build must fail.

### Q3: Are public registry behavior, template imports, and operator-facing terminology consistent across docs, CLI messages, and examples?

**Yes.**
AC12 mandates a consistent, three-way messaging model across all documentation, examples, CLI help, doctor messages, and import-state outputs:
1. **Core** is required.
2. **Gastown** is external-optional (unless the Gastown template is selected).
3. **Maintenance** is retired.

This ensures unified operator-facing terminology. Additionally, Example Mapping Edge Case 7 specifies that under offline/air-gapped conditions, if the pinned remote pack is cached locally in the repo cache, resolution succeeds and reports provenance; otherwise, it fails with an actionable missing-cache diagnostic and *never* falls back to in-tree example paths. This prevents inconsistent or silent fallback behaviors.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

To ensure a high-fidelity implementation that conforms to first-principles design, downstream designers must address the following critical edge cases and cross-document nuances:

### 1. Nondeterministic Idempotence (NDI) and Pinned Versioning
- **The Risk:** If `gc init --template gastown` or existing city imports resolve the public Gastown pack using a mutable reference (such as a branch name like `main` or a moving tag like `v1-latest`), the system violates Nondeterministic Idempotence (NDI). A remote pack commit could silently alter agent behaviors, prompts, or formulas in an active city, leading to non-reproducible runtime failures.
- **The Design Constraint:** Downstream design must mandate that all remote pack imports in the generated configuration use immutable version specifiers (specifically, git commit SHAs like `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` or immutable semantic version tags verified by a lockfile checksum). The configuration parser must fail if a mutable branch is specified without an explicit override.

### 2. Registry Cache Isolation and Offline Recovery
- **The Risk:** In-tree caches could become corrupted or shared across sandboxes, leading to cross-sandbox contamination. Conversely, an offline operator must be able to work seamlessly if the pinned pack is already cached.
- **The Design Constraint:** The pack resolution engine must use a secure, isolated local repository cache (e.g., `.gc/cache/packs/`). When running offline, the loader must verify the integrity of the cached pack (matching the pinned SHA) before loading it. If the cache is missing or corrupt, it must fail cleanly with a clear description of how the operator can manually pre-seed the cache (e.g., via `gc pack fetch`).

### 3. Graceful Version Skew Window (Open Question 3)
- **The Risk:** A Gas City SDK release may deprecate or modify internal template variables or event types, while the external public Gastown pack continues to use them, causing a crash or unrecognized event errors at runtime.
- **The Design Constraint:** The design must specify a formal, bidirectional compatibility protocol. The SDK and the public pack should declare supported API/template schema versions (e.g., in `pack.toml`). When resolving imports, the configuration loader must compare these versions and issue an explicit, actionable version-skew warning or block execution if the versions are incompatible.

### 4. Recursive Scans for Stale System Pack Cache Cleanup
- **The Risk:** Stale system pack caches or synthetic state (such as `.gc/system/packs/maintenance` or old tmux environment variables) from prior runs might survive an upgrade, leading to ghost imports.
- **The Design Constraint:** In accordance with the key design principle: **No status files — query live state**. Rather than relying on static markers, the startup sequence and `gc doctor` must perform a recursive scan of the city's active `.gc/` directory. Any stale, legacy system pack state must be programmatically ignored, pruned, or flagged for safe, operator-approved cleanup.

---

## Verdict & Transition to Design

**Verdict: APPROVE-WITH-RISKS**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases. The open questions (specifically Q1 regarding ledger ownership, Q2 regarding manifest storage, and Q3 regarding version skew) are excellent, high-fidelity boundary decisions that are properly framed to be resolved as part of the initial design draft.
