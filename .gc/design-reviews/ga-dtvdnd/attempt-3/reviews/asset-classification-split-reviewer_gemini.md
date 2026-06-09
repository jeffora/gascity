# Hugo Bautista — DeepSeek V4 Flash (Asset Classification & Split Review)

**Verdict:** block

**Scope:** Split and core-renamed ownership; file-by-file migration table completeness; clean halves; resolution of review-marked assets.

> ### Lane Note (Verify-Don't-Copy + Path Handling)
> 1. **Re-grounding & Independence:** This review is an independent DeepSeek V4 Flash evaluation of the Attempt 3 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-3/design-before.md` updated 2026-06-09T01:20:00Z). I have not inherited conclusions from prior rounds. I read AC6 and the Example Mapping, and verified the split/duplicate hazards against the live source tree to expose critical risks.
> 2. **Dual-Placement Strategy:** Due to the known `gc.attempt=1` path override on active beads (which causes them to write output metadata to `attempt-1/` and blocks attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and `.gc/design-reviews/ga-dtvdnd/attempt-3/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 3 synthesis.

---

## Executive Summary

The Attempt 3 requirements document for the Core and Gastown Pack Split represents a disciplined evolution. Moving the inline file-by-file markdown table from the requirements body into an external machine-validated **Asset Migration Ledger** (AC6) successfully resolves the requirements schema conflict. 

However, from the perspective of an **Asset Classification & Split Reviewer**, this document must be **BLOCKED**. Under the current AC6 specification, the ledger's contract is structurally incapable of ensuring file-level safety and preventing behavioral drift. Specifically, we have verified that critical template fragments like `architecture.template.md` exist in both `gastown` and `maintenance` packs with identical basenames but completely divergent content. The current single-output path schema and lack of bidirectional git-tracked validation will result in orphaned split behavior, phantom files, and silent overwrites during extraction.

---

## Top Strengths (Grounded Evidence)

1. **Resolution of Requirements Schema (AC1 & AC6):** Purging the inline file-by-file table and delegating it to an external ledger contract keeps the requirements high-level, satisfying requirements schema compliance while retaining strict, auditable file-granularity ownership as a hard gate.
2. **Explicit Block on Unresolved Reviews (AC6):** Mandating that ledger validation fails on unresolved `review` rows is a major gate that prevents deferred, ambiguous ownership decisions from slipping into the implementation phase.
3. **Structured Behavior-Preservation Mandate (AC7):** Requiring a behavior-preservation manifest that verifiably asserts that stripped Core behavior is re-homed and triggerable under Gastown provides the necessary guarantees before any code merges.

---

## Critical Risks & Gaps

### 1. [Blocker] Multi-Output Split Schema Defect (AC6 Split Gaps)
* **The Risk:** AC6 defines a ledger schema with a single "target output path or retirement action" plus a prose "split boundary." However, our analysis of the active tree reveals that same-named, dual-pack assets must split (such as `architecture.template.md` which exists in both `examples/gastown/packs/gastown/template-fragments/` and `examples/gastown/packs/maintenance/template-fragments/` with completely different content). A single output path field physically cannot record the Core output *and* the Gastown output *and* which content lands on each side. One half of the split will be left completely undefined or orphaned, violating the "fails on orphaned split behavior" assertion.
* **Mitigation Recommendation:** The Asset Migration Ledger schema must explicitly require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the content assigned to each target for any file classified as `split` or `core-renamed`.

### 2. [Major] Stale and Phantom Rows via One-Way Ledger Validation
* **The Risk:** AC6 fails on "missing current paths" and "unrepresented active source files" (ensuring every active file has a row). However, it does not fail on the reverse direction: a row whose `current path` no longer resolves to an active tracked file at the chosen snapshot. In previous iterations, static lists carried multiple phantom/deleted files. If the ledger is not validated bidirectionally, dead rows can drive implementation to create unnecessary directories and drift from the live process tree.
* **Mitigation Recommendation:** Make ledger validation strictly bidirectional against a named snapshot: every active tracked file under the roots must map to exactly one row, and every row's current path must resolve to a valid file in that snapshot.

### 3. [Major] Basename Collisions and Silent Overwrites of Divergent Fragments
* **The Risk:** In our repository scan, we ran a direct diff on `architecture.template.md` between `gastown` and `maintenance` packs. They are vastly different (the former defines the high-level actor roles like Mayor/Deacon, while the latter defines the low-level CLI context like `city.toml`/dogs/beads). A naive ledger validator verifying that both paths exist will successfully pass, but during extraction, the basename collision will result in one silently overwriting the other in the output directory if a collision-scanning check is absent.
* **Mitigation Recommendation:** Require the ledger validation tool to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

---

## Required Changes & Actions

1. **Mandate Dual-Output Fields for Splits:**
   Amend AC6 to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
2. **Require Bidirectional Git-Tracked Validation:**
   Tighten AC6 to specify that the ledger validation must be bidirectional against a deterministic git-backed baseline (using `git ls-files` under the legacy directories at a named snapshot), failing explicitly on stale/phantom rows.
3. **Add Closed Classification Vocabulary:**
   Add a closed target-owner vocabulary in AC6 (such as `core`, `core-renamed`, `gastown`, `split`, `retire`, `review`) with clear semantics defining when each is allowed.
4. **Implement Basename Collision Scanning:**
   Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.
5. **Add Example Mapping Split Row:**
   Add an Example Mapping row demonstrating a real split asset (like `architecture.template.md`), illustrating its Core-neutral output, Gastown-specific output, validation evidence, and behavior-preservation proof.

---

## Questions

- **Reconciliation of `architecture.template.md`:** Since the two copies are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
- **Closed Vocabulary Guardrails:** How will the ledger validator mechanically enforce that no implementer invents a custom target classification outside the six allowed values?
