# Petra Novak — DeepSeek V4 Flash (Embed, Materialization & Build Review)

**Verdict:** block

**Scope:** Builtinpacks registry, embed path migration, Maintenance retirement, and downstream reference safety.

Reviewed against the Attempt 4/3/2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-4/design-before.md` updated 2026-06-09T01:20:00Z, which is byte-identical to Attempt 2) and the live Go codebase.

---

## Executive Summary

As the **Embed, Materialization & Build Reviewer**, I must issue a **BLOCK** verdict on the current Requirements Document for the Core and Gastown Pack Split. 

While the requirements document has structured the high-level goals well—such as retiring Maintenance and loading Gastown explicitly—it remains completely identical between Attempts 2, 3, and 4. This means multiple critical, low-level materialization, build, and compiler-safety blockers raised in previous rounds remain completely unaddressed.

Specifically, the requirements document fails to define:
1. **Core's canonical post-migration embedded path** or how existing reference paths under `internal/bootstrap/packs/core` are handled (blocking compile-time predictability).
2. **The Go-level embed and import removals** required to retire Maintenance without breaking the compiler.
3. **A downstream "reference closure" contract** (AC6 ledger only tracks asset *sources*, not asset *consumers*), exposing us to catastrophic runtime breaks in scripts, doctor checks, and examples.

Until these gaps are resolved and concrete acceptance criteria are added to handle both compiler and reference closure, this document cannot safely transition to the implementation phase.

---

## Lane-Specific Detailed Responses

### Q1: When Core moves to its canonical embedded location, are all importers, embed.go files, registry entries, hook code, and generation commands updated together?

**No.** The current requirements document is completely silent on Core's target embedded home and how compatibility/deprecation of the old path is handled.
- **The Live State:** `internal/builtinpacks/registry.go:20` compile-imports `github.com/gastownhall/gascity/internal/bootstrap/packs/core`. Line 53 hardcodes the subpath `{Name: "core", Subpath: "internal/bootstrap/packs/core", FS: core.PackFS}`.
- **The Risk:** There are **12 live test/readme/doc sites** hard-pinned to this path:
  - `registry_test.go:23,52-58,197,214,281`
  - `bundled_import_test.go:44,68`
  - `remotesource_test.go:16,18`
  - `prompt_test.go:781-782`
  - `hooks/config/README.md:18,59`
- **The Blocker:** Because string-path test fixtures, READMEs, and markdown docs do not trigger compilation failures, moving Core without an explicit, sweep-enforcing Acceptance Criterion (under AC13) guarantees that stale old-path references will rot silently in the codebase. We must decide if the old `internal/bootstrap/packs/core` path is rejected, legacy-aliased, or warned against, and require a comprehensive stale-reference sweep.

### Q2: Does builtin materialization stop embedding and auto-including retired Maintenance while still materializing Core, bd, and dolt correctly?

**No.** While AC5 says Maintenance is retired as a standalone pack, the requirements do not name or constrain the Go-level code removals needed to enforce this.
- **The Live State:**
  - `cmd/gc/embed_builtin_packs.go:237` hardcodes `required := []string{"core", "maintenance"}`.
  - `cmd/gc/embed_builtin_packs.go:265` comments: `"Core and maintenance are always included."`
  - `internal/builtinpacks/registry.go:19` compile-imports `"github.com/gastownhall/gascity/examples/gastown/packs/maintenance"`.
  - `internal/builtinpacks/registry.go:56` lists `maintenance` in the `All()` slice.
  - `internal/builtinpacks/registry.go:128-129` hardcodes `case "gastown", "maintenance"` to return their public subpaths in `publicSubpathForPack`.
- **The Risk:** Retiring Maintenance is a *Go compile-time risk*. If we delete the Maintenance pack files but fail to remove the compile-time import in `registry.go:19`, the build breaks instantly. If we keep the import but remove the pack, it fails.
- **The Blocker:** Builtin materialization is self-enforcing at compile time, but leaving the explicit Go-level steps un-gated by ACs allows half-done removals that leak references. We need a clear negative contract in AC5 mapping the removal of:
  1. The compile-import in `registry.go:19`
  2. The bundled `All()` slice entry in `registry.go:56`
  3. The `publicSubpathForPack` case alias in `registry.go:128`
  4. The hardcoded required list in `embed_builtin_packs.go:237` and `embed_builtin_packs.go:265`.

### Q3: Are downstream references to moved Maintenance scripts repointed to Core homes without dangling paths?

**No.** This is the most severe unaddressed runtime risk. AC6 mandates an asset migration ledger to track *source* files, but does not enforce **downstream consumer/reference closure**.
- **The Live State:** There is a broad surface of active scripts, configs, and doctor checks pointing directly at legacy Maintenance assets:
  - **Runtime break:** `examples/dolt/assets/scripts/port_resolve.sh:6` comment indicates coupling with `.gc/system/packs/maintenance/assets/scripts/dolt-target.sh`. Sourcing is inverted in `examples/gastown/packs/maintenance/assets/scripts/dolt-target.sh:160` where it evaluates `. "${DOLT_PORT_RESOLVE_SCRIPT:?port_resolve.sh not resolved}"` against the dolt pack's materialized `port_resolve.sh`. If Maintenance is retired and the script is rehomed or retired, but the sourcing line is not repointed, Dolt port resolution silently fails at runtime.
  - **Silent feature loss:** `examples/gastown/packs/gastown/assets/scripts/status-line.sh:14-16` sources `_bd_trace.sh` from `packs/maintenance`. It uses a `-f` guard, so when Maintenance is retired, the script continues but silently drops the tracing capability.
  - **Doctor failures:** `cmd/gc/jsonl_archive_doctor_check.go:59-60` explicitly searches `.gc/runtime/packs/maintenance/jsonl-export-state.json`. 
- **The Blocker:** An asset migration ledger that only covers source-to-target mapping misses these consumers. We must extend AC6 (or add a separate AC) requiring **downstream reference closure**: every consumer point (`source` lines, config paths, doctor searches) must be repointed or retired, validated by build and runtime test executions (e.g. running dolt port-resolution post-migration).

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

### 1. In-Flight Session State vs. Stale Materialized State
Under AC10, the migration must handle stale `.gc/system/packs/maintenance` or old synthetic cache directories. 
- *The Inconsistency:* While W6H and AC10 talk about "diagnosing and repairing" existing cities, they do not resolve Open Question 4: "For existing cities with in-flight sessions using prompts or formulas from retired paths, should the migration allow those sessions to finish, require immediate restart, or expose an operator decision?"
- *The Risk:* If a city is upgraded while an agent session is active, and the materializer prunes/moves Maintenance, the active session will instantly crash or behave unpredictably when it next attempts to load a prompt or script from `.gc/system/packs/maintenance`. We must decide this transition policy before implementation.

### 2. Live Process Table Discovery for Clean Up
AC10 demands the termination of running background processes and tmux sessions associated with retired Maintenance directories before isolation.
- *The Constraint:* Developers must not use stale lock files or PID files under `packs/maintenance` to find processes to kill. Per the design principle: **No status files — query live state**. The diagnostics/repair tool must scan the live process table (`ps`, `lsof`, or tmux socket queries) to identify and kill running Maintenance executors.

### 3. Bidirectional Ledger Validation
AC6 fails on "missing current paths" and "unrepresented active source files".
- *The Gap:* The validation contract is unidirectional. It fails if an active source file is unrepresented in the ledger, but it does **not** fail if the ledger contains a "phantom row" pointing to a path that does not exist in the snapshot. With 107 files moving, any typo in the ledger's `current path` will result in a silent failure to migrate that file. Validation must be **bidirectional** (using `git ls-files` under the roots as the canonical file set).

---

## Missing Evidence

1. **Stated Core canonical path** (e.g., whether `internal/packs/core` or `internal/bootstrap/packs/core` is the target).
2. **Go-level embed/import removal contracts** explicitly added to AC5.
3. **Reference-closure acceptance criterion** mapping consumer script and doctor search repointing.
4. **Offline first-time `gc init --template gastown` behavior** when the public cache is completely empty and no internet connection is present.

---

## Required Changes

1. **Add a Downstream Reference Closure Criterion (AC15):** Explicitly require that all active consumer code, examples, doctor check paths (`jsonl_archive_doctor_check.go`), and test assertions referencing moving/retired paths are repointed, with runtime-test verification (e.g. dolt port-resolution).
2. **Resolve Core's Target Path & Deprecation Policy:** Define Core's canonical post-migration location, and update AC13 to require a full stale-reference sweep covering all 12 Core-pinned test/doc files.
3. **Detail Go-level removal contracts in AC5:** Explicitly list the deletion of `registry.go` compile-imports, `All()` entries, `publicSubpathForPack` cases, and `embed_builtin_packs.go` required lists.
4. **Enforce Bidirectional Ledger Validation in AC6:** The ledger validation tool must fail if any row's `current path` does not resolve to an active file in the named `git ls-files` baseline snapshot.

---

## Questions

- Post-split, will the Go compiler enforce that `core` does not contain any Gastown role names, or are we relying solely on the absence-scan test suite (AC8)?
- For scripts like `port_resolve.sh` and `status-line.sh`, which repo/pack will house the moved helper scripts (`dolt-target.sh`, `_bd_trace.sh`)? Will they go to Core, Gastown, or get fully inlined?
- Does first-time offline `gc init --template gastown` fail-closed with a clear diagnostic, or do we provide a binary-embedded bootstrap bundle for Gastown?
