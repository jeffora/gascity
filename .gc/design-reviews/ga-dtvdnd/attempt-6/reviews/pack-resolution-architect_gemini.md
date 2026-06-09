# Priya Menon — DeepSeek V4 Flash Perspective Independent Review (Iteration 6 / Attempt 6)

**Verdict:** approve-with-risks

**Scope:** Required Core loading, pack registry behavior, import resolution mechanics, legacy import retirement, and multi-pack resolution precedence.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-6/design-before.md`, status updated 2026-06-09), the `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the live `examples/gastown/packs/maintenance` assets. I did not inherit prior conclusions; all findings are verified fresh against the repository tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automatic execution tools to write and read from `attempt-1/reviews/pack-resolution-architect_gemini.md` (which can block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/pack-resolution-architect_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-6/reviews/pack-resolution-architect_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 6 synthesis.
3. **Verdict Rationale.** The Attempt 6 requirements and implementation plan represent a remarkably mature, cohesive, and robust architecture for securing Gas City's dependency boundaries. By isolating Core as the required runtime foundation, retiring Maintenance as an active standalone pack, and defining fail-closed, bootstrap-safe loader diagnostics, the system successfully eliminates untraceable legacy fallbacks. However, from the strict perspective of **Pack Resolution Architecture**, a few subtle edge cases and systemic assumptions must be pinned down (such as the dev/test escape hatch safety boundary, staging clobbering in concurrent promotions, and post-unification compatibility validation in diamond dependencies). I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins to ensure complete runtime determinism.

---

## Evaluation of the Three Key Questions

### 1. How does Core become the canonical required runtime pack across init, doctor, CLI load, and runtime resolution?
**Architect Finding: Fully Resolved.**
The requirements and implementation plan successfully establish required-Core loading across all lifecycle phases:
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
**Architect Finding: Fully Resolved.**
Precedence order across required Core, provider-conditioned support packs (`bd`/`dolt`), and remote/local imports is explicitly specified. The implementation plan's "zero-duplicate-active and zero-merge gates" compare active bundled, active public, and cached sources, failing closed if the same behavior ID is active from more than one source or if Core and public Gastown both claim an unresolved behavior row. To make this resolution fully bulletproof, we must enforce strict design-level pins for diamond dependencies and concurrent staging.

---

## Critical Risks & Architectural Gaps (Architectural Findings)

### 1. [Major] Dev/Test Escape Hatch Leakage (The Production Safety Gap)
* **The Risk:** AC2 and the implementation-plan allowlists specify a "dev/test Core-less escape hatch" to allow tests to construct partial configs. If this escape hatch is triggerable via standard CLI flags or environment variables in a production binary, operators could run Core-less or partially-loaded cities, bypassing SDK self-sufficiency and violating core invariants.
* **The Gap:** Standard runtime flags (e.g., `--skip-core` or `--test-mode`) are prone to abuse or accidental activation in real deployments.
* **The Pin:** The dev/test Core-less escape hatch must be hard-coded to activate *only* within native Go test environments (e.g., verifying `flag.Lookup("test.v") != nil`), strictly preventing production/CLI activation via command-line flags or environment variables.

### 2. [Major] Staging Stale Cache Collision (The Temp-Directory Clobber Race)
* **The Risk:** AC16 and the implementation plan require cache writes and promotions to be atomic under concurrent commands, utilizing temp-directory staging.
* **The Gap:** If multiple concurrent `gc` commands try to promote the same or different packs concurrently and use a shared or fixed temporary folder name (such as `.gc/cache/tmp-core`), they will clobber each other's files, resulting in corrupted cache states before the atomic rename.
* **The Pin:** The temporary directories used for cache promotion staging must utilize process-isolated, randomized unique naming structures (such as `.gc/cache/tmp-<UUID>`) before performing the final atomic `os.Rename`.

### 3. [Major] Transitive Diamond Pin Resolution Policy (The Version-Lock Deadlock)
* **The Risk:** AC3 specifies deterministic resolution across root city imports and locked/cached remotes, but does not specify how operator-driven overrides are verified.
* **The Gap:** If the resolver simply unifies transitive pins under a root-level override, the resulting unified graph could be semantically incompatible with one of the dependent packs.
* **The Pin:** Root-level `city.toml` locks must override and unify all nested transitive pins, but a post-unification `packcompat` verification must still run on the resolved graph to ensure the overridden version satisfies the dependent packs' expected contract.

---

## Required Changes for Finalization

1. **Hard-Gated Escape Hatch:** Restrict the dev/test Core-less escape hatch solely to native Go test runners (`flag.Lookup("test.v") != nil`).
2. **Randomized Staging Cache Paths:** Enforce process-isolated or randomized unique temporary paths (e.g., `.gc/cache/tmp-<UUID>`) for all atomic promotion staging.
3. **Post-Unification Compatibility Checks:** Require a post-unification `packcompat` check to verify unified graphs under root-level locks.

---

## Open Questions for Implementation

* **How will `gc doctor` display conflicting transitive imports?** It should print the full import paths (e.g., `City -> Pack A -> Pack C@sha1` vs `City -> Pack B -> Pack C@sha2`) so operators can easily copy-paste the required pin into `city.toml`.
* **Should `.gc_provenance.json` use SHA-256 for local cache integrity validation?** Yes, SHA-256 is recommended for consistency with existing remote-pack digests.
