# Priya Menon — DeepSeek V4 Flash Perspective Independent Review (Iteration 5 / Attempt 5)

**Verdict:** approve-with-risks

**Scope:** Required Core loading, pack registry behavior, import resolution mechanics, legacy import retirement, and multi-pack resolution precedence.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-5/design-before.md`, 135 lines, status updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` dog assets this migration retires, and the public `gascity-packs/gastown` pack source. I did not inherit prior conclusions; all findings are re-verified against the tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect documented in `attempt-4/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/pack-resolution-architect_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/pack-resolution-architect_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 5 synthesis.
3. **Verdict Rationale.** The Attempt 5 requirements and the corresponding implementation plan represent a major milestone in securing Gas City's pack loading and dependency boundaries. By isolating Core as the required runtime foundation, retiring Maintenance as a standalone active pack, and defining clear bootstrap-only loader diagnostics, the system successfully eliminates untraceable legacy fallbacks. However, from the strict perspective of **Pack Resolution Architecture**, several critical edge cases and systemic gaps remain unaddressed (such as the static asset collision gap, cache corruption vulnerability, and transitive diamond lock conflicts). I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins to ensure robust, deterministic pack resolution.

---

## Evaluation of the Three Key Questions

### 1. How does Core become the canonical required runtime pack across init, doctor, CLI load, and runtime resolution?
**Architect Finding: Fully Resolved.**
The Requirements and Implementation Plan successfully establish required-Core loading across all lifecycle phases:
- **Initialization & CLI Load:** `gc init` configures Core as a required system-pack identity in resolved config for real cities, with trusted provenance that user imports cannot shadow.
- **Loader Encapsulation:** The implementation introduces `internal/systempacks` as the single production boundary for required host packs, exposing `LoadRuntimeCity` and `LoadRuntimeCityNoRefresh`. Callers do not hand-build includes.
- **Verification Gates:** Runtime loading enforces two fatal gates—pre-resolution validation of Core/provider file-sets, and post-resolution validation of typed `RequiredSystemPackParticipation` records. Bypasses are prevented via scanner tests in CI that reject direct calls to `config.Load*` in `cmd/gc` and behavior-driving packages.
- **Diagnostics:** Under AC11, `gc doctor` and `gc import-state` can run in bootstrap-only mode when pack resolution is broken, reporting missing Core with exact source layer attribution.

### 2. Are implicit Maintenance and in-tree examples/gastown imports explicitly retired while gc init --template gastown imports gascity-packs/gastown?
**Architect Finding: Fully Resolved.**
The migration cleanly retires legacy dependencies:
- **Maintenance Retirement:** AC5 enforces that Maintenance is no longer bundled, public-source recognized, auto-included, materialized, or selected. `internal/builtinpacks/registry.go` is pruned.
- **Legacy Path Authority:** `internal/packsource` is established as the sole classifier for retired paths. Legacy directories on disk are safely ignored during startup and active discovery.
- **Explicit Gastown Init:** AC4 mandates that `gc init --template gastown` configures the public `gascity-packs/gastown` pack explicitly via `pack.toml` with an immutable `sha:` pin, with in-tree and implicit fallbacks strictly forbidden.

### 3. What is the deterministic resolution order when Core, Gastown, bd, and dolt all participate?
**Architect Finding: Partially Resolved.**
The requirements specify that pack resolution is deterministic across Core, provider-conditioned packs (`bd`/`dolt`), and remote/local imports. The Implementation Plan introduces "zero-duplicate-active and zero-merge gates" to compare active bundled, active public, and cached sources, failing closed if the same behavior ID is active from more than one source. 
However, while *behavior-level* resolution is strictly gated, *file-level* and *asset-level* precedence (overlays, custom scripts, templates) remains undefined, creating potential shadowing ambiguities (see Critical Risk 1).

---

## Critical Risks & Architectural Gaps (Architectural Findings)

### 1. [Major] The Overlay Resolution & Shadowing Ambiguity (The Static Asset Collision Gap)
* **The Risk:** AC3 and the Implementation Plan specify that runtime loading fails closed if Core and public Gastown both claim an unresolved "behavior row" (beads or formulas). However, they do not define a resolution policy for non-bead files, static prompt assets, custom helper scripts, and custom hook overlays. 
* **The Gap:** If both Core and the explicitly imported public Gastown pack contain a file with the exact same path (e.g., `overlay/git_hook.sh` or a prompt fragment template), the resolution behavior is undefined. Without a clear precedence pin, the runtime will either shadow non-deterministically, crash, or silently load stale Core default assets when Gastown-specific customizations are expected.
* **The Pin:** We must explicitly define asset-level precedence in AC3. Explicitly imported packs (like Gastown) must take precedence and override required Core assets for static assets, templates, and overlays, whereas Core assets serve as default fallbacks only when no imported pack overrides them.

### 2. [Major] Corrupt Cache Staging and Partial Cache-Hit Loading (The Cache Mutation Race)
* **The Risk:** AC16 and the Implementation Plan require cache writes and promotions to be atomic under concurrent commands, utilizing temp-directory staging and atomic rename operations.
* **The Gap:** If a process is abruptly terminated (via `SIGKILL`, system crash, or power loss) in the split second *after* a promotion creates the cache directory but *before* all files are fully written, subsequent runs will hit the cache because the folder exists. The resolver will then silently load a corrupted, partial pack, breaking runtime stability.
* **The Pin:** The resolver must never assume a cached directory is healthy merely because the directory exists on disk. Every cached pack must require a complete integrity/completion marker (such as `.gc_provenance.json` containing file hashes and a completion timestamp) written as the final atomic step of promotion. If this marker is missing, incomplete, or fails checksum verification, the resolver must treat the entry as a cache-miss, delete the corrupted folder, and fail-closed.

### 3. [Major] Transitive Diamond Pin Resolution Policy (The Version-Lock Deadlock)
* **The Risk:** AC3 specifies deterministic resolution across root city imports and locked/cached remotes, but does not specify a reconciliation policy for conflicting transitive version locks (e.g., City imports Pack A and Pack B, where Pack A transitively imports Pack C at `commit-1` and Pack B transitively imports Pack C at `commit-2`).
* **The Gap:** If the resolver simply fails on any mismatch in the transitive graph, operators will face version-lock deadlocks that are impossible to resolve without manually modifying third-party packs.
* **The Pin:** Introduce a strict root-override rule: any version pin or lock specified in the root city's `city.toml` (or `pack.toml` lockfile) must override and unify all nested transitive pins. If no root-level lock exists and transitive pins conflict, the resolver must fail-closed with a clear diagnostic indicating the conflicting import paths.

---

## Required Changes for Finalization

1. **Asset Precedence Pin:** Update AC3 to specify that explicitly imported packs take precedence over required Core for static assets, templates, and overlays.
2. **Cache Integrity Verification:** Amend AC16 to require a complete `.gc_provenance.json` completion marker for all cached packs, causing the resolver to fail-closed on incomplete or corrupted cache hits.
3. **Transitive Lock Overrides:** Specify in AC3 that root-level `city.toml` locks override and unify all nested transitive pins to prevent diamond dependency deadlocks.

---

## Open Questions for Implementation

* **How will `gc doctor` present conflicting transitive imports?** It should explicitly output the full dependency paths (e.g., `City -> Pack A -> Pack C@sha1` vs `City -> Pack B -> Pack C@sha2`) so the operator can easily copy-paste the correct pin into `city.toml`.
* **Should the `.gc_provenance.json` marker use cryptographically secure hashes (SHA-256) or fast CRC32 checksums for local cache validation?** SHA-256 is recommended for consistency with existing remote-pack digests.
