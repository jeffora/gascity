# Marcus Driscoll — DeepSeek V4 Flash Perspective Independent Review (Iteration 19 / Attempt 19)

**Verdict:** approve-with-risks

**Lane:** Builtin registry identity, synthetic cache pruning, system pack materialization, and provider-dependent pack continuity.

Reviewed against the Iteration 19 draft of `design.md` (`updated_at: 2026-06-09T08:40:42Z`) and grounded in the live codebase in `internal/builtinpacks/`, `cmd/gc/embed_builtin_packs.go`, and the `bd` / `dolt` provider packs.

---

## Executive Summary

The Iteration 19 design represents a highly sophisticated and mature evolution of the Core/Gastown pack split. It successfully introduces explicit generated metadata contracts (`provider-exceptions.generated.yaml`, `behavior-manifest.generated.yaml`, `loader-inventory.generated.yaml`, etc.) and formal witness tables to ensure provider asset continuity. This effectively reconciles active role-cleaning with byte-level continuity requirements.

From the Registry, Cache, and Materializer perspectives, however, several residual risks and bootstrap paradoxes remain. Specifically, the retention of `SyntheticContentHash` (§2943) without explicit stale cache eviction rules, potential concurrency collisions during parallel startup self-healing (`MaterializeRequiredPacks` in §2989), remote loader crashes when executing git commands on `.git`-less legacy directories, and the optional nature of offline cache promotion (§2954) pose substantial hazards. These must be resolved prior to final approval.

---

## Top Strengths & Design Evolution

1. **Explicit Generated Metadata Contracts & Witness Tables:**
   The addition of explicit generated schemas—namely `provider-exceptions.generated.yaml`, `behavior-manifest.generated.yaml`, and `loader-inventory.generated.yaml`—replaces ad-hoc string comparisons with a statically verifiable metadata-driven pipeline. This ensures strict tracking of provider-dependent assets across the migration boundary.
2. **Resolver-Produced Required-Pack Provenance (§2989):**
   The table-driven provenance model provides clear, predictable mapping from specific loading/doctor failures to explicit, atomic recovery states, preventing recovery loops and safeguarding system-pack integrity.
3. **Adopted Prune-and-Quarantine Model:**
   The design successfully integrates the recommendation to prune or quarantine unexpected effective files *before* Gate 2 validation, eliminating startup deadlocks caused by stray or injected files in `.gc/system/packs/core/`.
4. **Strict Preflight Recipient Validation (§1378–1381):**
   Mandating immediate preflight failure when required recipients are missing guarantees that alert-routing and message-routing boundaries remain airtight.

---

## Critical Risks & Architectural Gaps

### 1. [Blocker] Legacy 5-Pack Combined-Hash Cache Stale Silent-Serving Risk (§2943–2953)
* **The Risk:** The design retains `SyntheticContentHash` (§2943) for verifying synthetic cache validity. However, legacy cities initialized before the migration contain a combined 5-pack synthetic cache.
* **The Hazard:** If the layout or contents of the cached packs change during the split (such as removing Gastown and Maintenance), the old combined-hash cache may be treated as a valid hit rather than being recognized as stale. This would lead the system to silently serve obsolete, combined-pack files, causing unpredictable runtime failures or silent role pollution.
* **Required Change:** Mandate an explicit, concrete test and corresponding loader/materializer logic to guarantee that any legacy 5-pack combined-hash synthetic cache is treated as stale, evicted, and completely re-materialized on the first run of the migrated binary.

### 2. [Major] Parallel Startup Self-Healing and Quarantine Write Collisions (§2989)
* **The Risk:** In highly concurrent environments (e.g., parallel hook executions, telemetry runs, or concurrent integration tests), multiple `gc` commands may start simultaneously on a city with missing or tampered Core assets.
* **The Hazard:** Each concurrent process will attempt to execute `MaterializeRequiredPacks` (§2989) and prune/quarantine files. This creates severe race conditions (file-not-found, half-written files, or permissions errors) resulting in spurious startup failures.
* **Required Change:** Require `MaterializeRequiredPacks` to acquire a lightweight, fast, advisory directory write-lock (e.g., via a lockfile in `.gc/system/packs/` with a short 500ms timeout) before mutating, pruning, or quarantining files in `.gc/system/packs/core/`.

### 3. [Major] Remote Loader Crash on `.git`-less Legacy Directories (§2951–2953)
* **The Risk:** Standard remote loader functions will route retired public Gastown and Maintenance aliases through normal `git` commands (like `git status` or `git rev-parse`) since they are no longer marked as `IsSource`.
* **The Hazard:** Legacy synthetic cache directories on disk do not contain a `.git` sub-directory. Running git commands inside these `.git`-less folders will crash the CLI with raw, unhandled shell errors.
* **Required Change:** Mandate that standard remote loaders explicitly check for the presence of a `.git` sub-directory or legacy synthetic manifests in target cache directories. If missing or detected, they must safely prune, ignore, or delete the directory as stale rather than running git commands.

### 4. [Major] Manual/Optional Cache Promotion Offline Upgrade Failure (§2954–2958)
* **The Risk:** Upgrading air-gapped or offline cities requires promoting legacy synthetic cache keys to ordinary cache keys. The design treats this promotion as optional or requiring manual activation (§2954–2958).
* **The Hazard:** If left optional, offline upgraded cities will immediately fail closed on their first command run because they cannot resolve pinned public Gastown commits over the network.
* **Required Change:** Automate the cache promotion migration step at the loader/materializer boundary. Under offline constraints or network resolution failure, the loader must automatically copy and re-key a valid, un-tampered legacy cache on disk to the ordinary remote-cache directory.

---

## Evaluation of the Three Key Questions

### 1. Do registry and embed tests assert that only the intended built-in packs remain, with Core sourced from the new path and no Gastown or Maintenance aliases?
**Yes.** The design successfully restricts the registry to return exactly `{core=internal/packs/core, bd=examples/bd, dolt=examples/dolt}`. Retiring `publicSubpathForPack` aliases and using explicit negative tests in `registry_test.go` to assert rejection of retired Maintenance/Gastown aliases directly satisfies this requirement.

### 2. Does `MaterializeBuiltinPacks` (now `MaterializeRequiredPacks`) repair missing or tampered Core while preserving provider-dependent `bd` and `dolt` behavior exactly as before?
**Satisfactory with Risks.** The table-driven resolver-produced provenance model and prune-and-quarantine repair phase are excellent. However, as noted in Critical Risk #2, directory-level advisory write-locking is necessary to prevent concurrent startup collisions.

### 3. Do synthetic cache tests reject modified manifests, unexpected files, and stale retired pack sources rather than checking file existence only?
**Yes.** The cryptographic digest validation and file-set integrity checks over manifests explicitly reject altered assets, satisfying the requirement to validate content and provenance instead of performing simple path-existence checks.

---

## Evaluation of Red Flags

### 1. Old `internal/bootstrap/packs/core` or retired aliases are still accepted silently
* **Status: Resolved.** The registry actively rejects retired aliases, and tests assert loud failures on legacy paths.

### 2. `bd` or `dolt` pack behavior changes while repairing Core
* **Status: High Risk.** Without the recommended directory-level advisory locking during self-heal, parallel repairs risk mutating shared/provider assets or crashing due to write collisions.

### 3. Tamper tests count paths instead of validating content and provenance
* **Status: Resolved.** Cryptographic digest and file-set validation are formally mandated in the design.

---

## Actionable Requirements & Proposed Adjustments

1. **Legacy combined-hash cache eviction:** Mandate that any existing cache containing legacy 5-pack combined-hash synthetic caches is treated as stale and pruned on first run.
2. **Runtime self-healing lock:** Implement a 500ms timeout advisory lock during `MaterializeRequiredPacks` mutations to secure parallel startup repair paths.
3. **Guard git validation on legacy caches:** Add preflight checks for `.git` folders in cached directories before executing remote git commands.
4. **Automate offline cache promotion:** Force automatic promotion of un-tampered legacy synthetic caches under offline/network-failure conditions to avoid breaking upgrades.

---

## Open Questions

1. **How are custom/operator modifications to embedded files classified?** If an operator manually edits an embedded pack file, does the `MaterializeRequiredPacks` classifier always quarantine it as an unexpected file, or is there a way to preserve minor local edits safely?
2. **What is the cleanup latency for quarantined files?** Do quarantined assets in the recovery location persist indefinitely, or is there an automated garbage-collection or prune step implemented during doctor fixes?
