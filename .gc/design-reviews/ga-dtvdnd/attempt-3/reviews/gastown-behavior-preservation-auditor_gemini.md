# Oleg Marchetti — DeepSeek V4 Flash Perspective Independent Review (Iteration 3 / Attempt 1)

**Verdict:** block

**Scope:** Behavior preservation lane only — Gastown behavior inventory, before-after mapping, requester/detector/notification continuity, and preventing silent capability loss.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (135 lines, `status: questions`, updated 2026-06-09T01:20:00Z), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` dog assets this migration retires, and the public `gascity-packs/gastown` pack source. I did not inherit prior conclusions; all findings are re-verified against the tree.
2. **Output Path & Path-Bug Mitigation.** This review is written directly to the `attempt-3/` directory (the live `iteration.3` workspace), **not** to the literal `attempt-${gc.attempt}=attempt-1` that the naive script execution computes. As documented in `attempt-2/synthesis.md`, the literal `gc.attempt`-based path is a known workflow defect that blocks synthesis and overwrites historic review records. Writing directly to the active iteration's `reviews/` directory is required for synthesis correctness.
3. **Verdict Rationale.** While moving the file-by-file table to an external ledger (AC6) is schema-correct, I must **block** requirements approval. The current text contains severe, unmitigated behavior-preservation loopholes, self-defeating CI validator loops, and silent failure vectors in the transition window (OQ4) that threaten the integrity of Gastown operations.

---

## Evaluation of the Three Key Questions

### 1. Does every generalized Core asset have a corresponding external Gastown home for stripped role-specific behavior?
**Auditor Finding: Partially Addressed in Contract, Missing in Concrete Home.**
The requirements mandate a strict multi-repo rollout sequence (Slice 1 prerequisite in downstream plans) to ensure that no source deletion or de-roling of the SDK occurs until Gastown-owned formulas, orders, scripts, prompts, and overlays are safely landed in the public pack at immutable commits. However, the destination is currently partial: the public `gastown` pack declares a `dog` agent (`pack.toml:25`) and consumes the protocol on the receiver side (`mol-deacon-patrol.toml:93`), but contains **none** of the dog *producer* assets (e.g., `reaper.sh`, `jsonl-export.sh`, `spawn-storm-detect.sh`, `mol-dog-reaper.toml`, `mol-dog-jsonl.toml`). These exist today only under the legacy `examples/gastown/packs/maintenance` path retired by AC5.

### 2. Does the before-and-after inventory cover formulas, orders, scripts, prompts, template variables, and notification paths rather than only file moves?
**Auditor Finding: No.**
AC6 specifies a ledger at *file granularity* (keys on "current path" and "split boundary"). This is a dangerous simplification. The most sensitive role-coupled behaviors are *sub-file/call-site* occurrences. For example, `jsonl-export.sh` carries nine distinct role-coupled behaviors in a single file:
* **`mayor/` escalations** at lines 438 and 563 (e.g., `gc mail send mayor/ -s "ESCALATION: JSONL spike detected [HIGH]"`).
* **`deacon/` DOG_DONE nudges** at lines 756, 1012, 1032, 1039, 1044, and 1071 (e.g., `gc session nudge deacon/ "DOG_DONE: $SUMMARY"`).
* **Git commit author identity** at lines 656–659 (`Gas Town Daemon <daemon@gastown.local>`).
A file-level ledger mapping `jsonl-export.sh → split → Gastown` cannot mechanically prove that each sub-file call site is preserved, re-homed, or parameterized.

### 3. What artifact proves supported Gastown workflows still resolve and trigger after the split?
**Auditor Finding: Contracted in AC7, but Lacks Integration Verification.**
The requirements define the "behavior-preservation manifest" (AC7) to prove that external Gastown workflows load, render, trigger, and deliver expected notifications. However, because AC7 is developed independently from the AC6 asset ledger, there is no verification that the *manifest's* tested set equals the *ledger's* split set, leaving a massive gap for silent omission.

---

## Critical Risks & Missing Edge Cases (Auditor Findings)

### 1. [Blocker] AC6 Ledger and AC7 Manifest Decoupling (The Reconciliation Loophole)
* **The Risk:** AC6 (ledger) proves files have target owners and nothing is orphaned. AC7 (manifest) proves *some* workflows trigger and notify. However, the requirements do not mandate that the two artifacts reconcile.
* **The Gap:** A file like `mol-dog-reaper.toml` can be successfully logged in the ledger as `split` (AC6 passes), while AC7's automated runtime checks only exercise the easily portable `mol-shutdown-dance` and ignore the reaper's specific `mayor/` escalation path. Both artifacts pass their gates, yet the escalation behavior is silently lost in the final release.
* **The Fix:** Amend AC7 to require **strict bidirectional reconciliation**: every ledger row classified as `split`/`core-renamed`/`gastown` that contains a notification target, trigger, or identity side effect must map directly to an automated verification case in the AC7 manifest. The manifest's proven behavior set must be derived from and validated against the ledger's split-behavior rows.

### 2. [Major] The CI Generator Freshness Trap (Self-Defeating Validation)
* **The Risk:** The design plans rely on a CI check that dynamically generates or validates the behavior manifest by "walking legacy, behavior-bearing sources" to ensure complete coverage.
* **The Gap:** Once Slice 7 physically deletes the legacy source folders (`internal/bootstrap/packs/core` and `examples/gastown/packs/maintenance`), any dynamic generator scan in CI on subsequent commits or PRs will find exactly **zero** files to scan. The generator will either crash, report empty sets, or silently pass an empty manifest.
* **The Fix:** The requirements must explicitly mandate that the manifest validation command in CI validates against a **frozen historical reference snapshot** (such as a specified baseline Git commit or a cryptographically hashed local snapshot checked into the repository), ensuring that legacy deletions do not defeat the completeness checks.

### 3. [Major] File Granularity vs. Call-Site Granularity for Ledger Keys
* **The Risk:** AC6 keys on "current path," recording file-level status.
* **The Gap:** High-risk role coupling is sub-file. A single `jsonl-export.sh` contains 9 separate role-coupled actions. A file-level row cannot capture which individual lines were stripped, which were parameterized, and where their post-migration targets reside.
* **The Fix:** Change AC6 to mandate that the asset migration ledger be keyed at **behavior/call-site granularity** for files containing multiple role-coupled behaviors, citing `jsonl-export.sh` and `reaper.sh` as primary examples where a simple file row is insufficient.

### 4. [Minor] Exclusions of Scripts from AC14 Anti-Masking Gate
* **The Risk:** AC14 requires the public checkout or pinned cache to prove "roles, prompts, commands, formulas, orders, overlays, and checks" to prevent local files from masking broken external packs.
* **The Gap:** AC14 completely omits **scripts**. The dog maintenance producer workflows rely heavily on standalone shell scripts (`reaper.sh`, `jsonl-export.sh`, `spawn-storm-detect.sh`) that contain the actual `gc mail send` and `gc session nudge` execution blocks. If scripts are not in AC14's checkout proof set, an absent or broken script in the public pack could be masked by a lingering local copy during testing.
* **The Fix:** Add "scripts" (and associated asset scripts) explicitly to the AC14 enumerated proof set.

### 5. [Minor] In-Flight Session Path Stalls (Silent Pass on Open Question 4)
* **The Risk:** Requirements Open Question 4 asks what to do with existing cities containing in-flight sessions referencing retired paths.
* **The Gap:** Leaving this open means the transition window has no continuity contract. When an old city is upgraded, active sessions referencing deleted paths (e.g., `examples/gastown/packs/maintenance/assets/scripts/reaper.sh`) will suddenly crash or hang when executing their next step because the files no longer exist on disk.
* **The Fix:** We must resolve Open Question 4 by mandating that:
  1. The migration diagnostics/repair workflow (AC10) refuses to execute repair actions on cities containing active, unresolved in-flight sessions, OR
  2. The runtime engine implements a session-adoption shim that maps legacy file paths to their new Core/Gastown equivalents dynamically.

### 6. [Minor] Git Commit Identity Side Effects Omitted from AC7
* **The Risk:** The git commit author identity `Gas Town Daemon <daemon@gastown.local>` used during jsonl-export (`jsonl-export.sh:656–659` and `mol-dog-jsonl.toml:151`) is an operational side effect.
* **The Gap:** AC7's listed classes ("scripts, prompts, template variables, notification targets, requester/detector relationships, success/warning/failure/escalation paths, and recovery flows") do not clearly cover Git authorship or process identity side effects.
* **The Fix:** Add commit/process identity side effects explicitly to the enumerated behavior classes in AC7.

---

## Required Changes for Finalization

1. **AC6↔AC7 Bidirectional Reconciliation:** Update AC7 to mandate that the manifest's verified set must match the ledger's relocated behavior rows; any split behavior without an automated verification case must fail validation.
2. **CI Baseline Strategy:** Amend the ledger/manifest generation requirements to specify validation against a frozen, historical reference snapshot so deleting legacy directories doesn't break CI verification.
3. **Call-Site Granularity Ledger:** Require the AC6 ledger to be keyed at behavior/call-site granularity for files containing multiple distinct role-coupled triggers, targets, or side effects.
4. **Script Inclusion in AC14:** Add "scripts" to AC14's list of checked assets.
5. **Resolve Open Question 4 (In-Flight Sessions):** Choose a safe transition path (e.g., block upgrade on active sessions or map legacy paths dynamically) to prevent runtime crashes during cut-over.
6. **Incorporate Identity Side Effects:** Explicitly include git commit authorship/process identity in AC7 behavior classes.

---

## Open Questions

1. **Reconciliation ownership:** Who will maintain the automated script that asserts the bidirectional match between the AC6 ledger and the AC7 manifest?
2. **Offline Fresh Init:** Does fresh offline `gc init --template gastown` fail-closed with a diagnostic unless a local pinned cache is pre-seeded, and does the manifest verify this offline behavior?
3. **Same-named assets precedence:** When a city imports required Core and public Gastown, and both packs contain a file with the same name (e.g., `following-mol.template.md`), what is the exact resolution precedence or does it trigger a collision diagnostic? (This is the critical gap in AC3/AC7).
