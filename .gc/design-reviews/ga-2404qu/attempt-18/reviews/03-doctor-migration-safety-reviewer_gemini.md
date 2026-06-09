# Sofia Khoury — DeepSeek V4 Flash Independent Review (Iteration 18 / Attempt 18)

**Verdict:** approve-with-risks

**Lane:** doctor fix idempotency, legacy import rewrites, custom data preservation, operator-safe diagnostics.

Reviewed against the Iteration 18 / Attempt 18 draft of `design.md` and `requirements.md` in the active repository workspace.

---

## Executive Summary

The Iteration 18 / Attempt 18 design integrates the rigorous **Attempt 17 Review Resolution Contracts (§2399–2693)**, introducing a robust, multi-layered defense-in-depth model for migration safety. These new contracts explicitly address the critical gaps identified in my previous Iteration 13 review, including:
1. **The Post-Migration Spurious Conflict Loop (§1884–1888):** Fully resolved by the **Runtime-State Migration Binding Table (§2564–2574)**, which specifies a copy-on-absence semantic keyed on whether the Core destination is present, transitioning active control to the new Core path without subsequent back-level digest comparison loops.
2. **Rollout Sequencing Hazards (Slices 2–4):** Fully resolved in **Slice 2: Gas City public-pin adoption and packcompat slice (§3396–3399)**, which explicitly disables legacy whole-file TOML rewriting or routes it through the new mutation coordinator, neutralizing the risk of older binaries running unsafe fixes.
3. **TOCTOU Concurrency Windows (§1873–1879):** Fully resolved by the coordinator's **post-lock digest refresh (§2557)**, ensuring that file states are re-validated under lock before any staged edits are written.
4. **Silent Overwrite of Operator Edits (§1909–1910):** Fully resolved by the "Generated-vs-custom pack detection" logic (§2577–2579), which classifies operator-modified files as "custom local forks" and routes them to manual diagnostics.
5. **Air-Gapped Network Failures (§1869):** Relaxed to validate cache presence via the `public-gastown-pins.yaml` ledger and ordinary remote-cache hits, ensuring offline operators are not blocked (§2528–2535).

While these additions represent an outstanding evolution in the safety architecture of the Gas City doctor subsystem, this independent review highlights three subtle **implementation risks** and **undefined behaviors** that must be carefully managed as implementation begins.

---

## Top Strengths

1. **Structured Runtime-State Migration Binding Table (§2564–2574):**
   The formal binding table clearly separates read paths, write paths, fix behaviors, and rollback behaviors for all critical runtime elements (JSONL export state, JSONL archive repo, push-failure counters, spawn-storm ledgers, and order skips). Specifying that copying is performed *only when the Core destination is absent*, and that *Core wins after the marker is written*, is an elegant, first-principles fix that prevents post-migration divergence from throwing false conflicts.
2. **Defensive Legacy Doctor Containment in Slices 2–3 (§3396–3399):**
   Explicitly mandating that *"legacy whole-file automatic import rewrites are disabled or routed through the new mutation coordinator; unsafe legacy `gc doctor --fix` behavior must not run in slice 2 or 3 binaries"* is a major win for data integrity. This blocks the old, destructive parser from mangling operator comments and TOML layouts.
3. **Default-Deny Production Loader Inventory (§2408–2448):**
   Requiring all production configuration loads to route through `internal/systempacks.LoadRuntimeCity` or generated partial-read helpers listed in `loader-inventory.generated.yaml` prevents low-level bypasses from executing without required Core.
4. **OS advisory locks with post-lock digest refresh (§2555–2562):**
   Combining a crash-released OS directory lock with a mandatory post-lock digest refresh completely closes the Time-of-Check to Time-of-Use (TOCTOU) race window. If a concurrent process completes a write while another is waiting for the lock, the waiting process will immediately detect the changed digests and abort.

---

## Critical Risks & Gaps

### 1. [Minor] Undefined Location and Schema of the Migration "Marker" (§2568–2571)
* **The Risk:** The runtime-state migration table references "conflict is manual before marker" and "Core wins after marker". However, the design does not explicitly define where this marker resides, what its schema is, or how it is written atomically.
* **The Gap:** If the marker is written to `.gc/runtime/packs/core/jsonl-export-state.json` but a crash occurs before the spawn-storm ledger migration completes, the system may enter a partially migrated state. Without an explicit, centralized, single-source-of-truth migration marker file (e.g., `.gc/runtime/migration-marker.json`), different sub-system checks could disagree on whether the city is "before marker" or "after marker."
* **Recommendation:** Mandate that a single, atomic, JSON-formatted migration marker containing the migration generation ID, completion timestamp, and a hash of the migrated files be written to `.gc/runtime/migration.json` as the very last step of the publish phase.

### 2. [Minor] Rigidity of "Custom Local Fork" Refusal for Minor Formatting Edits (§2577–2579)
* **The Risk:** To prevent silent data loss, any file with a digest drift that cannot be classified as binary-generated is treated as a "custom fork" and routed to manual diagnostics.
* **The Gap:** Minor non-semantic formatting edits (e.g., an extra trailing newline or a custom comment added by an operator in a standard pack file) will cause a digest mismatch, classifying the file as a custom fork. The doctor will refuse all automatic fixes for that file, forcing the operator into a manual intervention loop for trivial changes.
* **Recommendation:** Ensure the Retired-Source Classifier has a linting/cleanup preflight that strip-compares formatting, whitespace, and comments when performing digest validation, so that non-semantic changes do not trigger false "custom fork" blocks.

### 3. [Minor] Absence of Standardized Manual Recovery Instructions on Aborted Staged Publish (§2560–2562)
* **The Risk:** The design notes: *"A failed publish records the exact generation and file state so a later `--fix` can resume or report manual recovery without guessing."*
* **The Gap:** While recording the state is excellent for observability, the design does not specify what the "report manual recovery" output actually looks like. If an operator is left with a partially written manifest and no clear instructions on how to restore their city, they may attempt unsafe manual rewrites.
* **Recommendation:** Mandate that when a publish fails, `gc doctor --fix` prints a highly structured, copy-pasteable shell script or `git` commands showing exactly how to revert `.gc/` to its preflight state using the recorded preflight digests.

---

## Evaluation of Sofia's Critical Questions

### 1. Is the Core presence doctor fix a proven no-op on a healthy city, including repeated or concurrent runs with a controller active?
**Yes.** On a healthy city, the pre-resolution TOML parser and the preflight diagnostics stage zero `FixIntent` records, making the write phase a complete byte-identical no-op. Concurrency is fully handled:
* Concurrent runs are serialized via the crash-released OS directory advisory lock.
* Active controller presence is discovered via live process tables rather than stale PID files, preventing concurrent doctor edits while workflows are running.
* The post-lock digest refresh prevents concurrent doctor processes from overwriting each other with stale preflight candidate states.

### 2. When `gc doctor --fix` removes redundant Core or legacy Maintenance imports, what prevents it from deleting user-added imports or custom pack edits?
**The Retired-Source Classifier and CST-preserving TOML editor (§2444–2446, §2577–2579).** The system scans TOML edits at a Concrete Syntax Tree (CST) level. Any user-added imports or custom pack edits are detected via digest mismatched signatures and classified as "custom local forks." The doctor coordinator explicitly refuses automated edits on custom local forks, leaving them entirely untouched and reporting manual diagnostics instead.

### 3. If a local Gastown import is rewritten to a public remote, does the fix verify reachability and immutable provenance or fail with explicit operator guidance?
**Yes.** Provenance and immutability are verified through the three-row `public-gastown-pins.yaml` ledger (§2523–2529) which matches files against cryptographic digests. Reachability is validated during preflight, but is designed to be satisfied offline if a validated ordinary remote-cache hit occurs (§2534). If neither remote registry nor local cache satisfies the pin, the fix coordinator aborts safely and outputs explicit, operator-focused troubleshooting instructions.

---

## Required Changes for Finalization

1. **Define the Migration Marker Schema and Location:** Explicitly add a clause to the Runtime-State Migration section (§2564) requiring a single, atomic, JSON-formatted migration marker at `.gc/runtime/migration.json` to act as the single source of truth for migration completion.
2. **Add Non-Semantic Formatting Normalization to Classifier:** Update §2577 to mandate that the classifier performs basic whitespace and comment normalization before classifying a pack file as a "custom local fork," preventing false blocks on trivial edits.
3. **Specify the Structured Reversion Hint for Aborted Publishes:** Update §2561 to require that if an automated recovery fails after an aborted publish, the command output must include explicit, step-by-step restoration instructions based on the preflight digest log.

---

## Open Questions

1. **How does the live process scan handle containerized/sandbox environments?** If the controller is running inside a separate Docker container or Kubernetes pod than the doctor command, does the live process/runtime discovery accurately identify the active controller?
2. **Is there an automatic cleanup timer for the coordinator's recorded file states in `.gc/tmp/`?** Do stale preflight files get purged after a successful execution, or do they persist indefinitely until manual cleanup?
3. **What happens if a rollback is initiated after the migration marker has been written?** Does the system support a `gc doctor --rollback` option that restores the legacy Maintenance write path and updates the marker, or is rollback strictly a manual operator operation?
