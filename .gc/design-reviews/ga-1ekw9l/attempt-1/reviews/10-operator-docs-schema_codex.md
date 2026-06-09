# Claire Dubois - Codex

**Verdict:** approve-with-risks

I did not inspect Claude's raw review output for this persona before writing this review.

**Top strengths:**
- The plan now names a single vocabulary authority, `plans/core-gastown-pack-migration/support/terminology-matrix.yaml`, and requires docs wording, generated references, CLI help, tutorial transcripts, doctor output, OpenAPI, dashboard types, and schema artifacts to regenerate before linting.
- It explicitly handles the hardest wording distinction for this migration: retired standalone Maintenance-pack references versus legitimate lowercase maintenance, Dolt/store-maintenance terms, Core-maintenance-worker bindings, public Gastown docs, historical examples, and operator recovery text.
- It makes `docs/reference/system-packs.md` a required artifact rather than optional docs debt, and it states canonical operator wording for Core, `bd`, `dolt`, retired Maintenance, explicit public Gastown, and stale generated paths.

**Critical risks:**
- [Major] The terminology matrix is specified structurally but not yet as an executable policy. The plan lists token, class, allowed/denied contexts, owner, examples, false-positive rule, and golden fixture, but it should also say how severity is assigned, how expired allowlist rows fail, how generated output freshness is tied to the matrix digest, and how whole-file or whole-directory suppressions are rejected. Without that, this can become a noisy wording lint that gets broadly suppressed.
- [Major] The operator recovery story is still scattered across loader, doctor, cache, runtime-state, and docs sections. The plan should require one canonical operator journey that starts from missing/corrupt Core, retired Maintenance/Gastown imports, stale generated directories, old public pins, offline cache misses, and version skew, then leads through diagnostics, `gc doctor --fix --non-interactive` when allowed, manual remediation when not allowed, and post-repair verification. Right now the wording is correct, but an upgrading operator could still have to assemble the path from several docs and golden outputs.
- [Major] Public Gastown docs and Gas City docs need an explicit cross-repo release boundary. The plan says the scanner distinguishes public Gastown docs and that public Gastown tests cover moved behavior, but it does not require a matching public-pack docs/golden update before Gas City consumes the activation pin. If public Gastown still tells users to look at in-tree examples or Maintenance-owned assets, Gas City's local docs can be correct while the actual user path remains stale.
- [Major] Generated references are named broadly, but the concrete output set is not pinned. Current operator-visible surfaces include docs/reference CLI output, init help text, rig help examples, prompt resolution help, doctor/import-state messages, tutorial transcripts, generated schema files, OpenAPI/dashboard types when API surfaces change, and examples. The plan should list the generator commands or freshness target that proves each surface was regenerated from current source.
- [Minor] `docs/reference/system-packs.md` does not exist in the current tree. The plan correctly says to create it, but the task decomposition should treat its absence as a first docs slice dependency, not as cleanup after behavior changes land.

**Missing evidence:**
- The exact `terminology-matrix.yaml` schema, including severity, expiry semantics, suppressions, path scopes, generated-output coverage, and negative fixtures.
- A `docs-authority-audit.yaml` row model that maps every legacy Maintenance/Gastown/current Core/public Gastown reference to allowed current docs, migration-history text, test fixture, generated output, or required rewrite.
- A single operator recovery document or golden transcript sequence covering missing Core, retired imports, stale generated paths, public pin/cache failures, version skew, offline operation, live-controller refusal, and rollback/downgrade guidance.
- The generated-reference command list and output paths under `plans/core-gastown-pack-migration/support/generated-references/`.
- Cross-repo proof that public `gascity-packs/gastown` docs and examples use the same terminology and do not point operators back to retired in-tree Gas City paths.

**Required changes:**
- Make the terminology matrix executable: define the schema, severity levels, fail conditions, expiry handling, no-whole-file-suppression rule, generated-output digest/freshness check, and fixtures that prove valid lowercase maintenance and store-maintenance references are accepted while retired Maintenance-pack authority references fail.
- Add a canonical operator recovery narrative to `docs/reference/system-packs.md` or a linked troubleshooting page, with golden text/JSON outputs for the key diagnostics and explicit stdout/stderr and exit-code expectations where commands are shown.
- Require `docs-authority-audit.yaml` to cover docs, examples, CLI help, doctor/import-state strings, generated references, tutorial transcripts, public Gastown docs, schema/OpenAPI/dashboard generated files when touched, script comments, prompts, and test fixtures. Each legacy reference should have an owner, current classification, rewrite target or historical exception, and expiry when exceptional.
- Pin the generator/freshness commands that create `support/generated-references/` and make docs lint consume those generated outputs rather than relying on manually collected examples.
- Add a cross-repo docs gate before activation-pin consumption: public Gastown docs/examples must agree with Gas City's canonical Core/public Gastown/retired Maintenance vocabulary and must not direct users to `examples/gastown/packs/*`, `.gc/system/packs/gastown`, or standalone Maintenance as active authority.

**Questions:**
- Will `docs/reference/system-packs.md` be the canonical upgrade/recovery page, or should troubleshooting own the workflow and system-packs own only the reference model?
- Which command regenerates tutorial transcripts and CLI help goldens, and where are those generated outputs checked in or recorded for the docs gate?
- How will the wording scanner distinguish public Gastown docs that legitimately describe Gastown roles from Gas City Core docs where those same role names are forbidden?
