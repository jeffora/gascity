# Hugo Bautista — Gemini (Asset Classification & Split Review)

**Verdict:** block

**Scope:** Split and core-renamed ownership; file-by-file migration table completeness; clean halves; resolution of review-marked assets.

> ### Lane Note (Verify-Don't-Copy + Path Handling)
> 1. **Re-grounding & Independence:** This review is an independent Gemini 3.5 Flash evaluation of the Attempt 4 requirements (`.gc/design-reviews/ga-dtvdnd/attempt-4/design-before.md`) and draft implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md` updated 2026-06-09T07:28:00Z). I have not inherited conclusions from prior rounds. I evaluated AC6, the Example Mapping, and verified the split/duplicate hazards against the live source tree to expose critical risks.
> 2. **Dual-Placement Strategy:** Due to the known `gc.attempt=1` path override on active beads (which causes them to write output metadata to `attempt-1/` and blocks attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and `.gc/design-reviews/ga-dtvdnd/attempt-4/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 4 synthesis.

---

## Executive Summary

The Attempt 4 requirements and draft implementation plan for the Core and Gastown Pack Split represent a highly disciplined progression. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract** (AC7) successfully resolves the requirements schema conflict.

However, from the perspective of an **Asset Classification & Split Reviewer**, this document must be **BLOCKED**. Under the current AC6 and implementation plan specifications, the ledger's contract is structurally incapable of ensuring file-level safety and preventing behavioral drift. Specifically, the schema assumes a single-output destination per row, which physically cannot represent split assets that yield both a Core-neutral output and an external Gastown-specific output. Additionally, the lack of an explicit file-level basename collision scanner will result in silent overwrites of highly divergent duplicate fragments (such as `architecture.template.md` which exists under both `gastown` and `maintenance` roots).

---

## Top Strengths (Grounded Evidence)

1. **Clean Separation of Schema concerns (AC1 & AC6):** Purging the normative file-by-file markdown table from the requirements body resolves the requirements schema compliance issue while preserving a machine-validated external ledger as a hard gate.
2. **Explicit Block on Unresolved Reviews (AC6 / Implementation Plan):** Mandating that ledger validation fails on unresolved `review` rows is a major gate that prevents deferred, ambiguous ownership decisions from slipping into implementation.
3. **Structured Behavior Evidence Manifest (AC7 / Implementation Plan):** Requiring stable row IDs, source-kind enums, and both old and new witnesses (such as unit tests, process tests, or golden transcripts) ensures that any re-homed behavior is verifiably triggerable before code is merged.
4. **Git-Backed Historical Baseline (Implementation Plan):** Scanning against a Git historical baseline rather than only active workspace files ensures that deleted or moved legacy files cannot escape classification.

---

## Critical Risks & Gaps

### 1. [Blocker] Multi-Output Split Schema Defect (AC6 & AC7)
* **The Risk:** AC6 defines a ledger that records "target output path or retirement action, split boundary." The implementation plan's "Behavior Evidence Contract" defines "new path." Neither accommodates split assets that must yield both a Core-neutral output and a Gastown-specific output. If there is only a single output field, one half of the split will be left completely undefined or orphaned, violating the "fails on orphaned split behavior" assertion.
* **Mitigation Recommendation:** The Asset Migration Ledger (AC6) and Behavior Evidence manifest (AC7) schemas must explicitly require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the content assigned to each target for any file classified as `split` or `core-renamed`.

### 2. [Major] Basename Collisions and Silent Overwrites of Divergent Fragments
* **The Risk:** In our repository scan, template fragments like `architecture.template.md` exist under both `examples/gastown/packs/gastown/template-fragments/` and `examples/gastown/packs/maintenance/template-fragments/` with identical names but completely different contents (the former defines roles like Mayor/Deacon, while the latter defines low-level CLI contexts like `city.toml`/dogs). A naive file-presence validator will pass them, but a manual or automated move will result in silent overwrites unless an AST- or file-level "basename collision scanner" is explicitly mandated.
* **Mitigation Recommendation:** Require the ledger validation tool to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

### 3. [Major] Stale and Phantom Rows via One-Way Ledger Validation
* **The Risk:** AC6 fails on "missing current paths" and "unrepresented active source files" (ensuring every active file has a row). However, it does not fail on the reverse direction: a row whose `current path` no longer resolves to an active tracked file at the chosen snapshot. In previous iterations, static lists carried multiple phantom/deleted files. If the ledger is not validated bidirectionally, dead rows can drive implementation to create unnecessary directories and drift from the live process tree.
* **Mitigation Recommendation:** Make ledger validation strictly bidirectional against a named snapshot: every active tracked file under the roots must map to exactly one row, and every row's current path must resolve to a valid file in that snapshot unless marked historical/retired by schema.

### 4. [Minor] Missing Example Mapping for Splits
* **The Risk:** The Example Mapping table lacks a concrete row for a split file (such as `architecture.template.md` or `command-glossary.template.md`), leaving the exact verification evidence and behavior-preservation proof for splits completely hypothetical.
* **Mitigation Recommendation:** Add a concrete Example Mapping row demonstrating a real split asset (like `architecture.template.md`), illustrating its Core-neutral output, Gastown-specific output, validation evidence, and behavior-preservation proof.

---

## Required Changes & Actions

1. **Mandate Dual-Output Fields for Splits:**
   Amend AC6 and AC7 to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
2. **Implement Basename Collision Scanning:**
   Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.
3. **Require Bidirectional Git-Tracked Validation:**
   Tighten AC6 to specify that the ledger validation must be bidirectional against a deterministic git-backed baseline (using `git ls-files` under the legacy directories at a named snapshot), failing explicitly on stale/phantom rows.
4. **Add Closed Classification Vocabulary:**
   Add a closed target-owner vocabulary in AC6 (such as `core`, `core-renamed`, `gastown`, `split`, `retire`, `review`) with clear semantics defining when each is allowed.
5. **Add Example Mapping Split Row:**
   Add an Example Mapping row demonstrating a real split asset (like `architecture.template.md`), illustrating its Core-neutral output, Gastown-specific output, validation evidence, and behavior-preservation proof.

---

## Questions

- **Reconciliation of `architecture.template.md`:** Since the two copies are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
- **Closed Vocabulary Guardrails:** How will the ledger validator mechanically enforce that no implementer invents a custom target classification outside the six allowed values?
- **Stale Directory Cleanup:** How will the loader or doctor distinguish stale directories on disk from operator custom edits without risking accidental data loss?
