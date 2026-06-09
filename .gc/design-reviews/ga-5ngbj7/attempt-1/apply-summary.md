# Apply Summary

- Source synthesis: `.gc/design-reviews/ga-5ngbj7/attempt-1/synthesis.md`
- Global verdict: `block`
- Apply verdict: `blocked`
- Blocked reason: `wrong_artifact_schema_mismatch`
- Requirements file changed: no
- Design after: `.gc/design-reviews/ga-5ngbj7/attempt-1/design-after.md`
- Diff: `.gc/design-reviews/ga-5ngbj7/attempt-1/design.diff`

The synthesis recommends `Recommended apply verdict: blocked`. The blocking
issue is a workflow/artifact mismatch: `internal/session/REQUIREMENTS.md` is a
module-local session behavior ledger, while `gc.mayor.requirements.v1` describes
a plan-level Mayor `requirements.md` artifact and explicitly excludes
module-local requirements ledgers.

No automatic edit was applied to `internal/session/REQUIREMENTS.md`. The next
non-design prerequisite is to decide the governing artifact/schema: either route
this ledger through a ledger-specific review contract, or provide a separate
plan-level Mayor requirements artifact. The evidence-frame and proof-inventory
findings also require external verification before another review can approve
the ledger.
