# Amara Diallo — DeepSeek V4 Flash (Independent Review, Attempt 3)

**Verdict:** block

**Persona:** target taxonomy, alias precedence, conflict cases, caller behavior preservation.

**Reviewed against:** `internal/session/DESIGN.md` (current checkout), `internal/session/REQUIREMENTS.md`, `internal/session/resolve.go`, `internal/session/named_config.go`, `internal/session/lifecycle_projection.go`, `internal/api/session_resolution.go`, `internal/api/handler_beads.go`, `internal/api/handler_mail.go`, `internal/mail/resolve.go`, attempt-1 reviews (all three lanes), attempt-1 persona synthesis, and attempt-2 Gemini review.

---

## Top Strengths

- The target classification contract is the right extraction to go first. Separating what a token *is* from what an operation may *do* with it is the minimum viable boundary for a session ownership model, and the design names this explicitly at `DESIGN.md:165–170`.
- The operation policy matrix (`DESIGN.md:185–193`) is structurally correct. It maps seven operation surfaces against six target kinds and gives implementers a per-cell permission rule rather than vague prose.
- The adversarial collision test requirement (`DESIGN.md:207–209`) is the right regression gate for a change this central. Same-token inputs across bead ID, session_name, alias, named identity, path alias, and template prefix must all produce stable, classified results.
- The vocabulary constraint ("do not add a broad session facade") at `DESIGN.md:571` and the per-slice fact introduction rule at `DESIGN.md:79–86` prevent the most common scope-creep failure mode for this kind of extraction.
- The mutation landscape inventory (`DESIGN.md:102–155`) and command decision inventory (`DESIGN.md:157–230`) are thorough and give implementers a concrete baseline to track.

## Critical Risks

### [Blocker] The Precedence Table Misorders Three Active Code Paths — Same Finding As Attempts 1 and 2, Still Unresolved

The design's precedence table at `DESIGN.md:172–183` orders `live-session-name` (2) and `live-alias` (3) before `configured-named-identity` (4). The actual API resolver does the opposite: `resolveConfiguredNamedSessionIDWithContext` runs at `session_resolution.go:441` before `ResolveSessionID` at `:446`. This is not a stylistic difference — it determines whether a bare token like `mayor` resolves to the configured named canonical bead or to an unrelated live session whose `session_name` happens to match the bare leaf. Both prior Gemini reviews identified this misorder. The design has not corrected it.

Three additional misalignments compound this, also unchanged since attempt 1:

1. **Template-prefix rejection is a pre-classification guard, not tier 9.** `parseAPITemplateTarget` runs at `session_resolution.go:433`, *before* exact-ID resolution at `:436`. The design places `template-factory` at position 9 — after live/closed/historical targets. If the classifier follows the design's table, a token that is both a valid bead ID and a `template:` prefixed string (e.g., `template:ga-abc123`) would resolve as a direct session ID rather than being rejected. Current behavior rejects it immediately. The classifier must specify template-prefix rejection as a surface-specific pre-classification guard, not as a low-precedence classifier result kind.

2. **Exact bead ID always returns closed beads regardless of closed-lookup policy.** `ResolveSessionIDByExactID` (`resolve.go:42–45`) delegates to `ResolveSessionBeadByExactID` (`resolve.go:50–63`), which returns any session bead by ID regardless of status, including closed ones. The design says `direct-session-id` means "Closed IDs are returned only if policy allows closed targets." This is inaccurate — the current resolver always returns exact-ID matches. The `allowClosed` flag only gates session_name/alias fallback at `resolve.go:37–38` and `session_resolution.go:464–474`. If the classifier changes this, `ResolveSessionIDAllowClosed` callers that pass a closed bead ID directly will break, and `handler_sessions.go` GET session by ID would start returning 404 for closed exact-ID matches that currently return 200.

3. **The post-resolve config rejection check has no classifier representation.** Between `ResolveSessionID` (`:446`) and `resolveLiveSessionByPathAlias` (`:459`), the API checks whether a resolved live named-session bead has a backing config (`:447–455`). If config was removed, it returns `apiSessionTargetRejectedByConfig`. The precedence table does not represent this check at all — it shows live targets at positions 2–3 and path alias at position 5 as if they are adjacent. The classifier must either include `rejected-by-config` as an inter-tier result or document it as a post-resolution policy step. Placing it nowhere is a behavior gap.

### [Blocker] The Classifier Returns a Single Result, But at Least Three Callers Need Multi-Step Fallthrough — Same Finding As Attempts 1 and 2, Still Unresolved

The design defines the classifier as returning one result kind from the precedence table. At least three callers need fallthrough to lower tiers:

1. **Allow-closed resolution.** `resolveSessionTargetIDWithContext` at `session_resolution.go:459–473` first tries live resolution, then falls through to `ResolveSessionIDAllowClosed` when `opts.allowClosed` is true. If the classifier returns only `not-found` for a live miss, the allow-closed caller must re-invoke the classifier with a different flag — meaning the classifier needs a second call mode and the caller needs to know when to retry. The current design does not define this retry protocol.

2. **Bead assignee normalization.** `normalizeRawBeadAssignee` (`handler_beads.go:63`) first calls `resolveSessionTargetIDWithContext` with `materialize: false`, then retries with `materialize: true` on `ErrSessionNotFound` (`handler_beads.go:72`). This means bead assignment can create a named session as a side effect of resolution. The operation policy matrix has an "Assignee normalization" row but does not address the retry-on-not-found pattern or the `RepairEmptyType` side effect at `handler_beads.go:89`.

3. **Mail send resolution.** `resolveMailSendRecipientWithContext` (`handler_mail.go:70`) runs `resolveLiveConfiguredNamedMailTarget` first, then falls through to `resolveSessionTargetIDWithContext`, then to `configuredMailRecipientAddress`. The mail send path has its own three-tier fallthrough that does not match the general API resolution chain. The operation policy matrix has a single "Mail recipient" row that cannot express this ordering.

The design must define whether the classifier returns a single result (with a multi-call retry protocol), all candidates (with a caller-side precedence contract), or a tiered result that includes fallthrough candidates. Leaving this undefined means every adapter invents its own retry logic.

### [Blocker] `RepairEmptyType` Write Side Effect Violates the Pure-Classifier Claim

`ResolveSessionBeadByExactID` (`resolve.go:50–63`) calls `RepairEmptyType(store, &b)` at line 56 when the bead has an empty type field. `RepairEmptyType` calls `store.Update` (`resolve.go:227`), which is a write. The same repair call occurs in `listSessionBeadsByMetadata` at `resolve.go:142`. And `normalizeRawBeadAssignee` also calls `RepairEmptyType` at `handler_beads.go:89`.

The design states that the classifier separates classification from operation policy, and the vocabulary section at `DESIGN.md:64–66` says fact types "describe already-gathered state; they do not perform store, runtime, config, or event I/O." But the exact-ID resolution path — which the classifier must handle — currently writes to the store as a side effect of reading. The design must address this:

- Either make `RepairEmptyType` a pre-classification adapter step (run before the classifier, documented as a repair guard), or
- Document that the classifier inherits the side effect and soften the "read-only" claim for the exact-ID path, or
- Skip the repair in the classifier and add a test plan proving the behavior change is safe.

The mutation landscape table at `DESIGN.md:139` mentions "Doctor, migration, repair paths" and "RepairEmptyType" as "Repair-only helpers with trace/log evidence", but does not acknowledge that the core resolution function itself performs this repair as an implicit side effect.

### [Major] Mail Resolution Has Its Own Precedence Chain, Not Addressed

`resolveLiveConfiguredNamedMailTarget` (`handler_mail.go:248`) bypasses the general session resolver entirely. It queries `NamedSessionIdentityMetadata` directly, iterates matching beads, filters closed sessions, matches identity basenames, and builds mail-specific display/recipients. It has its own conflict detection (`ErrAmbiguous` on multiple named-session mailboxes). This is not a policy wrapper over the proposed classifier — it is a completely separate resolution path.

Additionally, `resolveMailSendRecipientWithContext` has its own three-tier fallthrough: (1) `resolveLiveConfiguredNamedMailTarget`, (2) `resolveSessionTargetIDWithContext`, (3) `configuredMailRecipientAddress`. And `resolveMailQueryRecipientsWithContext` has a different order: (1) `findNamedSessionSpecForTarget` directly, (2) `resolveSessionTargetIDWithContext`, with no `resolveLiveConfiguredNamedMailTarget` step.

The design must state whether:
- The classifier subsumes `resolveLiveConfiguredNamedMailTarget`, or
- Mail resolution remains a separate adapter that calls the classifier as one of its stages, or
- The classifier and mail resolution are independent parallel paths

The current "Mail recipient" row in the operation policy matrix cannot represent the actual three-tier mail send fallthrough. At minimum, split it into separate "Mail send" and "Mail query" rows, and document `resolveLiveConfiguredNamedMailTarget`'s relationship to the classifier.

### [Major] Allow-Closed Named-Session Guard Not Captured

When `opts.allowClosed` is true, `resolveSessionTargetIDWithContext` at `session_resolution.go:465–469` first calls `findNamedSessionSpecForTarget`. If the token matches a configured named session, it returns `apiSessionTargetNotFound` *instead of* falling through to `ResolveSessionIDAllowClosed`. This guard prevents closed lookup from returning a closed bead for a token that is also a configured named identity.

The design's `closed-session-name` tier at position 6 says "Read-only lookup after live matches fail and policy allows closed." It does not mention that closed lookup is blocked for tokens that match a configured named identity. The classifier must either:
- Return `configured-named-identity` (unmaterialized) instead of `closed-session-name` when the token matches both, or
- Document the guard as a post-classification policy step that suppresses `closed-session-name` results, or
- Include it as a classifier fact (e.g., `named_config_match: true`) so callers can apply the guard.

### [Major] Path-Alias Contract Is Underspecified

`resolveLiveSessionByPathAlias` (`session_resolution.go:392–427`) applies three filters that the design's path-alias entry does not mention:

1. **Named-session skip.** Line 408–409: `if apiIsNamedSessionBead(b) { continue }`. Named-session beads are excluded from path-alias matching.
2. **State filter.** Lines 414–416: only `active`, `awake`, and `none` states are considered. Draining, suspended, closed, and other states are excluded.
3. **Most-recent tiebreaker.** Lines 418–420: `if !found || b.CreatedAt.After(best.CreatedAt)` — when multiple beads match, the most recently created wins.

If `path-alias` becomes a session-domain classifier result kind, these constraints must be part of the classifier contract, not hidden implementation details. The design's path-alias entry (`DESIGN.md:177`) says only "Identifier is an accepted path alias for a surface that opts in" — this does not capture any of the three filters.

### [Major] Result Field `materializable` and `materialization_reason` Violate Classification-Policy Separation

Prior reviews (attempts 1 and 2) recommended removing `materializable` and `materialization_reason` from classifier result fields. The design at `DESIGN.md:191` still includes them.

The classifier should return `configured-named-unmaterialized` as a fact (the token matches a configured named identity with no live bead). Whether the calling surface may materialize is an operation-policy concern that the operation policy matrix already defines. Embedding materialization facts in the classifier result means the classifier must know about operation permissions, violating the "separate what a token is from what an operation may do with it" principle at `DESIGN.md:165`.

### [Major] `historical-alias` and `ordinary-config-target` in Slice-1 Table Without Requirement Rows

The precedence table includes `historical-alias` (position 8) and `ordinary-config-target` (position 10). Neither has a `REQUIREMENTS.md` scenario row, a current implementation path in the API resolver, or a test that exercises the behavior. Including them in the slice-1 classifier means the classifier must produce and test result kinds that have no current behavior to preserve.

Defer both to a later slice. Add a `REQUIREMENTS.md` scenario row for each before implementing it. The precedence table should mark them as "slice 2+" extensions.

### [Major] `candidates[]` Schema Is Underspecified

The result fields at `DESIGN.md:192` list `candidates[]` with "kind, ID/name, status, and conflict reason" but do not define:

- Which candidates are returned for each result kind. For `ambiguous` (position 11), should all equal-precedence candidates be included? For `configured-named-identity` with a conflict, should the conflicting live beads be included?
- Whether candidates include lower-precedence matches. If a classifier returns `live-session-name`, should `candidates[]` include any `configured-named-identity` or `path-alias` matches for the same input?
- Whether the `demoted_alias` dual-name case is exposed. `filterOutAliasMatches` (`resolve.go:148–169`) demotes dual alias/session_name matches when a separate session_name-only match exists. Should the demoted alias appear in `candidates[]` for diagnostics?

### [Minor] Dual Alias/Session-Name Demotion Visibility

`filterOutAliasMatches` (`resolve.go:148–169`) implements a non-obvious tiebreaking rule: when a bead has both `alias` and `session_name` equal to the identifier, and a separate bead has only `session_name` matching, the dual bead is demoted from the session_name list. The design does not mention this demotion behavior or whether it should be exposed as a classifier fact. Prior reviews asked about this; the design has not answered.

Exposing it as a `demoted_alias` candidate in `candidates[]` would help diagnostics and parity testing. Hiding it simplifies the contract. Either choice should be documented.

### [Minor] `normalizeRawBeadAssignee` Materialization Retry and `RepairEmptyType` Side Effect

The assignee normalization path at `handler_beads.go:63–90` has two behaviors not captured in the operation policy matrix:

1. **Materialize-on-not-found retry.** Lines 72–74: first resolves with `materialize: false`, then retries with `materialize: true` on `ErrSessionNotFound`. This is a two-phase resolution that can create a named session as a side effect of bead assignment. The "Assignee normalization" row says "configured named identity may normalize to configured mailbox/identity as today" but does not address the retry trigger.

2. **`RepairEmptyType` call.** Line 89: `session.RepairEmptyType(store, &b)` is called on the resolved bead after confirmation it is open. This is a write side effect in the assignee adapter, not in the classifier, but it must be documented as a pre-existing adapter concern.

## Required Changes

1. **Rewrite the precedence table to match the current API resolver order** or explicitly mark and justify any behavior change with product-owner approval and a `REQUIREMENTS.md` update. The correct order for API-style surfaces is: (a) `template:` prefix rejection (surface-specific pre-classification guard, not a tier), (b) exact bead ID (always returned, including closed), (c) configured named identity (canonical/conflict/unmaterialized), (d) live `session_name` with dual-name demotion, (e) live alias, (f) post-resolve config rejection check, (g) path alias with named-session skip and state filter, (h) allow-closed fallback (with named-session guard before closed lookup). Each row must cite the current implementation path and test.

2. **Define whether the classifier returns a single result or all candidates.** If single-result, specify the multi-call retry protocol for allow-closed, materialize, and assignee surfaces, including the exact flag changes and fallthrough conditions. If all-candidates, define the candidate schema and the contract for callers to apply their own precedence.

3. **Address `RepairEmptyType` as a pre-classification side effect.** Make it an explicit adapter step that runs before the classifier (documented as a repair guard for broken beads), or acknowledge that the exact-ID path is not purely read-only and add a test plan for the side effect.

4. **Add `rejected-by-config` as an explicit classifier result kind** representing the post-resolve check between `ResolveSessionID` and `resolveLiveSessionByPathAlias`. It is neither `not-found` nor `ambiguous` — it is a config-policy rejection of a resolved live target.

5. **Separate the mail resolution chain from the session resolution chain in the operation policy matrix.** Add distinct rows for mail send and mail query. State whether `mail.ResolveRecipient` and `resolveLiveConfiguredNamedMailTarget` are in scope for the session classifier or remain separate agent/mail recipient taxonomies. Document that `resolveLiveConfiguredNamedMailTarget` queries `NamedSessionIdentityMetadata` directly rather than going through the general session resolver.

6. **Add the allow-closed named-session guard to the classifier contract.** Before returning `closed-session-name` for a token that is also a configured named identity, the classifier (or its caller) must check `findNamedSessionSpecForTarget` and return `not-found` or `configured-named-unmaterialized` instead of `closed-session-name`, matching the current API behavior.

7. **Capture path-alias constraints in the classifier contract.** Named-session skip, active/awake/none state filter, and most-recent tiebreaker must be specified if `path-alias` becomes a session-domain result kind.

8. **Remove `materializable` and `materialization_reason` from classifier result fields.** Return `configured-named-unmaterialized` as a fact; let the operation policy matrix define materialization permissions per surface.

9. **Defer `historical-alias` and `ordinary-config-target` until `REQUIREMENTS.md` scenario rows exist.** Neither has a current implementation path or a requirements scenario row for slice 1. Mark them as slice 2+ extensions in the precedence table.

10. **Define the `candidates[]` return contract per result kind.** Specify which candidates are returned for `ambiguous`, `configured-named-identity` with conflict, and other multi-match results. State whether lower-precedence matches are included and whether `demoted_alias` dual-name cases appear.

11. **Correct the `direct-session-id` closed-bead note.** Change "Closed IDs are returned only if policy allows closed targets" to "Exact bead ID matches are always returned regardless of session status; the `allowClosed` flag gates only session_name/alias fallback, not exact-ID returns."

12. **Document `normalizeRawBeadAssignee`'s materialize-on-not-found retry and `RepairEmptyType` call** in the assignee normalization operation policy row.

## Questions

- When the classifier returns `direct-session-id` for a closed bead, should the `closed` field on the result cause mutating operations to reject it, or should the caller check status independently? If the classifier enforces closed rejection, it violates classification-policy separation. If the caller checks, it re-derives status — which partially defeats the classifier's purpose.

- Should the dual alias/session_name demotion (`filterOutAliasMatches`) be exposed as a classifier fact (e.g., a `demoted_alias` candidate) or hidden as an internal classifier implementation detail? Exposing it helps diagnostics and parity testing; hiding it simplifies the contract.

- Is the design's intent that the classifier replaces `resolveSessionTargetIDWithContext` entirely, or that they coexist during a migration period? If coexisting, the design needs a migration contract specifying when callers switch and how parity is verified during overlap. Both prior Gemini reviews asked this; the design has not answered.

- Should `resolveLiveConfiguredNamedMailTarget` be explicitly documented as out of scope for the session classifier? It resolves named-session mailbox addresses (not session bead IDs) and has its own precedence (named identity metadata scan → session name/alias fallback → configured recipient address). The mail handler uses both the session classifier and `resolveLiveConfiguredNamedMailTarget` in sequence. The design should state whether the classifier subsumes this path or whether they remain independent.

- Should `normalizeRawBeadAssignee`'s materialize-on-not-found retry be modeled as a classifier call with a different policy input (e.g., `materialize: true`), or should the assignee adapter handle the retry independently after receiving `configured-named-unmaterialized` from the classifier?

- Should `RepairEmptyType` be a pre-classification adapter step (repair-before-classify) or should the classifier accept that the exact-ID path has a write side effect for broken beads and document it explicitly?
