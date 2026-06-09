# Oleg Marchetti — DeepSeek V4 Flash Perspective Independent Review (Iteration 5 / Attempt 5)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, before-after mapping, requester/detector/notification continuity, and preventing silent capability loss.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-5/design-before.md`, 135 lines, status updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` dog assets this migration retires, and the public `gascity-packs/gastown` pack source. I did not inherit prior conclusions; all findings are re-verified against the tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect documented in `attempt-4/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/gastown-behavior-preservation-auditor_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/gastown-behavior-preservation-auditor_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 5 synthesis.
3. **Verdict Rationale.** The requirements draft for Attempt 5 represents a monumental step forward, completely resolving all five of the previous major open product questions and establishing a robust schema for the AC6 Asset Ledger and AC7 Behavior Manifest. However, from the strict perspective of **Behavior Preservation Auditing**, several highly subtle, cross-document inconsistencies and edge cases remain unaddressed (such as the exclusion of scripts in AC14 and the unmitigated CI Generator Freshness Trap). I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins to ensure Gastown behavior is preserved without silent capability loss.

---

## Evaluation of the Three Key Questions

### 1. Does every generalized Core asset have a corresponding external Gastown home for stripped role-specific behavior?
**Auditor Finding: Yes.**
The requirements mandate a strict multi-repo rollout sequence (Slice 1 prerequisite in downstream plans) to ensure that no source deletion or de-roling of the SDK occurs until Gastown-owned formulas, orders, scripts, prompts, and overlays are safely landed in the public pack at immutable commits. AC6 has been strengthened to ensure the ledger tracks "split boundary, fallback classification, rationale, and proof command" at file granularity, preventing orphaned or stripped behaviors.

### 2. Does the before-and-after inventory cover formulas, orders, scripts, prompts, template variables, and notification paths rather than only file moves?
**Auditor Finding: Yes.**
AC7 now explicitly covers "formulas, orders, scripts, prompts, template variables, notification targets, requester/detector relationships, identity side effects, success/warning/failure/escalation paths, and recovery flows." By explicitly including "identity side effects" (e.g., Git commit author identity) and "escalation paths," the before-and-after tracking captures logical behavior rather than mere directory organization.

### 3. What artifact proves supported Gastown workflows still resolve and trigger after the split?
**Auditor Finding: Yes.**
The machine-readable AC7 **Behavior Preservation Manifest** and the accompanying test/packcompat harness represent the definitive proof. The requirements now mandate "executable runtime or command checks that the external Gastown pack loads, resolves, renders, triggers, routes, notifies, and exercises failure/recovery paths with in-tree fallback disabled."

---

## Critical Risks & Missing Edge Cases (Auditor Findings)

### 1. [Major] The CI Generator Freshness Trap (Self-Defeating Validation)
* **The Risk:** AC6 (ledger) and AC13 (legacy test coverage transfer) require validation commands to assert that all legacy files and assertions have been re-homed. The ledger "fails on stale source snapshots" and AC13 "fails on unmapped retired assertions."
* **The Gap:** Once the legacy folders are physically deleted under `internal/bootstrap/packs/core` and `examples/gastown/packs/maintenance`, any live file-system walks on subsequent commits or PRs will find zero files. The generator and validators will either crash, report empty sets, or silently pass empty outputs.
* **The Fix:** The requirements must explicitly mandate that the ledger/test validators run against a **frozen historical reference snapshot** (such as a specified baseline Git commit or a cryptographically hashed local snapshot checked into the repository), ensuring that physical deletions do not defeat the completeness checks.

### 2. [Major] Exclusion of Scripts from AC14 Public checkout validation
* **The Risk:** AC14 requires the public checkout or pinned cache to prove "roles, prompts, commands, formulas, orders, overlays, and checks" to prevent local files from masking broken external packs.
* **The Gap:** AC14 completely omits **scripts**. Gastown operations rely heavily on standalone shell scripts (`reaper.sh`, `jsonl-export.sh`, `spawn-storm-detect.sh`) that contain the actual `gc mail send` and `gc session nudge` execution blocks. If scripts are not in AC14's checkout proof set, an absent or broken script in the public pack could be masked by a lingering local copy during testing.
* **The Fix:** Add "scripts" (and associated asset scripts) explicitly to the AC14 enumerated proof set.

### 3. [Minor] Silent Execution Failures in Generalized Scripts (Empty/Slash Recipient Guard)
* **The Risk:** Scripts like `reaper.sh` and `jsonl-export.sh` are generalized to consume recipients dynamically from formula/order metadata.
* **The Gap:** If a recipient target is left unconfigured, is empty, or evaluates to `/`, executing commands like `gc mail send ""` or `gc mail send /` inside the shell scripts will cause unhandled script crashes.
* **The Fix:** All generalized shell scripts must perform preflight verification of their recipient variables. If the target is empty or invalid, the script must gracefully log an audit warning to `stderr` and skip mail execution (exiting with code `0`) rather than crashing the entire workflow.

---

## Required Changes for Finalization

1. **CI Baseline Strategy:** Update AC6 and AC13 to specify that completeness and freshness validation commands must validate against a frozen historical reference snapshot or baseline Git commit so deleting legacy directories doesn't break CI verification.
2. **Script Inclusion in AC14:** Add "scripts" explicitly to AC14's list of checked assets.
3. **Script Recipient Guard:** Mandate that generalized scripts handle empty/slash recipients by logging a warning and skipping execution gracefully.
