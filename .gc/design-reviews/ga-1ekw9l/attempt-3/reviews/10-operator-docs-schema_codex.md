# Claire Dubois - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The docs scanner scope includes the right generated and operator-facing surfaces: Markdown, MDX, JSON, TXT, TS, OpenAPI, dashboard generated files, docs/schema outputs, public Gastown docs, generated help, CLI examples, scripts, prompts, tutorial transcripts, and doctor output (`design-before.md:327`-`design-before.md:330`).
- The plan gives a canonical vocabulary sentence that distinguishes required host Core, provider-dependent `bd`/`dolt`, retired standalone Maintenance, explicit public Gastown, and stale generated paths as ignored legacy state (`design-before.md:333`-`design-before.md:342`).
- Testing covers docs and release-facing proof with doctor output, CLI help, first-run text, tutorial transcripts, docs wording goldens, and `make dashboard-check` when API/dashboard/generated schema surfaces change (`design-before.md:441`-`design-before.md:446`).

**Critical risks:**
- [Major] The wording scanner needs an explicit terminology matrix, not only a canonical sentence. This migration must distinguish `Maintenance` as a retired pack name from legitimate lowercase maintenance, Dolt/store maintenance, supervisor cleanup, and runtime reliability language. Without a typed matrix, lint will either be noisy enough to suppress wholesale or too loose to catch stale operator guidance.
- [Major] The operator upgrade narrative is distributed across doctor, cache/pin, runtime-state, and docs sections, but the plan does not name one end-to-end runbook or tutorial flow for an upgrading city. Operators need a single path from stale paths or missing orders through doctor diagnostics, public pin/cache verification, repair refusal or mutation, and recovery/rollback guidance.
- [Minor] Generated schema/reference freshness is included in scanner scope, but the plan should require generated artifacts to be regenerated and checked in the same slice as any API or diagnostic shape change, not only when `make dashboard-check` is triggered.

**Missing evidence:**
- A terminology matrix with allowed and denied terms, contexts, examples, owners, and false-positive handling.
- A golden operator journey for missing Core, retired Maintenance import, stale `.gc/system/packs/maintenance`, and old public Gastown pin/version skew.
- A generated-reference freshness gate that fails on stale docs/schema/OpenAPI/dashboard references after diagnostic or config schema changes.
- Examples of JSON and text doctor/import-state diagnostics using the final vocabulary.

**Required changes:**
- Add an executable wording matrix that classifies retired pack references separately from legitimate maintenance terminology and public Gastown terminology.
- Add one upgrade runbook/tutorial path that starts from a legacy city symptom and ends with verified Core/public Gastown state or explicit manual recovery.
- Require doctor/import-state text and JSON goldens to use the same terms as docs/tutorials/CLI help.
- Tie generated docs/schema/OpenAPI/dashboard artifacts to the same slice as diagnostic/API shape changes, with a freshness check rather than optional docs debt.

**Questions:**
- Where will the terminology matrix live, and will it be consumed by both docs lint and generated-reference lint?
- What exact operator command sequence verifies the public Gastown pin/cache after doctor reports stale or retired paths?
- Which generated schema artifacts change if doctor/import-state JSON diagnostics add source/provenance fields?
