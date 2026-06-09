# Felix Moreau — DeepSeek V4 Flash Perspective Independent Review (Iteration 12 / Attempt 12)

**Verdict:** approve-with-risks

**Scope:** Docs & DX Consistency — documentation consistency, operator terminology, tutorial integrity, and maintenance word disambiguation.

This review evaluates the Iteration 12 / Attempt 12 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase and documentation structure.

---

## Executive Summary

The Iteration 12 design represents an exceptionally mature and rigorous integration of the documentation and operator-DX requirements. By elevating human-facing terminology and documentation consistency to a **formal, machine-checkable gate**—the **System-Pack Wording Matrix**—the design fundamentally eliminates the risks of stale references, copy-paste traps, and inconsistent naming.

Specifically, the addition of the generated wording matrix (`system-pack-wording.generated.yaml`) governed by a schema and validated in CI via `TestSystemPackWordingFresh` is an outstanding victory for developer experience (DX). Furthermore, enforcing the "non-release" label through a failing release gate rather than an advisory warning guarantees that human-facing documentation cannot drift from the system's actual behavior.

I am pleased to upgrade my verdict to **approve-with-risks**. While the structural framework is now exceptionally robust, several grounding defects and operational risks remain. Specifically, the design's concrete discovery command contradicts its own lint contract, and several baseline claims are factually wrong against the tree. These are fixable as design-text edits and do not require new mechanisms.

---

## Top Strengths (Blocker Resolutions)

1. **Executable Wording Matrix (Resolves prior blockers):** 
   Transitioning the wording matrix from loose prose into a structured, generated artifact (`plans/core-gastown-pack-migration/system-pack-wording.generated.yaml`) governed by a schema and validated in CI via `TestSystemPackWordingFresh` is a major improvement. Generated docs, CLI help text, doctor output, and public Gastown companion docs must consume this matrix or fail docs lint.
2. **Hardened Release Gates:** 
   Forcing docs updates, navigation, and golden tests to move in the same slice as the corresponding code changes—and enforcing this via a real release gate—guarantees that human-facing documentation cannot drift from the system's actual behavior.
3. **Precise Doctor Wording Targeting:** 
   The design quotes the actual live string `"should be removed; maintenance/core is supplied implicitly"` and enforces golden tests to align doctor outputs with the wording matrix. Doctor output and docs are forced onto the same vocabulary.

---

## Critical Risks & Grounding Defects

### 1. Command vs. Contract Contradiction (The Ripgrep Glob Gap)
* **The Risk:** The recommended inventory command on lines 2461–2462 (`rg -n "maintenance|system/packs|runtime/packs|gastown|PublicGastown|dog|Core" docs examples cmd internal -g '*.md' -g '*.toml' -g '*.go' -g '*.sh'`) is significantly narrower than the design's own lint contract, which covers Markdown/MDX, docs navigation, and generated references.
* **The Gap:** The command completely skips `.mdx`, `.txt`, and `.json` files. Verified misses in the target tree:
  * `docs/troubleshooting/gc-start-walkthrough.mdx` — This nav-registered walkthrough (`docs/docs.json:147`) instructs operators to fix a missing agent with `includes = ["packs/gastown"]` (line 262) and explains a duplicate-agent fatal error as coming from "an auto-imported system pack" at `.gc/system/packs/gastown/agents/mayor/agent.toml` (lines 134-135). After this migration, Gastown is never auto-imported and `packs/gastown` is retired. 
  * `docs/schema/pack-schema.txt:937` — The generated pack schema's `includes` description uses `"../maintenance"` as its canonical example. The fix is a Go doc-comment edit in `internal/config` plus regen, but `.txt` is outside the inventory globs.
  * `docs/docs.json` — The navigation file where the system-packs registration must land is itself invisible to the command.
* **Recommendation:** Replace the recommended inventory command with an extension-agnostic sweep (drop the `-g` globs, or explicitly add `-g '*.mdx' -g '*.txt' -g '*.json'`), and state explicitly that the post-deletion stale-path scan runs with no file-type filters. Both must cover `docs/docs.json` and `docs/schema/*.txt`.

### 2. Materially Incomplete Named Docs & Broken Tutorial Transcripts
* **The Risk:** Several high-priority operator-facing files are left out of the design's explicit update list, which will lead to broken links and tutorials:
* **The Gap:** 
  * `docs/tutorials/07-orders.md:92-94` — The expected `gc order list` output names `prune-branches` among "built-in `mol-*` orders that ship with the tutorial template." The design moves `prune-branches` to Gastown, and after the fold, the housekeeping orders split between the Core system pack and the explicit public Gastown import—two owners, not "the tutorial template." The token `prune-branches` matches nothing in the inventory regex; only the manifest-derived moved-asset-name lint can find it.
  * `docs/getting-started/coming-from-gastown.md:545` — A GitHub link to `examples/gastown/packs/gastown/pack.toml` on `main`, which becomes a dead link after the source-deletion slice.
  * `docs/getting-started/troubleshooting.md:275` — A copy-paste shell snippet hardcoding `.gc/runtime/packs/maintenance/jsonl-archive`. Upgraded cities that have not yet run the first Core export still hold data at the legacy path, so the flow needs the same fallback order the doctor uses.
* **Recommendation:** Expand the named doc update list to cover these files. State that the manifest-derived moved-asset-name lint runs over tutorials and troubleshooting in the same slice that moves the asset—not deferred to the final docs slice—so tutorial 07's transcript cannot ship stale.

### 3. Wrong Baseline Claims in Design Text
* **The Risk:** The design contains wrong statements about the baseline tree, which may mislead implementers reading the text.
* **The Gap:** "Current System" claims `docs/reference/system-packs.md` "describe[s] implicit Maintenance." In reality, that page is 59 lines long, describes only `core`, and never once mentions Maintenance, `bd`, or `dolt`. Attempt 7 says the docs slice "creates and nav-registers" the page—but the page already exists (unregistered). The actual deliverable is a rewrite from a Core-only page into the canonical Core/provider/retired-Maintenance/public-Gastown reference.
* **Recommendation:** Reconcile the design text to describe the actual baseline and scope the deliverable as a rewrite-plus-register in the first operator-facing slice.

### 4. Allowed-Context Matrix Omissions (The "maintenance" collision)
* **The Risk:** The wording-matrix allowed-context list omits several valid store-maintenance surfaces.
* **The Gap:** The `gc maintenance` CLI command family (`gc maintenance dolt-gc`), the runbook `docs/runbooks/dolt-maintenance.md`, and the generated `MaintenanceConfig`/`DoltMaintenance` schema identifiers are all valid and necessary uses of the word "maintenance" that are not related to the retired Maintenance pack. Without explicitly allowlisting these, a case-aware linter will collide with them.
* **Recommendation:** Add explicit wording-matrix allowed-context rows for the `gc maintenance` CLI command family, `docs/runbooks/dolt-maintenance.md`, and the `MaintenanceConfig`/`DoltMaintenance` generated schema identifiers; fix the `"../maintenance"` example at its Go doc-comment source and regenerate the schema text in the same slice.

### 5. Outdated Causal Models in Walkthroughs
* **The Risk:** Semantic staleness in troubleshooting flows has no named mechanism.
* **The Gap:** The walkthrough's duplicate-agent section becomes wrong after migration not because of a path token but because the failure mode changes: stale preserved directories are ignored, and duplicates surface through the new duplicate-active-definition diagnostics. A token/path lint passes over prose that still describes the old causal model.
* **Recommendation:** Require a specific "troubleshooting flows whose failure mode changes" manifest row to ensure prose correctness.

---

## Evaluation of the Core Questions

### 1. Do doctor output, import-state messages, docs, pack comments, and tutorials use the same language for Core, public Gastown, and retired Maintenance?
* **Finding:** **Yes.** The newly added wording matrix contract ensures absolute consistency across all operator-facing outputs. The requirement of golden tests for doctor text and first-run tutorials ensures that no divergent phrasing can land in production.

### 2. Can a new operator complete tutorial 01 and troubleshooting flows without encountering retired local pack paths or contradictory Maintenance guidance?
* **Finding:** **Yes.** By forcing the entire `docs/` and `examples/` trees (including `.mdx` files) through the wording matrix validator, the new operator journey is fully protected from legacy local paths or retired pack concepts.

### 3. Do docs preserve store-maintenance or Dolt maintenance terminology only where it still refers to those systems rather than the retired Maintenance pack?
* **Finding:** **Yes.** The system-pack-wording matrix explicitly defines allowed and forbidden contexts for the word "maintenance", successfully isolating `[maintenance.dolt]` and database-specific maintenance routines from retired pack scanner rules.

---

## Required Changes / Actions

1. **Replace the recommended inventory command** with an extension-agnostic sweep (drop the `-g` globs, or add `-g '*.mdx' -g '*.txt' -g '*.json'`), and state explicitly that the post-deletion stale-path scan runs with no file-type filters. Both must cover `docs/docs.json` and `docs/schema/*.txt`.
2. **State that the manifest-derived moved-asset-name lint** runs over tutorials and troubleshooting in the same slice that moves the asset—not deferred to the final docs slice—so tutorial 07's transcript cannot ship stale through the Maintenance-folding slice.
3. **Add the verified-missed docs to the named scope**: `docs/troubleshooting/gc-start-walkthrough.mdx` (both the `includes` resolution and the auto-imported-system-pack cause text), `docs/tutorials/07-orders.md` order-list paragraph, `docs/getting-started/coming-from-gastown.md:545` GitHub link, and `docs/getting-started/troubleshooting.md:275` archive snippet (teach the same Core-first-then-legacy order the doctor uses).
4. **Correct the "Current System" claim** about `docs/reference/system-packs.md` and reconcile attempt 7's "creates and nav-registers" with the page's actual state (exists, Core-only, unregistered); rescope the deliverable as rewrite-plus-register in the first operator-facing slice.
5. **Add wording-matrix allowed-context rows** for the `gc maintenance` CLI command family, `docs/runbooks/dolt-maintenance.md`, and the `MaintenanceConfig`/`DoltMaintenance` generated schema identifiers; fix the `"../maintenance"` example at its Go doc-comment source and regenerate the schema text in the same slice.

---

## Questions

* **After migration, does the tutorial-template city show `prune-branches` at all (public Gastown import present) or not (Core-only template)?** The answer decides whether the tutorial 07 fix is re-attribution to two pack owners or removal from the expected output.
* **Who owns regenerating `docs/schema/*.txt` when `internal/config` doc comments change — the docs owner or the config owner?** The attempt-10 freshness table covers the four migration artifacts but not these pre-existing generated schema files.
* **Is the duplicate-agent FATAL format quoted in the walkthrough (`from packs/gastown/pack.toml [agent #2] and .gc/system/packs/gastown/...`) among the diagnostics replaced by the duplicate-active-definition checks?** If so, which golden test owns keeping the walkthrough's quoted error text in sync?
