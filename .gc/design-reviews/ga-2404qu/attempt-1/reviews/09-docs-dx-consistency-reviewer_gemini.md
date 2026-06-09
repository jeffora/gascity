# Felix Moreau — Docs & DX Consistency Reviewer (Iteration 20 / Attempt 1, Independent DeepSeek V4 Flash Style Review)

**Verdict:** block

**Scope:** Docs & DX Consistency — documentation consistency, operator terminology, tutorial integrity, and maintenance word disambiguation.

This independent review evaluates the Iteration 20 / Attempt 1 snapshot of the Core and Gastown Pack Split design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md`) against the live codebase at `/data/projects/gascity/` and `requirements.md`.

---

## Executive Summary

As Felix Moreau, the **Docs & DX Consistency Reviewer**, I have conducted a rigorous audit of the revised design in Iteration 20.

The design continues to make major strides in formalizing the **System-Pack Wording Matrix** (`system-pack-wording.generated.yaml`) and establishing a bidirectional, generated documentation linter as a release-gate blocking invariant. Aligning terminology parity within the same implementation slice as its corresponding behavior represents a massive step forward for Gas City's developer experience (DX).

However, from an independent **DeepSeek V4 Flash perspective**, several critical, operator-visible DX contradictions, factual baseline errors, and material gaps remain in the Proposed Design text itself. The design recommends a ripgrep command that mathematically guarantees critical stale-path files will escape linting, frames a completely non-existent reference page as currently existing under "Current System," and fails to include high-priority tutorial/troubleshooting docs in the named update scope.

Furthermore, given that other lanes such as Sofia's Doctor review are currently a firm **BLOCK**, and these documentation gaps directly correlate with outstanding semantic and structural contradictions in the design, my verdict remains a firm **BLOCK**.

Until these contradictions are resolved and the files are explicitly added to the named update scopes, this lane remains a **BLOCK**.

---

## Top Strengths

1. **Executable Wording Matrix:**
   Transitioning the wording matrix from loose prose into a structured, generated artifact (`plans/core-gastown-pack-migration/system-pack-wording.generated.yaml`) governed by a schema and validated in CI via `TestSystemPackWordingFresh` is an outstanding victory. Enforcing this as a hard gate ensures developer consistency.
2. **Hardened Release Gates:**
   Forcing docs updates, navigation, and golden tests to move in the same slice as the corresponding code changes—and enforcing this via a real release gate—guarantees that human-facing documentation cannot drift from the system's actual behavior.
3. **Retired-Directory Preservation Policy:**
   Preserving `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` as ignored legacy state rather than deleting them protects existing operator environments and prevents catastrophic silent data loss on upgrade.

---

## Lane-Specific Detailed Responses

### Q1: Do doctor output, import-state messages, docs, pack comments, and tutorials use the same language for Core, public Gastown, and retired Maintenance?

**No. While the wording matrix framework is structurally designed to enforce consistency, the Proposed Design text retains direct vocabulary and concept contradictions that will confuse operators and fail linting.**

1. **"maintenance-worker" vs "maintenance agent" Contradiction**:
   * **The Conflict**: The Go binding defined in the design is explicitly `[gc.bindings.maintenance_worker]` with override key `maintenance_worker` and diagnostic prefix `core.maintenance_worker`. Yet, the design repeatedly and interchangeably refers to this as the "maintenance-agent" or "maintenance agent".
   * **Operator Impact**: If an operator reads docs or tutorials instructing them to configure the "maintenance agent", they will naturally attempt to set a non-existent configuration key (`maintenance_agent`). The vocabulary must be unified strictly under "maintenance-worker" or "maintenance worker" to align with the technical binding, or the allowed/forbidden contexts for both terms must be explicitly declared and validated in the matrix.

2. **Case-Aware Seed Violation**:
   * **The Conflict**: The design instructs changing the `cmd/gc/import_state_doctor_check.go` messaging to `"maintenance is retired; Core supplies generic maintenance and Gastown supplies Gastown-specific behavior"`.
   * **Operator Impact**: This seeded phrase uses lowercase "maintenance is retired" for the retired pack. Under the case-aware wording matrix rules, lowercase "maintenance" is reserved for ordinary English, while uppercase "Maintenance" is reserved for the retired pack. The design's own seed phrase would be rejected by its linter. It must be updated to `"Maintenance is retired; Core supplies generic maintenance..."`.

---

### Q2: Can a new operator complete tutorial 01 and troubleshooting flows without encountering retired local pack paths or contradictory Maintenance guidance?

**No. A new operator using a Core-only city template is guaranteed to encounter broken walkthroughs, dead links, and retired local pack paths because high-priority files are omitted from the update scope.**

1. **Tutorial Walkthrough Integrity (Tutorial 05 and Tutorial 07)**:
   * **The Conflict**: The design only scopes updating `docs/tutorials/01-cities-and-rigs.md`. However, other tutorials contain critical, hardcoded references to retired or moved assets.
   * **Tutorial 05**: `docs/tutorials/05-formulas.md` references formulas like `mol-dog-backup/-compactor/-reaper` that this migration moves to Core and/or rebinds to `core.maintenance_worker`, and references the retired "polecat worker scaffolds".
   * **Tutorial 07**: `docs/tutorials/07-orders.md` lists `prune-branches` as an expected order. But `prune-branches` is being migrated to public Gastown and will be completely absent on a Core-only template city. An operator completing Tutorial 07 on a Core-only setup will experience a broken run.

2. **Stale Troubleshooting Prose and Fallbacks**:
   * **The Conflict**: `docs/getting-started/troubleshooting.md` contains prose attributing active behavior to the retired pack: "The maintenance pack runs `jsonl-export` every 15 minutes", and instructs operators to configure `GC_JSONL_MAX_PUSH_FAILURES=99` "in the maintenance pack's environment".
   * **Operator Impact**: Operators will attempt to configure retired pack environments to handle push failures. The troubleshooting guide must be updated to re-attribute `jsonl-export` and push-failure environment variables to Core, and must explain the tool's Core→legacy→`.gc/jsonl-*` state fallback order rather than pointing operators blindly to a retired path.

3. **MDX Troubleshooting Surfaces and Dead Links**:
   * **The Conflict**: `docs/troubleshooting/gc-start-walkthrough.mdx` instructs operators facing a missing-agent error to use `includes = ["packs/gastown"]` and describes duplicate definitions as originating from the auto-imported system pack under `.gc/system/packs/gastown/`. Post-migration, Gastown is never auto-imported and `packs/gastown` is retired.
   * **The Coming-from-Gastown Dead Link**: `docs/getting-started/coming-from-gastown.md` contains a live GitHub link to `examples/gastown/packs/gastown/pack.toml` on `main`, which will 404 immediately upon the landing of the source-deletion slice.

---

### Q3: Do docs preserve store-maintenance or Dolt maintenance terminology only where it still refers to those systems rather than the retired Maintenance pack?

**Partially. While the design captures this principle, it lacks the concrete validation mechanics to distinguish these terms safely during linting.**

1. **Matrix Allowed-Context Omissions**: The wording matrix needs to explicitly allow:
   * The `gc maintenance` CLI command family.
   * `docs/runbooks/dolt-maintenance.md`.
   * Generated schema identifiers like `MaintenanceConfig` and `DoltMaintenance` in `docs/schema/openapi.json`.
2. **Outdated Schema Snippets**: `docs/schema/pack-schema.json` describes `includes` using a legacy example: `(e.g. "../maintenance")`. This is a pre-existing generated file, but leaving it as-is teaches operators the exact retired import pattern we are migrating away from.

---

## Blocker & Critical Gaps

### [Blocker] The Ripgrep Glob Gap Contradiction
- **Severity:** blocker
- **Confidence:** high
- **Quality dimension:** correctness
- **Gate impact:** blocker
- **Evidence:** `.gc/design-review-inputs/core-gastown-pack-migration/design.md`:3222
- **Pattern comparison:** The design mandates an extension-agnostic docs scan covering Markdown, MDX, JSON, and TXT. Yet, the recommended inventory-building command explicitly restricts files via: `-g '*.md' -g '*.toml' -g '*.go' -g '*.sh'`
- **Why it matters:** This glob selection mathematically guarantees that `.mdx` files (like `gc-start-walkthrough.mdx`), `.json` files (like the docs navigation index `docs.json` or generated schemas), and `.txt` files will escape the baseline inventory. Stale paths and contradictory terms will hide in these files until they trigger a CI failure during the final docs-validation gate.
- **Suggested fix:** Update the recommended inventory command on line 3222 to cover all documentation formats:
  `rg -n "maintenance|system/packs|runtime/packs|gastown|PublicGastown|dog|Core" docs examples cmd internal -g '*.md' -g '*.mdx' -g '*.toml' -g '*.go' -g '*.sh' -g '*.json' -g '*.txt'`

### [Blocker] Broken Baseline Claims for `system-packs.md`
- **Severity:** blocker
- **Confidence:** high
- **Quality dimension:** correctness
- **Gate impact:** blocker
- **Evidence:** `.gc/design-review-inputs/core-gastown-pack-migration/design.md`:3242
- **Pattern comparison:** The design refers to `docs/reference/system-packs.md` under "Examples And Docs" as a page to update. In reality, `docs/reference/system-packs.md` does **not** exist in the active tree of the checkout.
- **Why it matters:** Since this page does not exist, it cannot be merely "updated." This page must be created from scratch, and it must be registered in the docs navigation `docs/docs.json`. Confusing "create" and "edit" means the work is mis-scoped. The design must explicitly frame this as a new document creation and nav-registration rather than an edit, ensuring it lands in the first operator-facing behavior-changing slice.
- **Suggested fix:** Frame the creation of `docs/reference/system-packs.md` clearly as a new document creation (and nav-registration) rather than an edit, and ensure it lands in the first operator-facing behavior-changing slice.

---

## Required Changes & Actions

Before this design can be approved, the following modifications must be made:

1. **Unify terminology**: Standardize operator vocabulary strictly under "maintenance-worker" (matching the `maintenance_worker` binding) or define a clear allowed/forbidden context boundary for "maintenance agent" vs. "maintenance worker" in the wording matrix.
2. **Fix the inventory glob**: Update the recommended inventory command on line 3222 to cover all documentation formats:
   `rg -n "maintenance|system/packs|runtime/packs|gastown|PublicGastown|dog|Core" docs examples cmd internal -g '*.md' -g '*.mdx' -g '*.toml' -g '*.go' -g '*.sh' -g '*.json' -g '*.txt'`
3. **Correct the `system-packs.md` baseline**: Frame the creation of `docs/reference/system-packs.md` clearly as a new document creation (and nav-registration) rather than an edit, and ensure it lands in the first operator-facing behavior-changing slice.
4. **Expand named update scopes**: Explicitly add the following files to the named docs-update inventory:
   * `docs/troubleshooting/gc-start-walkthrough.mdx` (to fix `includes` guide and stale system-packs paths)
   * `docs/tutorials/05-formulas.md` (to align formula names and remove retired scaffolds)
   * `docs/tutorials/07-orders.md` (to adjust expected `gc order list` output to the Core-only template)
   * `docs/getting-started/coming-from-gastown.md` (to correct dead GitHub links)
   * `docs/getting-started/troubleshooting.md` (to correct legacy environment paths and fallback order)
5. **Add wording-matrix allowed contexts**: Ensure `gc maintenance` CLI command, `dolt-maintenance.md`, and generated schema definitions are registered as allowed contexts.

---

## Questions

* **Does the tutorial-template city show `prune-branches` at all (via public Gastown import), or is the template strictly Core-only?** If Core-only, Tutorial 07's expected output must be updated to remove `prune-branches` entirely.
* **Who is responsible for regenerating `docs/schema/*.json` files when config doc comments are edited?** This must be sequenced to prevent schema docs from drifting.
