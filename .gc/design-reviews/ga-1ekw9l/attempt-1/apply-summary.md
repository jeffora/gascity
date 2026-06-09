# Apply Summary

Source synthesis: `.gc/design-reviews/ga-1ekw9l/attempt-1/synthesis.md`

Global verdict: `block`

Recommended apply verdict: `iterate`

Applied changes:

- Rewrote `plans/core-gastown-pack-migration/implementation-plan.md` into the
  required `gc.mayor.implementation-plan.v1` section order.
- Removed the long attempt-history strata from the normative plan body.
- Consolidated durable state, generated artifacts, public-pack pins, cache
  authority, runtime-state migration, and recovery records under
  `Data And State`.
- Added enforceable contracts for required Core loading, public Gastown behavior
  evidence, doctor mutation safety, retired-source classification, role
  neutrality, bootstrap fixture isolation, docs/generated-reference gates, and
  staged rollout/recovery.
- Marked public `gascity-packs` commits, generated manifests, pin ledgers,
  ownership rows, and packcompat transcripts as external prerequisites for
  source deletion and Maintenance removal.

Unfixable in this document-only apply step:

- The public `gascity-packs` compatibility and activation commits are external
  artifacts and are not created by this apply step.
- Behavior-preservation manifests, role-surface manifests, docs wording
  manifests, and packcompat transcripts are future generated artifacts; the plan
  now names their required shape, authority, and rollout gates.

Outcome: set `design_review.verdict=iterate` so the design-review workflow can
review the schema-valid implementation plan and the explicitly declared
external prerequisites.
