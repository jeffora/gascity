# Felix Moreau — DeepSeek V4 Flash (Docs & DX Consistency Reviewer, Iteration 16)

**Verdict:** approve-with-risks

**Lane:** Docs & DX Consistency — documentation consistency, operator terminology, tutorial integrity, and maintenance word disambiguation.

Reviewed against the revised design document in Iteration 16 (`.gc/design-reviews/ga-2404qu/attempt-16/design-before.md` updated 2026-06-09T02:00:53Z) and grounded in the live codebase and documentation tree under `docs/` and `examples/`.

---

## Executive Summary

The Iteration 16 design represents an exceptionally mature, systems-aware iteration of the documentation and operator-DX requirements. By elevating human-facing terminology and documentation consistency to a **formal, machine-checkable gate**—the **System-Pack Wording Matrix**—the design fundamentally eliminates the risks of stale references, copy-paste traps, and inconsistent naming.

Specifically, the addition of the generated wording matrix (`system-pack-wording.generated.yaml`) governed by a schema and validated in CI via `TestSystemPackWordingFresh` is an outstanding victory for developer experience (DX). Furthermore, enforcing the "non-release" label through a failing release gate rather than an advisory warning guarantees that human-facing documentation cannot drift from the system's actual behavior.

I am pleased to upgrade my verdict to **Verdict: approve-with-risks**. While the structural framework is now exceptionally robust, several grounding defects and operational risks remain in the design text. Specifically, several baseline claims are factually wrong against the tree, and the concrete discovery command contradicts the design's own lint contract. These are fixable as design-text edits and do not require new mechanisms.

---

## Top Strengths & Design Evolution

1. **Executable Wording Matrix (Resolves prior blockers):** 
   Transitioning the wording matrix from loose prose into a structured, generated artifact (`plans/core-gastown-pack-migration/system-pack-wording.generated.yaml`) governed by a schema and validated in CI via `TestSystemPackWordingFresh` is a major improvement. Generated docs, CLI help text, doctor output, and public Gastown companion docs must consume this matrix or fail docs lint.
2. **Hardened Release Gates:** 
   Forcing docs updates, navigation, and golden tests to move in the same slice as the corresponding code changes—and enforcing this via a real release gate—guarantees that human-facing documentation cannot drift from the system's actual behavior.
3. **Precise Doctor Wording Targeting:** 
   The design quotes the actual live string `"should be removed; maintenance/core is supplied implicitly"` and enforces golden tests to align doctor outputs with the wording matrix. Doctor output and docs are forced onto the same vocabulary.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Do doctor output, import-state messages, docs, pack comments, and tutorials use the same language for Core, public Gastown, and retired Maintenance?
**Yes, structurally.** Under the design, the system-pack wording matrix (`system-pack-wording.generated.yaml`) enforces unified vocabulary for Core, public Gastown, and retired Maintenance. CI checks (`TestSystemPackWordingFresh` and bidirectional linter checks) ensure no developer can diverge in help messages, CLI output, or comments.

### Q2: Can a new operator complete tutorial 01 and troubleshooting flows without encountering retired local pack paths or contradictory Maintenance guidance?
**Yes, but only if we resolve the critical grounding gaps identified in this review.** Structurally, the design mandates scanning the entire `docs/` and `examples/` trees against the wording matrix. However, the recommended ripgrep command suffers from a major glob gap, and several tutorials and troubleshooting files contain undetected, unmigrated legacy strings and paths. Once these specific gaps are resolved, the operator's journey will be perfectly preserved.

### Q3: Do docs preserve store-maintenance or Dolt maintenance terminology only where it still refers to those systems rather than the retired Maintenance pack?
**Yes.** The wording matrix explicitly allowlists valid contexts (like `gc maintenance`, `dolt-gc`, `DoltMaintenance`, and specific `internal/config` database maintenance structures) while strictly forbidding references to the legacy Maintenance pack.

---

## Critical Risks & Grounding Defects

An audit of the live `docs/` workspace against the Iteration 16 design document reveals that several critical grounding defects have survived into the current design text. These must be addressed to ensure operator-facing documentation is 100% correct.

### 1. [Major] The Missing Reference Page (`docs/reference/system-packs.md`)
* **The Contradiction**: The design text claims that `docs/reference/system-packs.md` is a pre-existing page that describes Core/Maintenance and only needs a "rewrite". 
* **The Grounding Gap**: This page is completely missing from the baseline tree. In reality, the files in `docs/reference/` are:
  - `docs/reference/api.md`
  - `docs/reference/cli.md`
  - `docs/reference/config.md`
  - `docs/reference/events.md`
  - `docs/reference/exec-beads-provider.md`
  - `docs/reference/exec-session-provider.md`
  - `docs/reference/formula.md`
  - `docs/reference/index.md`
  - `docs/reference/trust-boundaries.md`
* **Required Change**: Reconcile the design text. State explicitly that `docs/reference/system-packs.md` must be **created and nav-registered as a new file** in the first operator-facing slice.

### 2. [Major] Command vs. Contract Contradiction (The Ripgrep Glob Gap)
* **The Risk**: The recommended inventory command on line 2927:
  ```bash
  rg -n "maintenance|system/packs|runtime/packs|gastown|PublicGastown|dog|Core" docs examples cmd internal -g '*.md' -g '*.toml' -g '*.go' -g '*.sh'
  ```
  is significantly narrower than the design's own lint contract, which covers Markdown/MDX, docs navigation, and generated references.
* **The Grounding Gap**: The command completely skips `.mdx`, `.txt`, and `.json` files. This causes several verified misses in the target tree:
  - `docs/troubleshooting/gc-start-walkthrough.mdx` — This nav-registered walkthrough (`docs/docs.json:147`) instructs operators to fix a missing agent with `includes = ["packs/gastown"]` (line 262) and explains a duplicate-agent fatal error as coming from "an auto-imported system pack" at `.gc/system/packs/gastown/agents/mayor/agent.toml` (lines 134-135). After this migration, Gastown is never auto-imported and `packs/gastown` is retired. 
  - `docs/schema/pack-schema.txt:872` — The generated pack schema's `includes` description uses `"../maintenance"` as its canonical example. The fix is a Go doc-comment edit in `internal/config` plus regen, but `.txt` is outside the inventory globs.
  - `docs/docs.json` — The navigation file where the system-packs registration must land is itself invisible to the command.
* **Required Change**: Replace the recommended inventory command with an extension-agnostic sweep (drop the `-g` globs, or explicitly add `-g '*.mdx' -g '*.txt' -g '*.json'`), and state explicitly that the post-deletion stale-path scan runs with no file-type filters. Both must cover `docs/docs.json` and `docs/schema/*.txt`.

### 3. [Minor] Dead GitHub Links in Documentation
* **The Grounding Gap**: `docs/getting-started/coming-from-gastown.md:545` links directly to a GitHub URL pointing to `examples/gastown/packs/gastown/pack.toml` on the `main` branch. Since `examples/gastown` will be deleted during this migration, this link will become dead.
* **Required Change**: Reconcile this link to point to the new remote repository location in `gascity-packs` or replace it with a valid relative path/reference.

### 4. [Minor] Outdated Maintenance Snippets in Troubleshooting Runbooks
* **The Grounding Gap**: `docs/getting-started/troubleshooting.md:275` provides a shell snippet:
  ```bash
  ARCHIVE=$(gc config get state_dir)/packs/maintenance/jsonl-archive
  ```
  Since the Maintenance pack is retired, this path is obsolete and state is now migrated to `.gc/runtime/packs/core/...`.
* **Required Change**: Update the troubleshooting runbook to instruct the operator to check the active `core` directory or teach the same Core-first-then-legacy fallback order that the doctor uses.

### 5. [Minor] Inaccuracies in Tutorial 01 (`01-cities-and-rigs.md`)
* **The Grounding Gap**: `docs/tutorials/01-cities-and-rigs.md:189-191` states:
  > *"The `dog` pool is a background utility agent from the built-in maintenance pack. It handles internal housekeeping like shutdown coordination."*
  Because the Maintenance pack is retired, this is structurally inaccurate.
* **Required Change**: Update the tutorial text to correctly explain that the `dog` pool is a utility agent of the Core system pack.

### 6. [Minor] Tutorial 07 Transcript Desync (`07-orders.md`)
* **The Grounding Gap**: `docs/tutorials/07-orders.md:96` lists `prune-branches` as a built-in housekeeping order:
  > *"Your output will also include a handful of built-in `mol-*` orders that ship with the tutorial template (`beads-health`, `gate-sweep`, `mol-dog-jsonl`, `mol-dog-reaper`, `orphan-sweep`, `prune-branches`, `spawn-storm-detect`, `wisp-compact`, etc.)."*
  Since `prune-branches` is now Gastown-only (moved to the public Gastown repository), it will not show up in a Core-only tutorial template city.
* **Required Change**: Reconcile this list to omit `prune-branches` from the built-in housekeeping list of the tutorial template, or explicitly explain the split of housekeeping orders between the Core system pack and the public Gastown import.

### 7. [Minor] Outdated Causal Models in Walkthroughs (`gc-start-walkthrough.mdx`)
* **The Grounding Gap**: The walkthrough's duplicate-agent section (`docs/troubleshooting/gc-start-walkthrough.mdx:134-135`) quotes a duplicate name error of the format:
  ```text
  FATAL: agent 'mayor': duplicate name
         (from packs/gastown/pack.toml [agent #2]
          and .gc/system/packs/gastown/agents/mayor/agent.toml)
  ```
  After migration, duplicates will be detected and surfaced via the new duplicate-active-definition diagnostics, and `.gc/system/packs/gastown` will no longer exist.
* **Required Change**: Update the troubleshooting text to quote the new duplicate-active-definition diagnostics instead of the old folder-conflict error.

### 8. [Minor] Outdated Schema Examples in `pack-schema.txt`
* **The Grounding Gap**: `docs/schema/pack-schema.txt:872` uses `"../maintenance"` as its canonical includes example.
* **Required Change**: Correct this Go doc-comment in `internal/config` (from which `pack-schema.txt` is generated) and regenerate the schema file in the same slice.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 16 Design | Status |
| :--- | :--- | :--- |
| **Doctor message wording diverges from documentation wording** | **Excellent.** Solved by introducing the generated wording matrix (`system-pack-wording.generated.yaml`) and verifying its freshness in CI via `TestSystemPackWordingFresh`. | **Pass** |
| **Examples or tutorials reference half-migrated local Gastown or Maintenance paths** | **Pass with Caveat.** The wording matrix and CI linter check for local paths, but the recommended ripgrep command misses `.mdx`, `.txt`, and `.json` files. | **Risk (Needs Action)** |
| **Operator guidance sends missing-order debugging to retired pack locations** | **Pass with Caveat.** The design moves `prune-branches` to Gastown, but `07-orders.md` still describes it as part of the built-in tutorial template. | **Risk (Needs Action)** |

---

## Actionable Requirements & Proposed Adjustments

1. **Reconcile `docs/reference/system-packs.md` Creation**: Explicitly state that `docs/reference/system-packs.md` is a new file that must be created and registered in docs navigation in the first slice.
2. **Close the Ripgrep Glob Gap**: Update the recommended inventory command to use extension-agnostic sweeps (drop the `-g` file filters or add `.mdx`, `.txt`, and `.json` files) to ensure no documentation files are missed during audits.
3. **Fix Dead Links and Snippets**: 
   - Re-route the GitHub link in `docs/getting-started/coming-from-gastown.md:545`.
   - Update `docs/getting-started/troubleshooting.md:275` to point to the active Core directory.
4. **Update Tutorial and Walkthrough Prose**:
   - Update `01-cities-and-rigs.md:189-191` to map `dog` to Core.
   - Remove `prune-branches` from the Core-only housekeeping list in `07-orders.md:96`.
   - Re-align `gc-start-walkthrough.mdx:134-135` and line 262 to avoid legacy paths and reflect duplicate-active-definition diagnostics.
5. **Update Config Doc Comment**: Correct the `"../maintenance"` example in `internal/config` and regenerate `docs/schema/pack-schema.txt` to align with the retirement of the Maintenance pack.
