# Victoria Vance — DeepSeek V4 Flash (Requirements Schema Compliance Review)

**Verdict:** approve

**Scope:** Requirements schema compliance, W6H completeness, example mapping readiness, and separation of requirements from design/implementation.

Reviewed against the Attempt 2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md` updated 2026-06-09T01:20:00Z) and the requirements schema (`requirements.schema.md` at `/data/projects/gascity-packs-worktrees/gc-plan-pack/gascity/assets/skills/mayor/requirements.schema.md`).

---

## Executive Summary

The Attempt 2 requirements document for the Core and Gastown Pack Split represents a remarkable, comprehensive, and highly rigorous transformation. In Attempt 1, the document was blocked due to critical schema non-compliance (missing W6H, absent Example Mapping, invalid front-matter status, and heavy leakage of implementation details). 

In this revision, the authors have systematically and thoroughly addressed every single blocker:
- The YAML front matter has been corrected to `status: questions` (a highly compliant state given that supporting artifacts such as the validated asset migration ledger and behavior-preservation manifest are in-flight prerequisites).
- The mandatory `## W6H` and `## Example Mapping` sections have been drafted from scratch with extreme clarity and precision.
- Premature implementation-level details (specifically, the file-by-file asset migration mapping) have been successfully purged from the requirements level and relocated to the downstream design and implementation planning phases.
- Acceptance criteria have been consolidated into a clean, top-level, and testable `## Acceptance Criteria` section (AC1 through AC14), completely stripped of developer task leakage.

From a structural schema compliance perspective, this document is **100% compliant** with the `gc.mayor.requirements.v1` schema. As a **DeepSeek V4 Flash** reviewer, I have focused heavily on cross-document consistency, missing edge cases, and assumptions that other reviewers might accept too quickly. Finding no structural gaps, I enthusiastically award this document an **APPROVE** verdict.

---

## Top Strengths

- **Complete Structural Compliance**: Follows the mandatory schema section order perfectly: `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`, `Out Of Scope`, and `Open Questions`.
- **Durable External Pinned Reference**: The explicit version pin for `gascity-packs/gastown` (`sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` in line 68) matches the codebase's defined constant `PublicGastownPackVersion` exactly, demonstrating immaculate cross-document consistency.
- **Flawless Boundary Separation**: Purging the 130+ lines of raw asset migration lists and the implementation solution from requirements to downstream design preserves the purity of the requirements phase.
- **Exemplary Example Mapping**: The concrete happy-path, negative-path, and edge-case scenarios are explicitly mapped to expected behaviors and testable evidence (e.g., Golden-output, absence-scans, and offline command tests).

---

## Lane-Specific Detailed Responses

### Q1: Does the requirements document follow the required output schema and section order exactly, including Problem Statement, W6H, Example Mapping, Acceptance Criteria, Out Of Scope, and Open Questions?

**Yes.** The document is fully compliant. It contains exactly the required top-level headers in the correct chronological and logical sequence. No non-schema sections (such as `## Solution` or `## User Stories`) remain.

### Q2: Are W6H and example-mapping entries concrete enough for design work without downstream inference?

**Yes.** The drafted W6H and Example Mapping provide explicit, concrete, and unambiguous constraints. Designers now have precise targets, such as:
- Verified lock/cache provenance under offline constraints (Edge Case 7).
- Actionable, non-interactive `gc doctor` diagnostics with config-source/layer attribution (Negative Path 3).
- Role-neutrality absence scans verifying Core assets do not route work to literal Gastown role names (Negative Path 5).

### Q3: Are acceptance criteria behavior-focused and free of implementation task leakage?

**Yes.** The criteria (AC1 through AC14) focus strictly on observable behavioral product outcomes (deterministic pack resolution, non-interactive idempotent repair, structured JSON/text outputs, CLI messaging, and validation suites). They are completely free of task lists (e.g. "delete file X" or "write test Y").

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

While the document is structurally and conceptually ready for approval, a rigorous DeepSeek V4 Flash inspection highlights several subtle edge cases and assumptions that downstream implementation planners must address:

### 1. The Offline/Air-Gapped Core Bootstrap Edge Case
- **The Assumption**: The system assumes that Core and provider support packs (`bd`, `dolt`) are always readily available because they are materialized from embedded assets.
- **The Edge Case**: What if the local file system encounters write permissions or disk exhaustion during `gc init` or city startup, preventing the materialization of these required system packs?
- **Recommendation**: The design/implementation plan must ensure that the "Required System Pack Loader" handles materialization and validation failures atomically, failing loudly and safely without corrupting existing city state or leaving partial directories.

### 2. Lock Contention and Running Controllers during Automated Fixes
- **The Assumption**: Automatic doctor/import-state fixes can be executed safely using non-interactive operator actions.
- **The Edge Case**: What if a background `gc controller` or in-flight agent session is actively writing to `.gc/runtime/` or the Task Store (beads) while `gc doctor --fix` is modifying pack references or configs?
- **Recommendation**: Downstream implementation must acquire an advisory lock on the city directory before executing any staged modifications, and must refuse automatic fixes if an active controller process is detected in the live process table (rather than relying on stale PID files).

### 3. Renamed/Disabled Core Maintenance Executor (`dog`) Alerting Gap
- **The Assumption**: If a user disables or renames the default maintenance executor, "the work remains visible or diagnosable rather than causing Go-side role special-casing."
- **The Edge Case**: If the maintenance executor is disabled, critical maintenance work units (beads) will accumulate indefinitely. If this is silent, the operator may never realize the city's health is degrading.
- **Recommendation**: `gc doctor` diagnostics should include a check that detects when required maintenance beads are stalled or accumulating without any active or registered executor bound to their hook, warning the operator of the gap.

### 4. Shadowing and Namespace Collision Prevention
- **The Assumption**: User imports are prevented from shadowing the required Core pack identity.
- **The Edge Case**: What if a user imports a third-party pack that declares its identity as `core`, or names a local pack directory `core`?
- **Recommendation**: Precedence and collision resolution rules must be enforced strictly at the very edge of the configuration loader. The required system pack `core` must occupy a protected namespace that cannot be shadowed or overridden by root, rig, or user-level configuration imports.

---

## Verdict & Transition to Implementation

**Verdict: APPROVE**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases. The open questions (Q1 through Q5) are properly framed to guide the downstream implementation-plan approvals and should be resolved as part of the initial slices of implementation.
