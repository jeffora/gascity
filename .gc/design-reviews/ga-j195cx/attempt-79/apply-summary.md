# Apply Summary

Verdict source: `.gc/design-reviews/ga-j195cx/attempt-79/synthesis.md`

Global verdict was `block`, so this apply pass addresses blocker and major
findings in `engdocs/design/formula-compiler-requirements.md` and sets the
workflow verdict to `iterate`.

Changes applied:

- Defined a Task Store-compatible durable diagnostic subject for formula-backed
  order failures and made grouped diagnostic persistence depend on explicit
  singleton/CAS store operations instead of a non-atomic read-list-create loop.
- Moved formula-backed order compile/accept preflight ahead of `order.fired`,
  fired metadata, order-run tracking, wisp creation, route metadata, hooks,
  convoys, and last-run history, with required fixtures for repeats,
  concurrency, restart, recovery, burst cadence, write failure, and config
  flaps.
- Added a single host-capability edge adapter contract and a bypass guard for
  legacy formula/molecule globals.
- Made the validation matrix vocabulary executable by aligning
  `caller_path` with the 18-row preflight lock, pinning authored
  `step.expand` semantics, and requiring decoded field-path construct
  predicates.
- Pinned operation-aware HTTP status rules and canonical warning
  message/remediation text across CLI, API, dashboard, events, reports, and
  accepted artifacts.
- Added release evidence for external discovery/outreach, alias-window clock
  anchoring, sliding alias-drain rules, and public notice/support-window
  fields.
- Specified `[pack] requires_gc` active-binary comparator behavior for release,
  prerelease, nightly, `main`, dirty, source-build, and unknown-version cases.
- Closed forward-compatibility gaps with axis identifier grammar, multiple
  unsupported future compiler minima, combined future-axis precedence, and
  default-capability artifact identity rules.

Unfixable blocker/major items: none documented; all synthesis blocker and
major findings were addressed by design text additions.

Artifacts saved:

- `.gc/design-reviews/ga-j195cx/attempt-79/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-79/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-79/apply-summary.md`
