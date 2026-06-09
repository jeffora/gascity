# Apply Summary

Verdict applied: `iterate`

Global synthesis verdict was `block`, so this pass addressed all blocker and
major findings in `engdocs/design/formula-compiler-requirements.md`.

Changes made:

- Added an attempt-102 review-closure table tying each blocker/major finding to
  a binding design correction.
- Split Phase 8 first-party requires-only conversion from Phase 9 parser alias
  removal, including a separate `--requires-only-conversion-gate`.
- Removed ignored runtime `.gc/system/packs` as a mandatory CI input and made it
  optional operator evidence unless deterministically fixture-generated.
- Replaced wildcard raw-consumer allowlist examples with exact file, symbol,
  and line-pattern constraints.
- Defined presence-aware diagnostic wire payloads, order failure compatibility,
  dashboard-less diagnostic status, grouped state/CAS expectations, and durable
  accepted-alias evidence.
- Preserved currently accepted legacy `contract` spellings as deprecated aliases
  during the compatibility window and added direct-pack validation identity plus
  old-reader pack-load fixtures.
- Split terminology for formula `[requires]`, pack `[[pack.requires]]`, and
  `[pack].requires_gc`; expanded stale-guidance and inventory coverage.
- Added the canonical `internal/formula` preflight API, workflow-root conflict
  semantics, parser matrix suites, fanout parent reuse rule, and live
  convergence call-graph anchoring.

Unfixable items: none.

Artifacts saved:

- `.gc/design-reviews/ga-j195cx/attempt-102/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-102/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-102/apply-summary.md`
