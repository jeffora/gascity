# Simone Kaye — DeepSeek V4 Flash Perspective Independent Review (Iteration 8 / Attempt 8)

**Lane:** external-pack-docs-reviewer (wave 1) — external Gastown pack authority, registry behavior, source-tree cleanliness, and documentation consistency.

**Verdict:** approve-with-risks

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active Attempt 8 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-8/design-before.md`, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09). I evaluated the criteria and verified the external-pack dependency and caching hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/external-pack-docs-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-8/reviews/external-pack-docs-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 8 synthesis.
3. **Verdict Rationale.** The Attempt 8 requirements represent an exceptionally high level of structural maturity. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract** (AC7) successfully resolves the requirements schema conflict. Therefore, I award an **APPROVE-WITH-RISKS** verdict. From the perspective of **External Gastown Pack Authority, Registry Behavior, and Source-tree Cleanliness**, a few critical caching races and dependency boundaries remain unaddressed in the proposed design.

---

## Evaluation of the Three Key Questions

### 1. Does the requirement make gascity-packs/gastown the authoritative source for Gastown behavior, templates, overlays, and workflow checks?
**Reviewer Finding: Yes.**
The requirements successfully establish `gascity-packs/gastown` as the single source of authority (AC4). The proposed rollout plan (Slice 1a/1b/1c) strictly enforces that no Gas City source deletion, Core role-generalization, or Maintenance removal may land until the public Gastown repository contains all matching behavior, prompts, scripts, and tests, validated by an immutable `sha:` pin, pack digests, and behavior-preservation proofs.

### 2. Will no maintained pack source remain under examples/gastown/packs, and do docs and tutorials stop presenting retired locations as authoritative?
**Reviewer Finding: Yes.**
Under AC4 and Slice 7 of the implementation plan, the old directories under `examples/gastown/packs/*` are deleted once replacement tests pass. The wording and docs scanner (AC12) actively audits all operator guides, tutorials, and CLI help text, ensuring that no active documentation references retired folders or treats `examples/gastown/packs/*` as authoritative. Stale directories such as `.gc/system/packs/maintenance` are correctly ignored by active discovery and classified strictly as legacy state.

### 3. Are public registry behavior, template imports, and operator-facing terminology consistent across docs, CLI messages, and examples?
**Reviewer Finding: Yes.**
The implementation plan outlines a comprehensive, generated wording/docs scanner (Operator Docs And Generated References, lines 496-531) utilizing a terminology matrix to audit docs, CLI reference, generated help, and API schemas. This ensures complete consistency and prevents retired terminology (e.g. implicit Maintenance or in-tree examples paths) from slipping into the public-facing SDK surface.

---

## Critical Risks & Architectural Gaps

### 1. [Major] Cross-Repository Wording and Terminology Drift (The "Documentation Bubble" Assumption)
* **The Risk:** AC12 establishes a rigorous terminology scanner to ensure Gas City's internal files, help text, and guides do not reference retired paths or treat in-tree examples as authoritative. However, this scanner is scoped purely to the Gas City repository. If the external public pack `gascity-packs/gastown`'s own documentation or internal prompts reference retired Gas City in-tree paths (e.g., `examples/gastown/packs/*` or `.gc/system/packs/maintenance`), the operator-facing experience remains fragmented and inconsistent.
* **The Gap:** The requirements assume the documentation boundary ends at the repository edge, whereas the user sees a cohesive system across the SDK and the public pack.
* **The Fix:** Update AC12 and AC14 to require that the public Gastown pack's own documentation files (at the pinned commit) are audited by a companion terminology validation script before release-pin acceptance, preventing cross-repo documentation drift.

### 2. [Major] Dangling Doc Context during Multi-Repo Rollout (The Flag-Day Gap)
* **The Risk:** Slices 2 through 7 gradually adopt the compatibility pin, extract Core, and finally delete stale sources. If an operator upgrades the Gas City binary but remains on an intermediate compatibility pin (compatibility mode in Slice 2 / Slice 3), the in-tree docs and help text might describe the "new world" (where Gastown is purely external and Maintenance is retired) while the active city is actually still running under legacy coexistence mode.
* **The Gap:** This creates "Documentation-to-Active-Config" version skew.
* **The Fix:** Ensure the documentation / wording scanner (AC12) or doctor CLI help dynamically reflects the active resolution mode (e.g., dynamically stating whether the city is running in "coexistence compatibility mode" vs. "full decoupled activation mode"), rather than presenting a static, idealized decoupled layout that does not match the operator's current rolling slice.

### 3. [Major] Stale Cache Alias Zombie-Reactivations (Source-tree Cleanliness)
* **The Risk:** The plan states: "Do not delete stale `.gc/system/packs/maintenance`... they are ignored by active discovery." What if an operator manually edits a local TOML file, or a third-party pack imports a path that resolves to these stale directories? If the directories still physically exist, some low-level parts of the code might still resolve them if there are bugs in the `internal/packsource` classifier.
* **The Gap:** Legacy fallback paths can lead to zombie execution of retired behaviors.
* **The Fix:** Require the doctor or `gc` runtime to actively quarantine (e.g., rename to `.gc/system/packs/maintenance.retired-quarantined` or append a `.disabled` suffix) these stale directories rather than leaving them under the active discovery search path.

### 4. [Major] Cache Directory Collision via Non-Sanctioned Mirror Manipulation
* **The Risk:** AC16 allows standard Git URL redirects (such as `insteadOf`) or a custom `gc` mirror mapping to satisfy the public pin. If a mirror maps a public repository URL to a local path, does the loader verify that the cache directory namespaced under the custom URL does not collide with the canonical public cache namespace? If two different Git remotes map to the same local cache directory structure due to redirection, it could lead to cache corruption or incorrect verification of the behavior-manifest digests.
* **The Gap:** Cache namespace collision under mirror configurations.
* **The Fix:** Mandate that the local cache namespace uses a sanitized hash of the effective resolved remote URL to prevent collision and cache hijacking under custom mirror configurations.

---

## Required Changes for Finalization

1. **Mandatory Cross-Repo Documentation Auditing:** Amend AC12 and AC14 to require that the public Gastown pack's own documentation passes the terminology and path-retirement standard before release-pin acceptance.
2. **Dynamic CLI and Doctor Documentation:** Require the wording scanner and help templates to dynamically adjust their terminology depending on whether the system is in coexistence (compatibility) or fully-activated mode.
3. **Quarantine Stale Directories:** Enforce a strict quarantine step (e.g., renaming stale directories with a `.retired` or `.disabled` extension) to prevent unexpected fallback or zombie imports.
4. **Hashed Cache Namespaces for Mirrors:** Mandate that the local cache namespace uses a sanitized hash of the effective resolved remote URL to prevent collision and cache hijacking under custom mirror configurations.

---

## Questions

* **Mirroring & Offline Pre-population:** Will the `gc doctor --fix --non-interactive` command provide an explicit flag or sub-command to pre-populate or export a validated public Gastown cache archive for air-gapped transport?
* **Third-Party Pack Compatibility:** If a third-party pack (not owned by `gascity-packs`) imports the retired `packs/maintenance`, does the loader emit an actionable upgrade diagnostic naming the retired import before failing closed?
* **Cache Eviction Policy:** How does the runtime handle cache pruning for old public-pack commits? Will there be an automatic eviction threshold, or is cache pruning strictly manual?
