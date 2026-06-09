# Simone Kaye — DeepSeek V4 Flash Perspective Independent Review (Iteration 6 / Attempt 6)

**Verdict:** approve-with-risks

**Scope:** External Gastown pack authority; registry behavior; source-tree cleanliness; and documentation consistency.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence.** I have re-grounded this independent review against the active Attempt 6 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-6/design-before.md`, 119 lines, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (657 lines, updated 2026-06-09). I evaluated the criteria and verified the external-pack dependency and caching hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/external-pack-docs-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-6/reviews/external-pack-docs-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 6 synthesis.
3. **Verdict Rationale.** The Attempt 6 requirements and proposed implementation plan represent a highly mature and comprehensive approach to decoupling required SDK Core assets from Gastown behavior and retiring Maintenance. Decoupling the legacy directories, enforcing strict required-pack participation checks (AC3), and introducing a generated wording scanner (AC12) successfully establishes a clean, role-neutral SDK core. However, from the perspective of **External Gastown Pack Authority, Registry Behavior, and Source-tree Cleanliness**, a few critical caching races and dependency boundaries remain unaddressed in the proposed design. Therefore, I award an **APPROVE-WITH-RISKS** verdict and mandate three critical required changes to prevent runtime collisions, caching race conditions, and test-suite drift.

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
The implementation plan outlines a comprehensive, generated wording/docs scanner (Operator Docs And Generated References, lines 397-425) utilizing a terminology matrix to audit docs, CLI reference, generated help, and API schemas. This ensures complete consistency and prevents retired terminology (e.g. implicit Maintenance or in-tree examples paths) from slipping into the public-facing SDK surface.

---

## Critical Risks & Architectural Gaps

### 1. [Major] The Cache Promotion Race Condition under Concurrent Invocation
* **The Risk:** Under "Pack Registry, Cache, and Retired Source Authority," the design specifies that public Gastown is promoted from ordinary remote-pack caches based on source, commit, and subpath digests. However, it does not define concurrency safety during promotion. If multiple concurrent `gc` commands trigger promotion of the same downloaded public commit simultaneously, they may write to `.gc/cache` at the same time, resulting in file truncation, lockfile corruption, or corrupted pack directories.
* **The Fix:** Mandate that all cache promotion writes are guarded by a SHA-scoped, advisory file lock on the cache directory. Furthermore, the promotion must stage the unpacked directory under a unique temporary folder on the same filesystem and perform an atomic `os.Rename` to publish it, preventing concurrent writes or partial/corrupted cache states.

### 2. [Major] Relative Path Traversal and Implicit Traversal in Cross-Pack Imports
* **The Risk:** The public `gascity-packs/gastown` pack relies on Core behavior and templates. If the public Gastown formulas or prompts reference Core assets via relative filesystem paths (e.g. `../../../system/packs/core`) or assume a standard host layout, it violates the "External Pack Authority" boundary. The pack will fail to resolve under custom user-configured environments, custom rig directories, or isolated k8s/subprocess runners where Core is materialized in non-standard paths.
* **The Gap:** The plan notes that patches still assume the host runtime supplies implicit Core layers but does not specify a strict import syntax or resolution guarantee for the public pack to locate Core assets safely.
* **The Fix:** Mandate that public Gastown references Core assets strictly via canonical system-pack imports (`[imports.core]`) or a formal cross-pack metadata import syntax in `pack.toml`, and completely prohibit relative path traversal or implicit directories. This must be validated by the `test/packcompat` gate.

### 3. [Major] Local Mocking and Caching Bypass Drift in CI
* **The Risk:** One of the most common failures of public pack resolution is "works locally but fails on public pack resolution." To keep CI fast and offline, tests in `test/packcompat` often use mock functions or synthetic cache overrides that bypass the real Git resolution, lock generation, and digest verification pipelines. This allows caching and resolver bugs to remain undetected until they hit production networks.
* **The Fix:** Require `test/packcompat` and offline tests to verify caching behavior by configuring a strict loopback redirect mapping (e.g. environment-controlled mapping of `https://github.com/...` to a local bare Git repository). This forces the actual Git clone, cache caching, and checksum verification pipelines to execute identically offline, rather than using synthetic test-only mock bypasses.

---

## Required Changes for Finalization

1. **Atomic Cache Promotion & Advisory Locking:** Amend the cache promotion design to require SHA-scoped file locking and atomic `os.Rename` from a temp staging directory to ensure concurrency-safe cache updates.
2. **Forbid Relative Cross-Pack Path Traversal:** Update AC4 and the implementation plan to mandate that public Gastown references Core assets strictly through canonical registered imports or formal metadata-bound APIs, explicitly prohibiting filesystem-level relative paths.
3. **Mandatory Git Loopback for Caching Tests:** Amend AC16 and offline testing guidelines to require loopback redirects to local bare Git repositories rather than synthetic bypass mocks, ensuring the real Git resolution and caching pipelines are executed in CI.

---

## Questions

* **Digest Verification Timeout Policy:** What is the exact fallback behavior if a public pack's digest verification fails on a network timeout or DNS resolution error during normal operations—does it fail closed immediately, or can it utilize a stale but valid local cache after prompting?
- **Wording Scanner Archival Scope:** How will we handle documentation versioning on the main website? Will old guides remain readable in an archive directory, or will the wording scanner purge all mentions of Maintenance from past release documentation?
