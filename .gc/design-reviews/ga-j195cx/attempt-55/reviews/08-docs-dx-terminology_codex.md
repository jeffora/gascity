# Avery Brooks - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now gives the terminology problem a real contract: `requires`, host capability, compiler implementation, formula `version`, pack revision, `contract`, and `schema` are defined in one glossary instead of left to inference.
- The docs rollout is phase-blocking rather than advisory. User-visible diagnostics are tied to reference docs, generated help/schema/API artifacts, examples, PackV2 docs, and dashboard types.
- The open questions from this lane are mostly resolved: exact requirement strings are byte-exact, warning transport is bounded by surface, and legacy `version` is preserved as metadata rather than used as a migration or compiler selector.

**Critical risks:**
- [Major] The proposed stale-guidance matcher is too broad to be a durable CI gate. `docs/release/formula-compiler-stale-guidance.yaml` matches the literal pattern `version` across README, docs, engdocs, tutorials, examples, PackV2 docs, and TOML fixtures. The current tree legitimately uses "version" for Go/Node/runtime versions, pack import versions, release versions, CLI prose, and unrelated architecture docs. Unless this is narrowed to formula snippets, parsed TOML keys, or line-context-aware matches, the gate will either block unrelated docs work or force awkward boilerplate around every normal use of the word.
- [Major] The docs gates name several required artifacts, but the design does not yet pin a single operator-facing command and schema for the formula docs doctest and stale-guidance report. It says the implementation owns the CI command, while also making the command release-blocking. That leaves authors without a local "run this, fix these rows" loop comparable to `make dashboard-check`.
- [Minor] The "copy-paste-safe" docs contract conflicts with placeholder snippets such as `requires_gc = ">=<minimum-floor-from-formula-compiler-min-floor.json>"` and table cells like `>=<floor>`. The doctest fixture can classify them as templates, but the reference page still needs a clear separation between runnable TOML and release-template TOML so pack authors do not paste invalid placeholders.

**Missing evidence:**
- A sample stale-guidance report against the current repository showing that the matcher set produces actionable findings rather than broad false positives.
- A concrete `docs/reference/testdata/formula-requirements-doctest.yaml` example with valid, invalid_expected, and template_expected rows, plus the exact local command that updates or verifies it.
- Generated help or CLI reference examples for the new `gc formula validate --provenance`, `--legacy-contract-report`, and alias-removal/reporting flags. The current reference docs only expose `gc formula show`.
- A before/after sketch for `docs/reference/formula.md` proving that readers can distinguish formula author, pack author, pack consumer, and city operator responsibilities from that page alone.

**Required changes:**
- Replace the broad `pattern: 'version'` stale-guidance rule with a structured or scoped check. Prefer parsing TOML blocks and formula files for `version =` in formula context, then use prose matchers only for known formula-reference files.
- Add a named local quality gate for the docs bundle, including the doctest fixture, stale-guidance scan, first-party inventory validation, and generated help/schema checks. Document the command next to the generated source ownership table.
- Mark template snippets visibly in the reference docs and require at least one runnable pack/formula example with a concrete `requires_gc` floor once the minimum-floor artifact exists.
- Seed the first docs PR with an inventory of currently stale live examples, including `docs/reference/formula.md`, `docs/reference/config.md`, generated CLI help, tutorials, examples, and first-party pack snippets, so the migration cannot pass with only aspirational gates.

**Questions:**
- Should PackV2 and general installation docs be excluded from the broad stale-guidance prose matcher and covered only by specific pack-compatibility checks?
- Is `docs/reference/testdata/formula-requirements-doctest.yaml` hand-authored, generated, or both? The answer affects how reviewers should evaluate drift.
- Does the release captain own the stale-guidance matcher config long-term, or does each subsystem own its file-family exceptions?
