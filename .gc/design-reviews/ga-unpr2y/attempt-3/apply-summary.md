# Apply Summary

Attempt: 3
Synthesis: `.gc/design-reviews/ga-unpr2y/attempt-3/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

## Changes Applied

- Updated `internal/session/DESIGN.md` to record attempt 3 as the latest
  blocking review and keep the design in iterate state.
- Added an attempt-3 response section that makes implementation blocked until
  the new inventories, proof, and guards are created or verified.
- Expanded mutation ownership into an owned-key taxonomy, writer inventory
  (`W-001` through `W-021`), and concrete AST/symbol static-guard contract.
- Rewrote target classification around compatibility resolver chains for
  package resolution, API/Huma, mail, extmsg, attach/observe/log/transcript,
  and assignee/circuit resolution instead of one universal precedence table.
- Strengthened scenario proof requirements so every touched scenario row needs
  fresh proof, and called out missing stale proof paths that block later slices.
- Made runtime start a single command slice covering prepare, commit, rollback,
  pending-create cleanup, `instance_token`, `session_key`, and runtime hashes.
- Replaced generic lock language with versioned commit markers, precondition
  re-reads, `instance_token` runtime identity, and partial-write repair proof.
- Added a per-event/reaction matrix that makes durable scans authoritative for
  work release, drain recovery, wake demand, and other critical convergence.
- Separated controller-owned scheduling, pool demand, provider health, progress,
  circuit, and budget state from session-owned lifecycle facts.
- Added per-slice coexistence, bake, revert, diagnostics, typed-wire, and
  performance/query-budget gates.

## Notes

- No code or tests were changed.
- The minor attempt-scoped persona artifact placement bug is workflow plumbing
  outside the requested `internal/session/DESIGN.md` update and was not changed.
