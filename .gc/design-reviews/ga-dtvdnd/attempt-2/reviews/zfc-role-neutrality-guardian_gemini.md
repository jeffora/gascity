# Alistair Sterling - DeepSeek V4 Flash (Gemini 3.5 Flash (High))

**Verdict:** approve-with-risks

### Top Strengths:
- **Systematic Schema Conformance**: The requirements document has been successfully restructured to conform exactly to the required `gc.mayor.requirements.v1` schema. The addition of structured front-matter, `W6H`, `Example Mapping`, and consolidated `Acceptance Criteria` completely resolves the process/format blocker from Iteration 1.
- **Explicit Role Decoupling (AC8 & AC9)**: The requirements in AC8 and AC9 explicitly establish role neutrality as a testable, automated invariant across Go business logic and Core assets, while clarifying that the default `dog` maintenance executor is a user-configurable pack default rather than a Go-side special case.
- **ZFC Controller Self-Sufficiency**: The edge cases and AC2 successfully codify that SDK infrastructure operations (gate evaluation, order dispatch, health patrol, bead lifecycle, doctor) must remain functional under the controller alone, independent of whether a specific executor or Gastown agent is active or even exists.

### Critical Risks:
- [Major] **Literal `dog` Leakage in Core Assets (Lack of Symbolic Bindings)**:
  While AC8 prohibits literal `dog` routing and AC9 mandates that `dog` is a configurable default, the actual binding mechanism inside Core formulas and orders remains a major risk. If a Core formula (e.g., `mol-dog-jsonl.toml` or `mol-dog-reaper.toml`) contains a hardcoded literal `assignee = "dog"` in its routing metadata, then renaming the default maintenance agent to `reaper-bot` in `city.toml` will immediately break that Core formula. Operators would be forced to duplicate and manually modify every single Core-owned formula to replace `"dog"` with their custom name. The requirements must explicitly mandate that **all Core-owned maintenance formulas and orders must reference their executor through a symbolic binding or template variable (e.g., `assignee = "{{.MaintenanceExecutor}}"` or `core.maintenance_worker`) rather than a literal string `"dog"`**.
- [Major] **Cross-Document Inconsistency in Public Gastown Repository Authority**:
  There is a direct naming and path mismatch throughout the requirements document regarding the external repository for Gastown:
  - In W6H (`Where` row) and AC14, it refers to `gascity-packs/gastown` (implying the GitHub organization is `gascity-packs`).
  - In W6H (`How` row) and AC4, it specifies the canonical URL as `https://github.com/gastownhall/gascity-packs.git//gastown` (implying the GitHub organization is `gastownhall` and the repository is `gascity-packs`).
  To prevent misdirected resolution logic, broken builds, or incorrect package-fetching behaviors, the requirements must standardize on a single, correct path and organization (the `gastownhall/gascity-packs` format) consistently.
- [Minor] **Bead Accumulation and Lifecycle under Disabled Executor**:
  Under GUPP, if a user disables or omits the default maintenance executor, the edge case states that "the work remains visible or diagnosable rather than causing Go-side role special-casing." However, the requirements do not specify whether the controller should actively suppress creating new maintenance beads when no executor is active, or if it should enforce a strict retention/compaction policy. Allowing maintenance beads (e.g., wisp compaction, gate sweep) to pile up indefinitely can degrade task store performance and generate false-positive health-patrol alerts on the event bus.

### Required Changes:
1. **Mandate Symbolic / Template Bindings in Core Assets**:
   Refine AC8 and AC9 to explicitly require that all Core-owned formulas, orders, and routing metadata must reference the maintenance executor via a symbolic binding slot (e.g., `core.maintenance_worker` or a configurable templated variable) instead of a hardcoded `"dog"` target, completely preventing literal `dog` leakage inside Core-owned pack assets.
2. **Reconcile Gastown Repository Path and Owner**:
   Update all occurrences of the Gastown repository in the requirements (W6H `Where` row and AC14) to consistently match the canonical `https://github.com/gastownhall/gascity-packs.git//gastown` URL.
3. **Specify Controller Behavior for Disabled Executor**:
   Add an acceptance criterion or edge-case specification detailing how the controller manages bead generation when the maintenance executor is disabled (e.g., suppressing wisp compaction or sweep beads to prevent database bloat, or applying a strict retention policy).

### Questions:
- If an operator overrides or renames the default `dog` executor in `city.toml`, does the `gc doctor` command validate that the new name is bound to a valid active provider, or does it only warn if the default `dog` is completely missing?
- Can the AST-based absence scanner mandated by AC8 be extended to automatically scan Core-owned TOML/YAML files to assert that no literal `"dog"` strings appear in the `assignee` or routing fields?
