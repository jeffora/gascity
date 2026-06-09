# Design Review Synthesis

## Overall Verdict: block

The required current-attempt persona syntheses are missing, so this synthesis
cannot produce a trustworthy design verdict from the mandated inputs. This is a
workflow-artifact block, not a conclusion that another edit to
`plans/core-gastown-pack-migration/implementation-plan.md` would resolve the
review.

## Consensus Strengths
- No current-attempt persona-synthesis consensus can be established because
  `.gc/design-reviews/ga-1ekw9l/attempt-4/persona-syntheses/` contains no files.
- The workflow did prepare ten persona definitions and closed ten persona
  synthesis beads with `gc.outcome=pass`, which provides enough metadata to
  identify the artifact-routing defect.

## Critical Findings

### [Blocker] Required current-attempt persona syntheses are absent
**Sources:** `.gc/design-reviews/ga-1ekw9l/personas/final_personas.tsv`, `.gc/design-reviews/ga-1ekw9l/attempt-4/persona-syntheses/`, `ga-twtldr`, `ga-yjh4m3`, `ga-1hfea3`, `ga-lv7o23`, `ga-7y51cx`, `ga-o83wq2`, `ga-d0yloa`, `ga-lrbbes`, `ga-m0rabx`, `ga-tzp6vv`, `ga-86h7y5`
**Actionability:** workflow-defect
**Issue:** The current attempt requires persona syntheses for ten slugs:
`01-required-core-loading`, `02-behavior-evidence-chain`,
`03-doctor-mutation-safety`, `04-role-neutrality-zfc`,
`05-public-pack-pin-cache`, `06-pack-boundary-containment`,
`07-bootstrap-fixture-isolation`, `08-runtime-state-migration`,
`09-rollout-decomposition-gates`, and `10-operator-docs-schema`. The expected
directory for Attempt 4 exists, but contains zero persona synthesis files.
**Required change:** Repair or rerun the persona synthesis stage so all ten
required syntheses are written under
`.gc/design-reviews/ga-1ekw9l/attempt-4/persona-syntheses/` before global
synthesis runs.

### [Major] Persona synthesis metadata points to the wrong attempt directory
**Sources:** `ga-yjh4m3`, `ga-1hfea3`, `ga-lv7o23`, `ga-7y51cx`, `ga-o83wq2`, `ga-d0yloa`, `ga-lrbbes`, `ga-m0rabx`, `ga-tzp6vv`, `ga-86h7y5`
**Actionability:** workflow-defect
**Issue:** Each closed persona synthesis bead has `gc.attempt=4` and
`gc.outcome=pass`, but its `design_review.output_path` points at
`.gc/design-reviews/ga-1ekw9l/attempt-1/persona-syntheses/<slug>.md`. That
wrong-attempt path is not safe to use as an Attempt 4 input; at least
`01-required-core-loading.md` has stale content and a modification time from
before Attempt 4 was created.
**Required change:** Fix the persona synthesis output-path calculation and add
an artifact consistency check that fails if `gc.attempt` and
`design_review.output_path` disagree.

### [Major] Current-attempt raw review artifacts are incomplete for required synthesis inputs
**Sources:** `.gc/design-reviews/ga-1ekw9l/attempt-4/reviews/`,
`.gc/design-reviews/ga-1ekw9l/attempt-4/output.json`, persona synthesis bead
descriptions
**Actionability:** workflow-defect
**Issue:** Attempt 4 contains only
`01-required-core-loading_gemini.md` and
`06-pack-boundary-containment_gemini.md` under `reviews/`, while the persona
synthesis contract says Claude and Codex reviews are required for each persona
and the third model may be absent only when skipped or soft-failed. The ten
persona synthesis beads nevertheless closed as passing.
**Required change:** Investigate why required Claude/Codex raw review artifacts
are absent from the current attempt directory, and ensure the persona synthesis
stage refuses `pass` when required raw inputs are missing.

## Disagreements
- No current-attempt persona-level disagreements can be assessed because the
  required current-attempt persona syntheses are absent.
- Some files under `attempt-1/persona-syntheses/` appear to have been modified
  during Attempt 4, while at least one is stale from before Attempt 4. My
  assessment is that mixing those files into this global synthesis would hide a
  workflow defect and produce an unreliable verdict.

## Missing Evidence
- All ten required current-attempt persona synthesis files.
- Current-attempt persona-level verdicts, consensus findings, disagreements,
  missing evidence, and required changes.
- Schema-conformance assessment from the persona syntheses. The schema file is
  non-empty, but this global synthesis cannot separate true design risks from
  output-shape mismatches without the required current-attempt syntheses.
- Required Claude and Codex raw reviews for each Attempt 4 persona.

## Convergence Assessment
- Remaining blocker class: workflow-defect
- Recommended apply verdict: blocked
- Reason: Another edit to `plans/core-gastown-pack-migration/implementation-plan.md`
  cannot make the next review pass because the blocker is missing and
  misrouted workflow artifacts, not document content.
- Next non-design work: fix the design-review persona synthesis artifact path
  and required-input validation, rerun the persona synthesis stage for Attempt
  4, then rerun global synthesis.

## Recommended Changes
1. Fix persona synthesis output-path resolution so Attempt 4 writes to
   `.gc/design-reviews/ga-1ekw9l/attempt-4/persona-syntheses/`, not
   `attempt-1/persona-syntheses/`.
2. Add a global-synthesis preflight check that compares the expected persona
   slugs from `output.json` or `final_personas.tsv` to files under the current
   attempt's `persona-syntheses/` directory and fails closed on missing,
   stale, or wrong-attempt paths.
3. Add persona-synthesis validation that refuses `gc.outcome=pass` unless the
   required Claude and Codex raw review files for that slug exist in the current
   attempt's `reviews/` directory, with only the third model allowed to be
   absent.
4. Rerun or repair Attempt 4 persona synthesis after the workflow fix, then
   rerun global synthesis from the corrected current-attempt inputs.
