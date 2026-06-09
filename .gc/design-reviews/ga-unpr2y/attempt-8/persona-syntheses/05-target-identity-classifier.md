# Amara Diallo

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Historical alias handling is unsafe as currently described. Claude and DeepSeek both found `alias_history` is used by pool resume and orphan-release logic as a live scheduling input, while Codex confirms package/API live lookup must deny historical aliases by default. The design must split mailbox expansion, pool read-for-routing, and selectable target behavior instead of labeling historical aliases broadly as mail/query or inspect-only.
- [Blocker] The read-only classifier boundary conflicts with current repair and materialization behavior. Claude found `ResolveSessionID` repairs empty-type session beads during resolution, DeepSeek found assignee normalization performs a materialization retry plus `RepairEmptyType`, and Codex found API assignment currently materializes configured named identities. The design does not yet say which adapter preserves these writes, which command owns them, or which behavior is intentionally amended.
- [Blocker] Target kinds and precedence are incomplete or ambiguous for caller parity. Codex found CLI qualified-alias basename fallback is missing as a distinct candidate kind; DeepSeek found the candidate-kind table can be read as contradicting API configured-named-before-package precedence; Claude found `rejected-by-config` is modeled too much like an independent tier rather than a post-filter on a selected package-resolver match.
- [Blocker] Closed exact-ID behavior is mis-mapped. Codex and DeepSeek both found exact session bead IDs can resolve even when closed, with live-only operation policy rejecting later; `allowClosed` only changes session_name and alias fallback. Slice 1 proof omits `SESSION-LIFE-007` and lacks a per-surface closed/closing/archived direct-ID matrix.
- [Blocker] Mail send and mail query are not safely modeled by one generic target row. DeepSeek found mail-specific helpers directly query `NamedSessionIdentityMetadata` with custom filters and ambiguity handling; the design must state whether those helpers are in scope for centralization or deliberately remain mail-specific.
- [Major] Assignee normalization remains too coarse. Codex and DeepSeek both require the configured-named materialization path, missing-work no-materialization path, open-bead validation, store mutation order, output/status shape, and repair behavior to be explicit before the classifier/adapters touch assignment.
- [Major] Candidate diagnostics are not enforceably non-authoritative. Claude and DeepSeek both flag that callers could reinterpret `candidates[]` or pick the first demoted fact unless the contract states that adapter `selected` output is the sole action authority and candidate snapshots are diagnostic-only.
- [Major] Collision and fallback behavior still needs exact fixtures. Reviewers called out demoted dual alias/session_name collisions, configured named conflicts, path-alias shadowing, path-alias tiebreakers, template targets, bare ordinary config targets, rejected-by-config ordering, and live mailbox ambiguity.

**Disagreements:**
- Claude says the per-surface compatibility chains are mostly accurate; DeepSeek says the candidate-kind list can still imply the wrong API precedence. My assessment is that both can be true: the surface chains may be right, but the taxonomy table must explicitly be unordered or be reordered to prevent implementer drift.
- Reviewers use different severity for the repair/materialization issue. Claude labels it major caller-preservation risk; Codex and DeepSeek effectively make it blocking because Slice 1 cannot stay read-only while preserving current writes without a separate adapter/command contract. This synthesis treats it as a blocker.
- Historical alias behavior is surface-dependent. Codex emphasizes denial for package/API/CLI selectable targets; Claude and DeepSeek emphasize pool resume's live scheduling dependency. The required model is not one global rule, but an allow/deny table with pool resume as read-for-routing-but-never-writable if preserved.
- Codex asks whether CLI qualified-alias basename fallback is long-term compatibility or a migration candidate. Until the owner decides otherwise through requirements, it must be modeled and tested as current behavior.
- DeepSeek asks whether closed exact IDs should be classified as `closed-not-allowed` directly or returned as candidates for adapter rejection. The common requirement is to document the split and prove every adopting surface's behavior.

**Missing evidence:**
- A Slice 1 fixture list for CLI `qualified-alias-basename`: single match, ambiguity, no-slash input, configured-named precedence, and distinction from API Title/path-alias behavior.
- A historical-alias allow/deny table covering package, API/Huma, CLI, mail send, mail query, assignee normalization, pool resume, orphan release, nudge, attach, inspect, log, transcript, and extmsg.
- Assignee normalization parity proof for configured-named materialization retry, missing-work no-materialization, open-bead validation, assignment key output, status code, response body, resulting session metadata, and `RepairEmptyType`.
- A direct closed-ID fixture matrix tied to `SESSION-LIFE-007` for wake, nudge, submit/message, attach, close/suspend, read-only inspect, log/transcript, assignee normalization, and extmsg surfaces.
- A decision on whether `resolveLiveConfiguredNamedMailTarget`, `mailRecipientsForNamedSession`, and other mail recipient helpers are centralized in Slice 1 or remain mail-specific store queries.
- Writer inventory and requirements coverage for `RepairEmptyType` writes in `internal/session/resolve.go`, including whether read-side repair is preserved, moved to adapters, or retired with owner approval.
- An enforceable `TargetSelection` schema that marks `selected` as the sole action authority and candidate snapshots as diagnostic-only, including demoted-candidate representation.
- Current oracles for `rejected-by-config` ordering, path-alias ambiguity/tiebreaking, live configured mailbox ambiguity, and template-factory rejection.

**Required changes:**
- Add a separate `qualified-alias-basename` candidate kind limited to CLI/local fallback surfaces, with precedence after configured named and package live resolution and before allow-closed lookup.
- State that candidate kinds are an unordered taxonomy and all precedence lives in per-surface compatibility chains, or reorder the table so it cannot contradict API configured-named precedence.
- Reconcile historical alias behavior into explicit policies: mailbox-address expansion, pool resume/orphan-release read-for-routing, and default denial for selectable targets unless current tests prove otherwise.
- Resolve the read-only classifier/write-side-effect contract. Use `repair-needed` and `requires-materialization` candidates with adapter-owned audited writes, or amend requirements to retire the writes; do not leave `ResolveSessionID` parity and no-write rules in conflict.
- Expand the assignee normalization contract with the two-call materialization retry, open-bead status check, missing-bead behavior, `RepairEmptyType` invocation, mutation order, and response/output parity.
- Document that exact-ID lookup can return closed session beads on all paths, and require adapters to prove live-only rejection versus read-only acceptance per surface.
- Split mail send and mail query compatibility chains, and explicitly scope direct `NamedSessionIdentityMetadata` lookups as centralized or mail-private.
- Make adapter `selected` output the only authority for action. Keep candidate diagnostics unexported to callers or add a guard/test that prevents callers from ranking `candidates[]` themselves.
- Specify `rejected-by-config` as a post-filter on a selected package-resolver named bead, and add fixtures for its position relative to path-alias fallthrough and template-factory rejection.
- Define path-alias equal-timestamp tiebreaking and ambiguity behavior, then add fixtures for shadowing and fallback cases.
