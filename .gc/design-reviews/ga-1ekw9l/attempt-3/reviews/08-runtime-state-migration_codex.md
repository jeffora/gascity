# Yelena Markovic - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan names the relevant runtime-state families instead of treating this as a generic path rename: JSONL archive state, spawn-storm ledgers, refs/remotes, escalation fields, pending archive push state, and order skip/tracking compatibility aliases (`design-before.md:266`-`design-before.md:269`).
- The migration marker is non-destructive and stateful: it records schema version, old/new paths, staged archive digest, completed steps, and old-binary post-marker write detection (`design-before.md:269`-`design-before.md:273`).
- Data And State carries the migration details forward, including staged archive-copy digest, push-cursor reconciliation, rollback state, re-upgrade state, and retained legacy state handling (`design-before.md:377`-`design-before.md:381`).

**Critical risks:**
- [Major] The test plan should explicitly cover interrupted archive-copy and duplicate-writer cases. It currently says runtime-state migration detects old-binary writes, reconciles push cursors, preserves retained legacy state, and supports downgrade/manual recovery, but this persona needs a round-trip test where archive copy is interrupted mid-stream, another writer appends legacy state, and rerun either completes safely or reports deterministic manual recovery.
- [Major] Destination-completeness semantics are not named. The plan records a staged archive digest, but it should define how destination archives are marked complete and how partial destination files are distinguished from valid migrated state after crash or concurrent writer interference.
- [Minor] Order skip/tracking compatibility aliases are named, but the plan should require a test proving renamed/moved orders neither lose existing skip-list entries nor duplicate pending order/push work.

**Missing evidence:**
- A round-trip fixture for interrupted archive copy, in-flight push cursor, duplicate legacy writer, rerun, and rollback to older binary.
- A completeness marker or atomic publish invariant for migrated JSONL/archive files.
- A test matrix for old binary writing before marker, during staged copy, after marker, and after new binary rollback.
- Proof that order skip/tracking aliases map old Maintenance-owned names to new Core-owned names without duplicate execution.

**Required changes:**
- Add an explicit migration round-trip test covering interrupted archive copy, in-flight push state, duplicate writers, post-marker old-binary writes, and rollback/downgrade behavior.
- Define destination archive completion semantics: staged file naming, digest validation, commit marker, and how partial copies are ignored or resumed.
- Require compatibility tests for order skip/tracking aliases and pending push state so moved Core paths do not double-send or drop work.
- State that retained legacy state remains read-only unless reconciliation proves safe; otherwise the operator receives manual recovery guidance.

**Questions:**
- What exact file or metadata field marks a migrated archive copy as complete?
- How does the new binary distinguish an old-binary append after marker from a staged copy that was still in progress?
- Are old Maintenance order identifiers preserved as aliases indefinitely, or only through a bounded migration window?
