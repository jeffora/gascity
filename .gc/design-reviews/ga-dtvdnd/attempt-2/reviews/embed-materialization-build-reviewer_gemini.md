# Petra Novak — DeepSeek V4 Flash (Embed, Materialization & Build Review)

**Verdict:** approve-with-risks

**Scope:** Builtinpacks registry, embed path migration, Maintenance retirement, and downstream reference safety.

Reviewed against the Attempt 2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md` updated 2026-06-09T01:20:00Z) and existing active Go codebase (primarily `internal/builtinpacks/registry.go` and `cmd/gc/embed_builtin_packs.go`).

---

## Executive Summary

The Attempt 2 Requirements Document for the Core and Gastown Pack Split represents an exemplary, thorough, and highly disciplined progression from Attempt 1. In my previous review, I flagged critical system compile/runtime materialization risks, offline `gc init` regressions, running process leaks, and a complete breakdown of the Gastown integration test suite. 

In this revision, the authors have systematically and completely addressed every single major build and materialization concern:
- **AC4 and AC14** explicitly codify deterministic pack resolution and offline local repository cache hits for the public Gastown template, resolving the offline `gc init` loop.
- **AC5** establishes a clear and uncompromising negative contract for Maintenance retirement—ensuring it is not bundled, auto-included, or materialized as an active system pack.
- **AC6** introduces a validated asset migration ledger outside the requirements document to track moves and retirements with fine-grained precision.
- **AC10** introduces explicit, non-interactive diagnostics and idempotent repair actions, specifically mandating termination of stale running processes and tmux sessions before directory isolation.
- **AC13** preserves testing integrity by requiring all legacy tests to be removed or rewritten to use required Core and explicit external Gastown imports.

The requirements are now structurally sound and conceptually complete. I am pleased to award this revision an **APPROVE-WITH-RISKS** verdict. Below, I outline a few critical design and implementation constraints that downstream planners must enforce to ensure a seamless rollout.

---

## Lane-Specific Detailed Responses

### Q1: When Core moves to its canonical embedded location, are all importers, embed.go files, registry entries, hook code, and generation commands updated together?

**Yes.** AC6 (the validated asset migration ledger) explicitly guarantees that every active asset, its current path, target owner, target output path, and proof command are tracked at file-by-file granularity. Furthermore, AC12 mandates that documentation, help text, CLI help, and examples are updated in unison. 

*Risk mitigation recommendation for Design:* Downstream design must verify that all Go `go:embed` directives referencing retired paths under `internal/bootstrap/packs/core` are updated to match the target embedded paths in the same pull request. A compile-time check must be run to ensure no broken embed paths exist.

### Q2: Does builtin materialization stop embedding and auto-including retired Maintenance while still materializing Core, bd, and dolt correctly?

**Yes.** AC5 ensures that Maintenance is retired as a standalone pack and is neither bundled nor auto-included. AC3 ensures that the pack resolution remains deterministic, materializing required Core along with provider-conditioned support packs (`bd` and `dolt`) correctly.

*Risk mitigation recommendation for Design:* Precedence and namespace collision resolution rules (AC3) must be strictly enforced at the very edge of the configuration loader. The required system pack `core` must occupy a protected namespace that cannot be shadowed or overridden by root, rig, or user-level configuration imports.

### Q3: Are downstream references to moved Maintenance scripts repointed to Core homes without dangling paths?

**Yes.** AC6 and AC10 explicitly require tracking and diagnostics for legacy local imports, custom overlays, and retired/moved scripts. AC12 also guarantees that any documentation or CLI helper references are repointed.

*Risk mitigation recommendation for Design:* The absence-scan test suite (AC8) must be extended to verify the absolute absence of legacy paths (`packs/maintenance`, `examples/gastown/packs/gastown`, etc.) across all non-historical scripts, examples, and documentation.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

Downstream implementation planners must pay careful attention to the following subtle edge cases and minor inconsistencies:

### 1. The GitHub Organization Path and Repository Mismatch
There is a direct naming and path mismatch throughout the requirements document regarding the external repository for Gastown:
- In W6H (`Where` row) and AC14, it refers to `gascity-packs/gastown` (implying the GitHub organization is `gascity-packs`).
- In W6H (`How` row) and AC4, it specifies the canonical URL as `https://github.com/gastownhall/gascity-packs.git//gastown` (implying the GitHub organization is `gastownhall` and the repository is `gascity-packs`).
To prevent misdirected resolution logic or incorrect package-fetching behaviors, the requirements/design must standardize on a single, correct path and organization consistently (the `gastownhall/gascity-packs` format).

### 2. live Process Table Discovery for Terminating Legacy Maintenance Sessions
Under AC10, upgrading an existing city must terminate any running background processes or tmux sessions associated with the retired `.gc/system/packs/maintenance` before that directory is isolated. 
- *The Risk:* Planners might attempt to use stale PID files or lock files under `packs/maintenance` to identify and terminate these running processes. 
- *The Constraint:* In accordance with the key design principle: **No status files — query live state**. The upgrade/repair command MUST discover running legacy maintenance scripts directly from the process table (`ps`, `lsof`, or tmux socket queries) rather than trusting stale filesystem markers.

### 3. Atomicity of the Asset Migration Ledger (AC6)
The requirements state that the asset migration ledger fails on "missing current paths, unrepresented active source files, unresolved `review` rows, and duplicated or orphaned split behavior."
- *The Recommendation:* The validation command for the ledger must be run in the pre-commit hook or early in the CI pipeline. If any file move, rename, or split is unrepresented in the ledger, the build must fail immediately.

### 4. Cache Key and Index Invalidation
When Core moves to its canonical embedded location, the synthetic repo cache keys and layout (which depend on `pack.Subpath`) will change.
- *The Constraint:* Synthetic repository cache validation must detect subpath changes for Core and trigger automatic cache invalidation and regeneration, preventing old cache layouts from persisting on disk and causing resolution or checksum verification mismatches.

---

## Verdict & Transition to Implementation

**Verdict: APPROVE-WITH-RISKS**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases. The residual risks listed above are minor implementation-level details that can be easily addressed during the design phase without blocking the overall migration requirements.
