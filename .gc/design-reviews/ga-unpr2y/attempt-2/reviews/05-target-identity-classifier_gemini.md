# Amara Diallo — DeepSeek V4 Flash (Independent Review, Attempt 2)

**Verdict:** block

**Persona:** target taxonomy, alias precedence, conflict cases, caller behavior preservation.

**Reviewed against:** `internal/session/DESIGN.md` (attempt 2), `internal/session/REQUIREMENTS.md`, `internal/session/resolve.go`, `internal/session/named_config.go`, `internal/session/lifecycle_projection.go`, `internal/api/session_resolution.go`, `internal/api/handler_mail.go`, `internal/api/handler_beads.go`, `internal/mail/resolve.go`, attempt-1 reviews, and attempt-2 cross-persona reviews.

---

## Top Strengths

- The target classification contract is the right extraction to go first. Separating what a token *is* from what an operation may *do* with it is the minimum viable boundary for a session ownership model, and the design names this explicitly at `DESIGN.md:165–170`.
- The operation policy matrix (`DESIGN.md:185–193`) is structurally correct. It maps seven operation surfaces against six target kinds and gives implementers a per-cell permission rule rather than vague prose. This is what was missing from the initial design.
- The adversarial collision test requirement (`DESIGN.md:207–209`) is the right regression gate for a change this central. Same-token inputs across bead ID, session_name, alias, named identity, path alias, and template prefix must all produce stable, classified results.
- The vocabulary constraint ("do not add a broad session facade") at `DESIGN.md:571` and the per-slice fact introduction rule at `DESIGN.md:79–86` prevent the most common scope-creep failure mode for this kind of extraction.

## Critical Risks

### [Blocker] The Precedence Table Misorders Three Active Code Paths — Same Finding As Attempt 1, Unresolved

The design's precedence table at `DESIGN.md:172–183` orders `live-session-name` (2) and `live-alias` (3) before `configured-named-identity` (4). The actual API resolver does the opposite: `resolveConfiguredNamedSessionIDWithContext` runs at `session_resolution.go:441` before `ResolveSessionID` at `:446`. This is not a stylistic difference — it determines whether a bare token like `mayor` resolves to the configured named canonical bead or to an unrelated live session whose `session_name` happens to match the bare leaf. The attempt-1 Gemini review identified this misorder and the design has not corrected it.

Three additional misalignments compound this:

1. **Template-prefix rejection is a pre-classification guard, not tier 9.** `parseAPITemplateTarget` runs at `session_resolution.go:434`, *before* exact-ID resolution at `:436`. The design places `template-factory` at position 9 — after live/closed/historical targets. If the classifier follows the design's table, a token that is both a valid bead ID and a `template:` prefixed string (e.g., `template:ga-abc123`) would resolve as a direct session ID rather than being rejected. Current behavior rejects it immediately. The classifier must specify template-prefix rejection as a surface-specific pre-classification guard, not as a low-precedence classifier result kind.

2. **Exact bead ID returns closed beads regardless of closed-lookup policy.** `ResolveSessionIDByExactID` (`resolve.go:48–57`) returns any session bead by ID regardless of status, including closed ones. The design says `direct-session-id` means "Closed IDs are returned only if policy allows closed targets." This is inaccurate — the current resolver always returns exact-ID matches. The `allowClosed` flag only gates session_name/alias fallback. If the classifier changes this, `ResolveSessionIDAllowClosed` callers that pass a closed bead ID directly will break, and `handler_sessions.go:302` (GET session by ID) would start returning 404 for closed exact-ID matches that currently return 200.

3. **The post-resolve config rejection check has no classifier representation.** Between `ResolveSessionID` (`:446`) and `resolveLiveSessionByPathAlias` (`:457`), the API checks whether a resolved live named-session bead has a backing config (`:447–455`). If config was removed, it returns `apiSessionTargetRejectedByConfig`. The precedence table does not represent this check at all — it shows live targets at positions 2–3 and path alias at position 5 as if they are adjacent. The classifier must either include `rejected-by-config` as an inter-tier result or document it as a post-resolution policy step. Placing it nowhere is a behavior gap that breaks at least one existing test (`session_materialization_guard_test.go:77`).

### [Blocker] The Classifier Returns a Single Result, But at Least Three Callers Need Multi-Step Fallthrough

The design defines the classifier as returning one result kind from the precedence table. At least three callers need fallthrough to lower tiers:

1. **Allow-closed resolution.** `resolveSessionTargetIDWithContext` at `session_resolution.go:459–473` first tries live resolution, then falls through to `ResolveSessionIDAllowClosed` when `opts.allowClosed` is true. If the classifier returns only `not-found` for a live miss, the allow-closed caller must re-invoke the classifier with a different flag — meaning the classifier needs a second call mode and the caller needs to know when to retry. The current design does not define this retry protocol or the allow-closed flag's interaction with the classifier.

2. **Bead assignee normalization.** `normalizeRawBeadAssignee` (`handler_beads.go:63–78`) first calls `resolveSessionTargetIDWithContext` with `materialize: false`, then retries with `materialize: true` on `ErrSessionNotFound`. This means bead assignment can create a named session as a side effect of resolution. The operation policy matrix has an "Assignee normalization" row but no materialization permission for the retry case, and the classifier contract does not define how `materialize: false` versus `materialize: true` changes the result.

3. **Mail send resolution.** `resolveMailSendRecipientWithContext` (`handler_mail.go:72–110`) first tries `resolveLiveConfiguredNamedMailTarget` (a completely separate metadata query path that searches by `NamedSessionIdentityMetadata`, not through the general resolver), then falls through to `resolveSessionTargetIDWithContext`, then tries `configuredMailRecipientAddress` as a third fallback. This is a three-stage chain with two different "configured named" lookups that return different types. The classifier contract must either subsume both lookups or explicitly document that the mail path uses a separate resolution chain.

### [Blocker] The Mail Resolution Chain Is Structurally Different From the Session Resolution Chain, and the Design Treats Them as the Same Surface

The operation policy matrix at `DESIGN.md:189` has a single "Mail recipient" row. But mail send and mail query use two different resolution paths:

- **Mail send** (`resolveMailSendRecipientWithContext`): first `resolveLiveConfiguredNamedMailTarget` (queries `NamedSessionIdentityMetadata` directly, matches by bare leaf, skips qualified names, checks for ambiguous multi-match), then `resolveSessionTargetIDWithContext`, then `configuredMailRecipientAddress` (finds named session spec by target, returns spec.Identity). The first step is a completely separate code path from the session classifier.

- **Mail query** (`resolveMailQueryRecipientsWithContext`): first `findNamedSessionSpecForTarget` (finds configured spec), then `resolveSessionTargetIDWithContext`, then falls through to raw `mail.ResolveRecipient` when no store is available. The order of these two lookups is reversed compared to mail send.

Both mail paths also use `mail.ResolveRecipient` when no store is available (`handler_mail.go:79–83, 121–127`). `mail.ResolveRecipient` has its own precedence: `human` → qualified name → bare name. The design must either state that `mail.ResolveRecipient` is out of scope for the session classifier, or define how the classifier subsumes it. A single "Mail recipient" row in the operation matrix is insufficient.

### [Blocker] `RepairEmptyType` Makes the "Read-Only" Classifier Not Read-Only

Both `ResolveSessionIDByExactID` and `listSessionBeadsByMetadata` call `RepairEmptyType` (`resolve.go:56, 142`), which writes to the bead store when a session bead has an empty `Type` field. The design says the classifier is read-only and the first slice has no mutation (`DESIGN.md:165`), but the existing resolution functions it must replace are NOT read-only. If the classifier wraps these functions, it inherits their side effects. If the classifier re-implements them, it must either also call `RepairEmptyType` (making it not read-only) or skip it (changing behavior for crashed-or-migrated beads that have empty types).

The design must specify one of: (a) `RepairEmptyType` runs as a pre-classification adapter step before the classifier sees the token, (b) the classifier calls it and the "read-only" claim is softened to "no lifecycle mutation," or (c) the classifier skips it and the test plan accounts for the behavior change.

## Major Risks

### [Major] The Allow-Closed Path Has a Named-Session Guard That the Design Does Not Represent

At `session_resolution.go:459–463`, when `opts.allowClosed` is true, the code first calls `findNamedSessionSpecForTarget`. If it finds a configured named session for that identifier, it returns `apiSessionTargetNotFound` — NOT `apiSessionTargetRejectedByConfig` and NOT falling through to `ResolveSessionIDAllowClosed`. This prevents closed lookup from accidentally hitting a named-session candidate that should be materialized, not inspected as a closed session. The design's closed-lookup tier (`DESIGN.md:178–179`) does not mention this guard. If the classifier returns `closed-session-name` for a token that is also a configured named identity, the allow-closed caller would return the closed bead instead of returning not-found (which would trigger the materialization path). This is a behavior change.

### [Major] Path-Alias Constraints Are Not Captured in the Classifier Contract

`resolveLiveSessionByPathAlias` (`session_resolution.go:392–423`) applies three constraints that the design's `path-alias` tier does not mention:

1. Named-session beads are excluded (`apiIsNamedSessionBead(b)` skip at line 403).
2. Only `active`, `awake`, and `none` states are accepted (state filter at line 406–408).
3. Most-recently-created bead wins as tiebreaker (line 410–412).

If `path-alias` becomes a general session-domain result kind without these constraints, an API-specific compatibility shim becomes durable session vocabulary. The attempt-2 Codex review for persona 09 raised this; the design should either keep path-alias fact gathering entirely adapter-owned or spell out these constraints in the classifier contract.

### [Major] `materializable` and `materialization_reason` Mix Classification Facts With Operation Policy

The classifier result fields at `DESIGN.md:180` include `materializable` and `materialization_reason` for `configured-named-identity` and `template-factory` result kinds. Whether a target *can* be materialized is an operation-policy question that depends on the calling surface. Whether a target *is* unmaterialized is a classification fact. The classifier should return `configured-named-unmaterialized` as a fact; the operation policy matrix already defines which surfaces may materialize. Adding `materializable` to the classifier result forces the classifier to know about operation policy, violating the design's own separation principle.

### [Major] The Dual Alias/Session-Name Demotion Is a Resolution Detail That Affects Classification Semantics

`filterOutAliasMatches` (`resolve.go:148–169`) demotes dual alias/session_name beads when another session_name-only match exists. The design's `live-session-name` tier at position 2 does not mention this demotion. The classifier must either: (a) expose it as a candidate detail (e.g., a `demoted_alias` candidate in `candidates[]`), or (b) apply it internally and document that `live-session-name` results already have alias matches demoted. The choice affects diagnostics and caller behavior — a caller that receives a `live-session-name` result should know whether an alias match was suppressed.

### [Major] Historical Alias Has No Current Implementation and No REQUIREMENTS.md Row

The design places `historical-alias` at position 8 in the precedence table. The current `ResolveSessionID` and `ResolveSessionIDAllowClosed` functions do not resolve historical aliases — they only search `session_name` and `alias` metadata keys with optional closed fallback. `REQUIREMENTS.md` has no `SESSION-ID-*` row for historical alias resolution. Including it in the slice-1 classifier adds a result kind with no current behavior to preserve, no tests to cite, and no requirement to justify. The attempt-1 Gemini review recommended marking it as a slice-2+ extension; the design has not done so.

### [Major] The Scenario Traceability Matrix Does Not Map Classifier Result Kinds to REQUIREMENTS.md Rows

The slice-1 traceability row at `DESIGN.md:221` lists `SESSION-ID-003` through `SESSION-ID-010` but does not map which classifier result kind corresponds to which scenario row. For example:

- `SESSION-ID-003` describes `ResolveSessionID` behavior (direct ID → session_name → alias). The API resolver adds configured-named-identity and path-alias tiers that `SESSION-ID-003` does not cover. Which test proves that the classifier's `configured-named-identity` result preserves the API's named-identity-before-live-name behavior?
- `SESSION-ID-009` says "Bare configured named session mail uses the configured mailbox without materializing a session." The design's operation policy matrix says mail configured named identity is "allowed." But `resolveLiveConfiguredNamedMailTarget` is a completely separate code path from the classifier. Which test proves that the classifier's `configured-named-identity` result preserves the mail send resolution behavior?

Without this mapping, reviewers cannot verify that the classifier preserves all affected scenario rows, and implementers cannot tell which test proves which classifier result kind.

## Minor Risks

### [Minor] `ordinary-config-target` Has No API Resolution Path Today

The design places `ordinary-config-target` at position 10. The current API resolver does not resolve bare ordinary config names to sessions — it only resolves configured named identities, live identifiers, and path aliases. A bare agent name like `claude` without a `[[named_session]]` entry is not resolvable through `resolveSessionTargetIDWithContext`. If the classifier returns `ordinary-config-target`, it invents a result kind with no current behavior. The design should defer it unless a requirement row exists.

### [Minor] The Operation Policy Matrix Lacks an Extmsg Bind/Deliver Row With Concrete Permission

The matrix row for "Extmsg bind/deliver" says "allowed by binding policy" for live targets and "no implicit materialization" for configured named identity. But `extmsg` delivery resolution (`internal/extmsg/`) has its own session-targeting logic that may differ from the general API path. The design should cite the exact extmsg resolution entry point or explicitly mark it as a separate adapter surface.

## Required Changes

1. **Rewrite the precedence table to match the current API resolver order** or explicitly mark and justify any behavior change with product-owner approval and a `REQUIREMENTS.md` update. The correct order for API-style surfaces is: (a) `template:` prefix rejection (surface-specific pre-classification guard, not a tier), (b) exact bead ID (always returned, including closed), (c) configured named identity (canonical/conflict/unmaterialized), (d) live `session_name` with dual-name demotion, (e) live alias, (f) post-resolve config rejection check, (g) path alias with named-session skip and state filter, (h) allow-closed fallback (with named-session guard before closed lookup). Each row must cite the current implementation path and test.

2. **Define whether the classifier returns a single result or all candidates.** If single-result, specify the multi-call retry protocol for allow-closed, materialize, and assignee surfaces, including the exact flag changes and fallthrough conditions. If all-candidates, define the candidate schema and the contract for callers to apply their own precedence.

3. **Separate the mail resolution chain from the session resolution chain in the operation policy matrix.** Add distinct rows for mail send and mail query. State whether `mail.ResolveRecipient` is in scope for the session classifier or remains a separate agent/mail recipient taxonomy. Document that `resolveLiveConfiguredNamedMailTarget` queries `NamedSessionIdentityMetadata` directly rather than going through the general session resolver.

4. **Specify `RepairEmptyType` handling.** Either make it a pre-classification adapter step, document that the classifier inherits the side effect and soften the "read-only" claim, or skip it with a test plan for the behavior change.

5. **Add `rejected-by-config` as an explicit classifier result kind** representing the post-resolve check between `ResolveSessionID` and `resolveLiveSessionByPathAlias`. It is neither `not-found` nor `ambiguous` — it is a config-policy rejection of a resolved live target.

6. **Remove `materializable` and `materialization_reason` from classifier result fields.** Return `configured-named-unmaterialized` as a fact; let the operation policy matrix define materialization permissions.

7. **Add the allow-closed named-session guard to the classifier contract.** Before returning `closed-session-name` for a token that is also a configured named identity, the classifier (or its caller) must check `findNamedSessionSpecForTarget` and return `not-found` instead of `closed-session-name`, matching the current API behavior.

8. **Capture path-alias constraints in the classifier contract.** Named-session skip, active/awake/none state filter, and most-recent tiebreaker must be specified if `path-alias` becomes a session-domain result kind.

9. **Defer `historical-alias` and `ordinary-config-target` until requirement rows exist.** Neither has a current implementation path or a `REQUIREMENTS.md` scenario row for slice 1.

10. **Add a per-result-kind-to-scenario-row mapping** to the scenario traceability matrix so reviewers can verify that each classifier result kind is covered by a proof test.

## Questions

- When the classifier returns `direct-session-id` for a closed bead, should the `closed` field on the result cause mutating operations to reject it, or should the caller check status independently? If the classifier enforces closed rejection, it violates classification-policy separation. If the caller checks, it re-derives status — which partially defeats the classifier's purpose.

- Should the dual alias/session_name demotion (`filterOutAliasMatches`) be exposed as a classifier fact (e.g., a `demoted_alias` candidate) or hidden as an internal classifier implementation detail? Exposing it helps diagnostics and parity testing; hiding it simplifies the contract.

- Is the design's intent that the classifier replaces `resolveSessionTargetIDWithContext` entirely, or that they coexist during a migration period? If coexisting, the design needs a migration contract specifying when callers switch and how parity is verified during overlap. The prior Gemini review asked this; the design has not answered.

- Should `normalizeRawBeadAssignee`'s materialize-on-not-found retry be modeled as a classifier call with a different policy input (e.g., `materialize: true`), or should the assignee adapter handle the retry independently after receiving `configured-named-unmaterialized` from the classifier?
