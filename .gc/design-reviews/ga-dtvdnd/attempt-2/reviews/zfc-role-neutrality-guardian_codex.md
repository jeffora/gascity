# Alistair Sterling - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The requirements now make the central ZFC contract explicit: Core must stay role-neutral, Gastown behavior must be explicit public pack configuration, and SDK infrastructure must remain self-sufficient with only the controller plus required Core support.
- `dog` is framed as configurable Core pack data rather than a Go-side exception, with renamed, omitted, and disabled executor cases called out for testing.
- AC8 is a strong absence-test hook: it covers Go production code, Core assets, formulas, orders, prompts, overlays, generated/materialized metadata, and route targets.

**Critical risks:**
- [Major] The role-name exception policy needs to be machine-enforceable before implementation planning relies on it. AC8 allows literal Gastown role names and `dog` only in "documented configured-default data, migration docs, generated review artifacts, and absence-test fixtures" (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:89`), but it does not define exact scan roots, denied token forms, allowed paths, or whether comments in Core source count as violations. The current Core source tree still contains role-specific content in formulas and skills, such as Polecat/Witness references and Gastown dispatch guidance (`internal/bootstrap/packs/core/formulas/mol-polecat-base.toml:2`, `:100`; `internal/bootstrap/packs/core/formulas/mol-polecat-commit.toml:2`, `:24`; `internal/bootstrap/packs/core/skills/gc-dispatch/SKILL.md:106`). The requirements point in the right direction, but the whitelist must be precise enough that these cannot survive accidentally.
- [Major] The default `dog` allowance is conceptually correct but still too broad at the wording boundary. AC9 says Core may ship a default maintenance executor named `dog` as pack configuration and Go must treat it as user-supplied config (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:90`), while the edge example says no-executor Core maintenance work should remain visible or diagnosable (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:72`). To prevent Go-side role leakage, the requirements should say exactly where unqualified `dog` literals are allowed: agent/pool config fields, Core pack defaults, test fixtures, or formula/order route targets. Without that line, a later implementation could treat Core formulas that route to literal `dog` as "configuration" even when they make executor renaming brittle.
- [Major] SDK self-sufficiency is stated, but the required proof boundary needs one more product definition. AC2 says normal SDK infrastructure operations must not depend on any Gastown role existing (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:83`), and W6H says SDK infrastructure remains self-sufficient with controller and required Core runtime support (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:57`). The requirements should enumerate which operations are non-negotiably controller-only: config resolution, doctor diagnostics, import-state checks, order discovery, gate evaluation, event observation, stale state reporting, and workflow bookkeeping. Otherwise "agent-executed Core maintenance" can blur into required infrastructure.
- [Minor] The open questions are appropriate, but they can still intersect ZFC. In-flight sessions using retired prompts/formulas and the exact repair workflow remain open (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md:129`, `:133`). Any compatibility behavior for those cases must remain path/provenance based, not role-name based, or it will recreate a framework-side Gastown exception.

**Missing evidence:**
- Exact denied-token list and allowed-path whitelist for Core role-neutrality scans.
- Positive-control and negative-control examples for the absence scan, especially `dog` in Core pack config versus `dog` in Go conditionals or hardcoded route decisions.
- A controller-only proof list naming which SDK operations must pass with no Gastown agents and with the default maintenance executor renamed or absent.
- A specific rule for whether Core formulas/orders may route to a literal default executor name or must resolve the executor through configurable pack metadata.

**Required changes:**
- Tighten AC8 with a concrete absence-scan contract: scan roots, forbidden token variants, allowed exception paths, generated-artifact exclusions, comment handling, and positive controls that prove the scan fails on a planted Core role reference.
- Tighten AC9 with a literal `dog` policy. State whether `dog` is allowed only in Core default agent/pool config, or also in Core formulas/orders, and require renamed/no-executor tests that exercise real route resolution rather than only config parsing.
- Expand AC2 with a short list of SDK infrastructure operations that must work with only the controller and Core config. Mark agent-executed maintenance as optional/configured work, not a prerequisite for SDK operation.
- Require the asset ledger to classify every existing Core role reference as moved, rewritten, retired, or migration-doc-only, so current `internal/bootstrap/packs/core` role references cannot be normalized as acceptable Core content.

**Questions:**
- Is an unqualified `dog` route target allowed inside Core formulas/orders if it comes from Core pack data, or must Core route through a configurable executor binding?
- Should role-name comments in canonical Core source fail the absence scan, or only executable/configuration content?
- Which operations define the minimum controller-only SDK surface for the no-Gastown/no-executor test case?
