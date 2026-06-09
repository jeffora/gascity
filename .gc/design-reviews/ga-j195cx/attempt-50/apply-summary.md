# Apply Summary

Source bead: `ga-j195cx`
Apply bead: `ga-wphtbjj`
Attempt: 50
Synthesis: `.gc/design-reviews/ga-j195cx/attempt-50/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

## Changes Applied

- Added parser and validation closure for attempt-50 edge cases: signed,
  zero/lower, leading-zero, overflow, whitespace/control, Unicode, NUL, and
  JSON-only requirement shapes; clarified that `check` is v2-only and must be
  dual-declared or rewritten during the alias window.
- Added blocking durable-write guard requirements for accepted artifacts,
  graph workflow synthesis, raw metadata consumers, convergence subset-parser
  use, and allowlist expiry.
- Tightened rollout gates with a calendar floor, release-captain authority,
  compatibility YAML as executable old-reader source, external-support status
  transitions, and first-party dual-declaration ownership.
- Reconciled diagnostics as a stable wire contract across CLI, API/Huma,
  API-routed CLI, dashboard/generated TypeScript, and events, including the
  disabled-host golden fixture and host-capability provenance.
- Added binding-aware provenance, migration hint, requirement-diff, root
  metadata, and persisted artifact requirements for external pack migration,
  including duplicate binding, aliased transitive import, shadowed winner,
  pinned SHA, and pack-floor fixtures.
- Specified read-only acquisition semantics for
  `gc formula validate --pack-source <url> --ref <ref> --all --json` and the
  `[pack] requires_gc` grammar, resolver behavior, and pack-floor remediation.
- Added a docs/reference acceptance gate and broadened stale-guidance scanning
  to reference docs, tutorials, generated docs, examples, first-party packs,
  and migration guidance including stale `version` and `GC_NATIVE_FORMULA=false`
  prose.
- Added compiler-owned convergence projection and active-root recovery
  contracts, including accepted-artifact reuse, migration repair, missing-source
  behavior, downgrade behavior, abandonment, cross-binary fixtures, and
  zero-write assertions.

## Unfixable Items

None. All synthesis `[Blocker]` and `[Major]` findings were addressed in the
design document. Because the global synthesis verdict was `block`, the workflow
must continue with `design_review.verdict=iterate`.
