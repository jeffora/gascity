# Simone Kaye — DeepSeek V4 Flash Perspective Independent Review (Iteration 7 / Attempt 7)

**Verdict:** approve-with-risks

**Scope:** External Gastown pack authority; registry behavior; source-tree cleanliness; and documentation consistency.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence.** I have re-grounded this independent review against the active Iteration 7 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-7/design-before.md`, 135 lines, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (835 lines, updated 2026-06-09). I evaluated the criteria and verified the external-pack dependency and caching hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/external-pack-docs-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-7/reviews/external-pack-docs-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 7 synthesis.
3. **Verdict Rationale.** The Iteration 7 requirements and proposed implementation plan represent an extremely mature, robust, and detailed architecture for decoupling required SDK Core assets from optional Gastown behavior. Decoupling the legacy directories, enforcing strict required-pack participation checks (AC3), and introducing a generated wording scanner (AC12) successfully establishes a clean, role-neutral SDK core. However, from the perspective of **External Gastown Pack Authority, Registry Behavior, and Source-tree Cleanliness**, a few critical caching races and dependency boundaries remain unaddressed in the proposed design. Therefore, I award an **APPROVE-WITH-RISKS** verdict and mandate three critical required changes to prevent runtime collisions, caching race conditions, and test-suite drift.

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
The implementation plan outlines a comprehensive, generated wording/docs scanner (Operator Docs And Generated References, lines 500-531) utilizing a terminology matrix to audit docs, CLI reference, generated help, and API schemas. This ensures complete consistency and prevents retired terminology (e.g. implicit Maintenance or in-tree examples paths) from slipping into the public-facing SDK surface.

---

## Critical Risks & Architectural Gaps

### 1. [Major] Absence of a Formal Local/Offline Mirror URL Override for Isolated/Air-Gapped Operators
* **The Risk:** Under AC16, offline and cache behavior for public Gastown is strictly fail-closed on cache misses and never falls back to stale local templates or synthetic aliases. While this prevents silent drift, it does not specify a way for air-gapped, isolated, or corporate network-restricted environments to configure a local mirror or Git redirect mapping for the public URL (`https://github.com/gastownhall/gascity-packs.git//gastown`). Without an explicit configuration option (e.g., `git config --global url...` integration or a `[registry.mirrors]` block in global `gc` configuration), operators in isolated environments are locked out of resolving the external pack unless they manually seed the cache folder, which is tedious and lacks automation.
* **The Fix:** Amend AC16 and the implementation plan to mandate that the pack resolver respects global/local Git configuration redirects (such as `url.<base>.insteadOf`) or a dedicated `gc` mirror redirection configuration, allowing operators to map public repository URLs to local/internal mirrors seamlessly.

### 2. [Major] Cross-Repository Wording and Terminology Drift (The "Documentation Bubble" Assumption)
* **The Risk:** AC12 establishes a rigorous terminology scanner to ensure Gas City's internal files, help text, and guides do not reference retired paths or treat in-tree examples as authoritative. However, this scanner is scoped purely to the Gas City repository. If the external public pack `gascity-packs/gastown`'s own documentation or internal prompts reference retired Gas City in-tree paths (e.g., `examples/gastown/packs/*` or `.gc/system/packs/maintenance`), the operator-facing experience remains fragmented and inconsistent.
* **The Gap:** The requirements assume the documentation boundary ends at the repository edge, whereas the user sees a single cohesive system across the SDK and the public pack.
* **The Fix:** Update AC12 and AC14 to require that the public Gastown pack's own documentation files (at the pinned commit) are audited by a companion terminology validation script before release-pin acceptance, preventing cross-repo documentation drift.

### 3. [Major] Cache Namespace Shadowing and Conflict Resolution for Same-Named Forked/Custom Packs
* **The Risk:** The implementation plan specifies that the `RepoCacheKey` is keyed by normalized source, commit, and subpath. However, there is no explicit collision detection or resolution if a user city imports a custom/forked public pack named `gastown` (e.g., `https://github.com/myorg/gascity-packs.git//gastown`) while a transitive dependency or template simultaneously imports the canonical public pack (`https://github.com/gastownhall/gascity-packs.git//gastown`).
* **The Gap:** AC3's resolution matrix ensures that custom local imports cannot shadow or satisfy required Core. However, it does not define deterministic collision resolution for same-named *optional* public packs from different Git remote sources.
* **The Fix:** Mandate that the pack resolution matrix (AC3) and loader explicitly fail-closed on same-named public packs from different remote sources unless they are explicitly namespaced or aliased in `pack.toml` imports, preventing accidental source shadowing or cache hijacking.

---

## Required Changes for Finalization

1. **Explicit Git Redirect & Mirror Support:** Update AC16 to mandate that public pack URL resolution honors standard Git redirect configurations (like `insteadOf`) or offers a configuration-level mirror mapping to support isolated/air-gapped operators.
2. **Mandatory Cross-Repo Documentation Auditing:** Amend AC12 and AC14 to require that the public Gastown pack's own documentation passes the terminology and path-retirement standard before release-pin acceptance.
3. **Prevent Optional Pack Shadowing & Collisions:** Update AC3 to specify that same-named optional packs from conflicting Git remote URLs must fail-closed unless namespaced or aliased, preventing namespace shadowing.

---

## Questions

* **Mirroring & Offline Pre-population:** Will the `gc doctor --fix --non-interactive` command provide an explicit flag or sub-command to pre-populate or export a validated public Gastown cache archive for air-gapped transport?
* **Third-Party Pack Compatibility:** If a third-party pack (not owned by `gascity-packs`) imports the retired `packs/maintenance`, does the loader emit an actionable upgrade diagnostic naming the retired import before failing closed?
* **Cache Eviction Policy:** How does the runtime handle cache pruning for old public-pack commits? Will there be an automatic eviction threshold, or is cache pruning strictly manual?
