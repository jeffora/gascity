# Amara Diallo - Claude

**Verdict:** approve-with-risks

Lane: target taxonomy, alias precedence, conflict cases, caller behavior
preservation. Scope reviewed: the Target Classification Contract
(DESIGN.md:195-293) against the live resolver. I checked the design's 8-step
precedence line-by-line against `internal/api/session_resolution.go` and
`internal/session/resolve.go`, and the error projection against
`writeResolveError` / `humaResolveError`.

From this lane the contract is accurate and faithful: the stated precedence is
exactly what the code does, historical aliases are correctly excluded, and the
HTTP error mapping matches. None of my three red flags is actively violated. The
risks below are real but are sharpenings the *first behavior-moving* slice
(Slice 1) needs; they do not threaten the non-mutating Slice 0 that is the only
authorized work, so I approve with required changes.

**Top strengths:**
- The 8-step precedence (DESIGN.md:220-244) matches the implementation exactly,
  verified against `resolveSessionTargetIDWithContext`
  (`session_resolution.go:429-477`) and `resolveSessionID`
  (`resolve.go:65-122`): template→not-found, exact ID, configured-named *before*
  live, ordinary live (session_name then alias) with dual alias/session-name
  demotion (`filterOutAliasMatches`, resolve.go:148-169), config-orphan
  rejection, path-alias (non-named beads in active/awake/empty state, newest
  `CreatedAt` wins — `resolveLiveSessionByPathAlias`, lines 405-426), then the
  allow-closed tail. This directly defends my red flag #2 (order differs from
  existing tests): it does not.
- Historical aliases are correctly kept out of live targeting. `ResolveSessionID`
  resolves only by `session_name` and `alias` metadata keys and its doc comment
  states it "does not fall through to template, agent_name, or historical alias
  compatibility identifiers" (resolve.go:17-22). The contract preserves this, so
  my red flag #3 (historical aliases become writable live targets) is not
  triggered.
- The typed result schema separates `selected/not-found/ambiguous/rejected/
  repair-needed/store-error` and includes a `terminal_error` field explicitly
  required to "preserve `errors.Is` behavior and current Huma/legacy HTTP
  rendering" (DESIGN.md:264). The error projection it must preserve is real and
  consistent: `writeResolveError` (handler_sessions.go:213-222) and
  `humaResolveError` (huma_handlers_sessions.go) both map ambiguous +
  configured-name conflict → 409, `ErrSessionNotFound` → 404, else → 500.

**Critical risks:**
- **[Major] The `RepairEmptyType` side effect is pervasive across precedence
  steps, not a single point — the contract under-scopes it.** The design's
  quarantine (DESIGN.md:246-251) reads as if repair happens at one place, but
  `RepairEmptyType` (a real `store.Update`, resolve.go:222-229) is invoked from
  *both* `ResolveSessionBeadByExactID` (resolve.go:56, the direct-ID step) *and*
  `listSessionBeadsByMetadata` (resolve.go:142, every session_name and alias
  lookup, open and closed). It also feeds `IsSessionBeadOrRepairable` candidate
  filtering. So the read path persists repairs at multiple steps and the repaired
  type affects which candidates are even considered. Returning `repair-needed`
  instead must be proven not to change the *selected* candidate at any step, and
  the contract must enumerate every inherited repair call site rather than
  treating it as one helper.
- **[Major] The "read-only query first adopter" shares the exact resolver
  functions with mail, extmsg, assignee, and mutating commands.** Verified call
  sites of `resolveSessionTargetIDWithContext` and its wrappers
  (`resolveSessionIDWithConfig`, `resolveSessionIDAllowClosedWithConfig`,
  `resolveSessionIDMaterializingNamedWithContext`): query endpoints
  (`handler_session_transcript.go:36`, `huma_handlers_sessions_stream.go:24`,
  `huma_handlers_sessions_query.go:119,148`), **mutating** commands
  (`huma_handlers_sessions_command.go` — 11 sites incl. close/wake/suspend),
  assignee (`handler_beads.go:56,72,74`), mail (`handler_mail.go:90,140`), extmsg
  (`handler_extmsg.go`), and `worker_operation_watch.go:28`. The per-surface
  matrix correctly holds the non-query surfaces at "characterization only," but
  the design's framing ("first adopter = query only") implies an isolation the
  code does not provide: all these callers route through the *same* function with
  different `opts` (materialize vs not, allowClosed vs not). Extracting a
  read-only classifier from this shared resolver must be proven byte-for-byte
  behavior-identical for the non-adopter callers too, or characterize all of them
  before touching the shared path. This is precisely where my red flag #1
  (classifier output reinterpreted differently by caller) would bite.
- **[Minor→Major] Config-orphan rejection must preserve a *dual* `errors.Is`
  match or 404 silently becomes 500.** `apiSessionTargetRejectedByConfig` returns
  `apiSessionTargetNotFoundError{rejectedByConfig:true}` whose `Is()`
  (session_resolution.go:44-46) matches *both* `session.ErrSessionNotFound` (so
  it renders 404) *and* the distinguishing `errSessionTargetRejectedByConfig`
  marker. The typed schema models this as `result_kind=rejected` (separate from
  `not-found`). If the `rejected` adapter wraps an error that no longer matches
  `ErrSessionNotFound`, the current 404 becomes a 500. The `terminal_error`
  requirement covers this in principle but does not name the dual-match
  obligation for the config-orphan case.

**Missing evidence:**
- No enumeration of which current callers depend on the persisted side effect of
  `RepairEmptyType` (vs the in-memory patch). Selection parity is necessary but
  not sufficient if a downstream consumer lists by `Type=session` and relies on
  the read path having persisted the repair.
- No statement of how `apiSessionResolveOptions` (`materialize`, `allowClosed`)
  map onto the typed result kinds, even though those options are the only thing
  distinguishing the shared resolver's behavior across the 7+ surfaces.
- The prose step 5 calls config-orphan "the existing config-orphan not-found
  case" while the typed schema calls it `rejected`; the document does not state
  which name is authoritative for adapters.

**Required changes:**
1. Enumerate every `RepairEmptyType` call site the read-only classifier inherits
   (`ResolveSessionBeadByExactID:56`, `listSessionBeadsByMetadata:142`) and
   require a parity fixture proving the *selected* candidate is unchanged at each
   precedence step when the classifier returns `repair-needed` instead of
   persisting.
2. State explicitly that `resolveSessionTargetIDWithContext` is shared by query,
   mutating-command, mail, extmsg, assignee, and worker-watch callers, and
   require Slice 1 to either keep the shared function behavior-identical for all
   non-adopter callers or add characterization tests for every current caller
   before extraction — not only the read-only query endpoints.
3. Add a `terminal_error` rule for the `rejected` (config-orphan) result: its
   wrapped error must preserve the dual `errors.Is` match
   (`ErrSessionNotFound` → 404 *and* the rejected marker), with a no-delta test
   asserting the status stays 404.
4. Reconcile the naming: state that config-orphan is a `rejected` result kind
   that *renders* as 404, and make the prose step 5 use that vocabulary so a
   caller cannot collapse `rejected` into `not-found` (or vice versa).

**Questions:**
- For the shared resolver, is the intent that Slice 1 wraps the existing function
  unchanged and only adds a typed view, or that it re-implements precedence
  inside `internal/session`? The parity surface is very different between the two.
- Do any current callers branch on `errors.Is(err, errSessionTargetRejectedByConfig)`
  (not just the 404 render)? If so, the `rejected` adapter must keep that marker.
- Is persisting the empty-type repair part of the product contract (some consumer
  depends on it) or an incidental side effect safe to defer to an audited repair
  command?
