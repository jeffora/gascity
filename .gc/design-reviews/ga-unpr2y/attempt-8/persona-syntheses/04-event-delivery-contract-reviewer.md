# Amara Osei

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The close/work-release recovery backstop is not concrete enough to prove crash-after-commit convergence. Codex blocks because the design does not name the closed-bead query, index, helper, bounded scan, or store coverage that rediscovers `work_release_pending` session beads after restart. Claude finds the current code has only inline release sites and no periodic recovery sweep. DeepSeek approves the `work_release_pending` concept, but still asks how the flag is cleared and how successor assignments are distinguished.
- [Blocker] Current clear-identity-then-release ordering can strand work, and the design does not name the concrete inversion. Claude shows `RetireNamedSessionPatch` clears `session_name`, `alias`, and `session_name_explicit` before the inline release helpers run from in-memory snapshots. A crash after metadata commit and before release can leave old work unreleasable unless a durable release-key snapshot or pending marker is written before identity clearing.
- [Major] Successor-safety for identity-based work release remains under-proven. DeepSeek warns that work beads do not inherently carry the closing session's `instance_token` or generation, so stale release scans for an old named identity can either release successor-owned work or suppress release and strand old work. Claude requires the live-successor guard to be proven in the scan path, and Codex requires the release snapshot to contain enough old assignment keys.
- [Major] The close/retire backlog contradicts or dilutes the main recovery contract. Codex notes the main text assumes `work_release_pending` or equivalent durable close facts, while the backlog still says to "define whether" close uses synchronous release or post-commit facts. Claude similarly notes the recovery gate tests are weaker than the step-5 crash-after-close contract.
- [Major] Pending-release completion and retry semantics are underspecified. Codex asks for a scanner state machine, completion fact, clearing rule, duplicate-scan key, partial-query deferral behavior, and trace/doctor output for indefinitely deferred release. Claude also asks for idempotency-key stability across retries of the same logical close commit.
- [Major] Event identity context needs to be sufficient for stale-hint rejection. The reviewers agree events should remain post-commit facts/hints, not commands. The remaining gap is that thin events and `SessionCrashed` need stable idempotency or episode keys, generation/instance token when known, and a hard rule that tokenless events only trigger scans and are never sufficient mutation input.
- [Minor] Mail/extmsg binding cleanup ownership is unresolved. Claude and DeepSeek both call out the "adapter retry or controller repair" disjunction. The design should pick the controller repair scan as the safety backstop or explicitly classify missed member notifications as best-effort.

**Disagreements:**
- Claude and DeepSeek return `approve-with-risks`, while Codex returns `block`. My assessment: the persona verdict is `block` because the unresolved issue is the load-bearing recovery path for stranded work after close commits. If that scan cannot rediscover closed pending-release beads after restart, the event-delivery contract fails its critical recovery promise.
- DeepSeek treats `work_release_pending=true` and `state=closing` as already solving the discovery problem; Codex says the store query and bounded scan are not specified, and Claude's source check shows no existing periodic release sweep. My assessment: the concept is sound, but the design must make discovery, boundedness, and completion executable before approval.
- Reviewers propose different successor-safety mechanisms: token/generation stamped on work beads, preserved release-key snapshots, or live-successor suppression. Any mechanism is acceptable if it proves old close scans neither steal successor work nor strand old assignments.
- Public event payload migration can remain deferred. The disagreement is internal only: subscribers still need stable fact identity, dedup keys, and token-aware stale-hint behavior to avoid treating delayed events as authoritative commands.

**Missing evidence:**
- A closed-session pending-release scanner contract: exact helper name, store query, status and metadata filters, city/rig store coverage, boundedness limits, cadence, and completeness guarantee.
- Tests for crash after close/retire commit before inline cleanup, controller restart, rediscovery of the closed pending-release bead, duplicate scans, partial work-query deferral, completion clearing, and stale successor suppression.
- A durable release-key schema or close transitional state that records bead ID, session name, alias/current and historical assignment keys if relevant, configured identity, close generation, instance token, and store ref before identity metadata is cleared.
- Proof that work assigned by session name or configured identity converges to released after `RetireNamedSessionPatch` clears the active identity fields and the closing process exits.
- A scanner state machine for pending -> deferred -> converged/blocked/failed, including when `work_release_pending` is cleared and what trace/doctor output exposes stuck recovery.
- Duplicate-dedup keys for `SessionCrashed` and retry-stable idempotency keys for each committed session fact.
- A chosen recovery owner for mail/extmsg binding cleanup and classification for missed member-notification.

**Required changes:**
- Add an explicit closed-bead recovery queue for work release. Name the scanner/helper, exact query shape, filters, store coverage, boundedness budget, cadence, and completeness guard for finding closed or closing session beads with pending release after process restart.
- Make the close/retire durable recovery mechanism mandatory in the backlog and command contract. Remove open "define whether" wording unless the alternate synchronous mechanism proves close cannot commit before release is durable.
- Require release identity snapshot or `work_release_pending` to be written before active identity fields are cleared, or release work before clearing identity. Name the current `RetireNamedSessionPatch` call-site inversion and W-008 as migration inputs.
- Define pending-release completion semantics: idempotency key, supersession key, duplicate-scan behavior, partial-query deferral, completion fact, clearing rule, and blocked/failed observability.
- Prove successor safety by stamping work assignments with session token/generation, preserving an equivalent durable assignment-key snapshot, or otherwise demonstrating stale old-session cleanup cannot affect successor-owned work.
- Add crash-after-commit, skipped-event, duplicate-scan, partial-query, process-restart, and stale-successor tests for close/work release and identity retirement.
- Give `SessionCrashed` a durable duplicate-dedup key and state that tokenless thin events are scan hints only, never sufficient mutation input.
- Choose the mail/extmsg binding cleanup recovery owner and classify missed extmsg member-notification as controller-recovered or explicitly best-effort.
