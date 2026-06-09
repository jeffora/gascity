# Camille Okafor — DeepSeek V4 Flash Perspective Independent Review (Iteration 6 / Attempt 6)

**Verdict:** approve-with-risks

**Scope:** Existing-city upgrades, legacy local paths, compatibility shims, version skew, and multi-repo / two-repo rollout mechanics.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-6/design-before.md` / `requirements.md` updated 2026-06-09), the `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the live `examples/gastown/packs/maintenance` assets. I did not inherit prior conclusions; all findings are verified fresh against the repository tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automatic execution tools to write and read from `attempt-1/reviews/migration-rollout-reviewer_gemini.md` (which can block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/migration-rollout-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-6/reviews/migration-rollout-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 6 synthesis.
3. **Verdict Rationale.** The Iteration 6 requirements and implementation plan represent a highly mature, production-ready specification for the Core and Gastown pack split. By decoupling Core as a required, role-neutral host system pack, retiring Maintenance as a standalone active pack, and formalizing robust doctor-owned state migrations, the system successfully eliminates untraceable legacy fallbacks. However, several subtle operational and rollout-safety risks remain: specifically, the deadlock risk for unmaintained transitive dependencies, manual recovery friction when old binaries write after migration, and offline cache-seeding constraints in air-gapped environments. I award an **APPROVE-WITH-RISKS** verdict and mandate three critical improvements to ensure a bulletproof upgrade path.

---

## Top Strengths

- **Thorough, Multi-Slice Rollout Plan (Slices 1a through 7):** The division of the rollout into discrete, logical steps ensures that Gas City never adopts a public pin before its behavior has been verified. The creation of `public-gastown-pins.yaml` and `behavior-preservation.yaml` under `gascity-packs` as external prerequisites ensures a strict, evidence-backed release gate before any source deletion occurs.
- **Durable, Doctor-Owned State Migration:** Offloading the runtime-state migration from startup reloads to `gc doctor --fix --non-interactive` is a major win for stability. The state migration is exceptionally well-designed, featuring a durable migration marker, advisory city locking, and old-binary post-marker write detection that fails closed before any corrupting operations can occur.
- **Bulletproof In-Flight Session Safety (AC10 / AC11):** The requirements and implementation plan now explicitly refuse mutating repairs when live controllers or active tmux sessions are detected. Querying live state (via the process table/lsof) rather than fragile status/PID files adheres perfectly to the project's core design principles and prevents mid-execution file resolution failures.
- **Configurable Core Maintenance Executor (ZFC Compliant):** Replacing Go-side hardcoded assumptions about `dog` with a configurable maintenance executor contract (AC9) and `[gc.bindings.*]` config precedence solves the "hardcoding paradox." The SDK remains fully self-sufficient with only the controller running.

---

## Critical Risks & Gaps

### 1. [Major] Transitive Dependency Deadlock (Upstream Lag)
While the "no compatibility shim and no silent fallback" stance (AC5/AC12) is the correct architectural choice to prevent hidden behavior, it introduces a severe deadlock risk for operators using older, unmaintained third-party remote packs:
- **The Risk:** If a city imports a third-party git pack that has nested imports referencing retired paths like `packs/maintenance`, the city will fail closed during loading.
- **The Gap:** Because these third-party dependencies are read-only and nested, the local operator cannot run a `gc doctor --fix` to rewrite them. Without a local dependency patching/override mechanism (like cargo `[patch]` or npm `overrides`), the operator is completely blocked from upgrading Gas City until the upstream author publishes an update. While this is the decided "unsupported-state" policy, the rollout impact of locking operators out of their own cities must be highlighted.

### 2. [Major] Manual Recovery Friction for Post-Marker Old-Binary Writes
The new binary's ability to detect if an older binary wrote to migrated Core-owned paths (via post-marker write detection) is an excellent safety shield.
- **The Gap:** When this version-skew condition is triggered, the system correctly fails closed and requires "manual reconciliation." However, the exact steps or tooling for this manual reconciliation are not defined. If an operator is left stranded with a corrupt or double-written task store without a clear recovery protocol, it represents a high-friction operational failure.

### 3. [Minor] Air-Gapped / Offline Cache Seeding
AC16 mandates that public Gastown validation is fail-closed and cache-backed, failing when offline if the exact pinned pack is missing.
- **The Gap:** The implementation plan specifies how cache hits and misses are handled, but does not provide a concrete administrative mechanism (such as a cache pre-seeding command, or a pack cache tarball export/import) for offline/air-gapped operators to easily seed the verified cache with the pinned public Gastown pack before running the migration.

---

## Missing Evidence

- **Manual Reconciliation Protocol:** A documented step-by-step instruction or a doctor sub-command that guides operators through resolving post-marker old-binary write conflicts.
- **Air-Gapped Cache-Seeding Mechanism:** An explicit CLI guideline or tool definition for exporting and importing verified pack cache tarballs to support offline upgrades.

---

## Required Changes

1. **Specify Post-Marker Manual Recovery Steps:** Add a requirement to AC10/AC11 stating that when old-binary post-marker writes are detected, the fail-closed diagnostic message must print a clear, step-by-step manual reconciliation procedure (e.g., how to inspect and restore the pre-migration config backup or merge divergent task-store states).
2. **Define the Advisory Lock File Path:** Ensure that the city advisory lock used by the mutation coordinator (`internal/doctorfix`) is bound to a file path that is explicitly exempted from prunes, cache clears, or repairs, preventing race conditions under concurrent CLI doctor invocations.
3. **Provide Guidance for Offline Cache Pre-Seeding:** Extend the offline cache-miss diagnostic under AC16 to include explicit instructions on how an administrator can manually seed the local repo cache with the pinned public Gastown pack in air-gapped environments.

---

## Questions

- What is the exact manual reconciliation flow for an operator whose city has triggered the "old-binary post-marker write" fail-closed diagnostic?
- Is there a plan to provide a simple `gc cache seed` or `gc pack import` command to facilitate air-gapped migrations for offline operators?
