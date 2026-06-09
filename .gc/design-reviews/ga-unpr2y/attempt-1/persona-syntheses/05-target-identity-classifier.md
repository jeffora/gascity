# Amara Diallo

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Major] The target-classification design now preserves the live resolver shape: `template:` rejection, exact ID, configured named-session handling before ordinary live names, live `session_name`/alias matching, config-orphan rejection, path-alias lookup, and allow-closed fallback.
- [Major] `repair-needed` is the main unresolved parity risk. Current exact-ID and metadata lookup paths call `RepairEmptyType` and then continue selection using an in-memory repaired bead, so each first-adopter endpoint must state whether repairable empty-type targets still select successfully, trigger separate repair, return a specific error, or change behavior through owner approval.
- [Major] The shared resolver is used by more than read-only query endpoints. Mutating commands, mail, extmsg, assignee normalization, worker watch, transcript, stream, pending, and query paths share resolver helpers with different options, so Slice 1 must either keep shared behavior identical for non-adopters or characterize those callers before extraction.
- [Major] Raw match facts must remain separate from surface policy. The classifier needs to preserve multiple match vectors for the same bead, bead state, config state, repair diagnostics, and terminal errors without letting callers infer policy from a flat candidate kind.
- [Major] Same-token collision coverage is required. Fixtures need to cover direct bead ID, configured named identity, live `session_name`, live alias, path `Title`, closed names/aliases, ordinary config names, `template:` targets, historical aliases, config-orphan names, and dual alias/session-name demotion.
- [Major] Config-orphan rejection must preserve current rendering and error identity. It is a `rejected` result that renders like not-found, including the `ErrSessionNotFound` 404 behavior and the distinguishing rejected-by-config marker where callers depend on `errors.Is`.
- [Minor] Path-alias behavior needs a complete compatibility row: `Title` matching, named-bead exclusion, accepted states, rejected states, newest-created selection, and a deterministic secondary tiebreaker if `CreatedAt` collides.

**Disagreements:**
- There is no verdict disagreement: Claude, Codex, and DeepSeek all return `approve-with-risks`.
- DeepSeek says the first API query adapter should map `repair-needed` to successful selection and defer repair; Codex says each endpoint must choose its adapter behavior; Claude requires parity proof that selected candidates do not change. My assessment: the first-adopter query rows must preserve current successful target resolution unless an owner-approved amendment explicitly changes it, and must state the repair mechanism.
- Codex wants ordinary config targets and historical aliases represented as explicit target kinds or negative candidates. Claude emphasizes that historical aliases must not become live targets. DeepSeek treats historical aliases mainly as a future mutating-surface risk. My assessment: the raw classifier may record forbidden evidence for diagnostics, but adapters must not select historical aliases or ordinary config targets unless that surface has an explicit parity row.
- DeepSeek calls out where the dual-match demotion algorithm should run; Claude verifies the current demotion rule; Codex asks for a same-token matrix. My assessment: the contract must name whether raw classifier output is unfiltered and adapter-owned, or whether classifier output already applies cross-candidate demotion, and tests must lock that choice.
- DeepSeek asks for a secondary path-alias tiebreaker; Claude and Codex do not elevate it. I treat it as a small but useful determinism requirement for Slice 1, not a design blocker.

**Missing evidence:**
- Materialized `TARGET_CLASSIFICATION_CONTRACT.yaml` rows for the first adopter.
- An exact route inventory for the first read-only API query delegation, including legacy mux and Huma routes, allow-closed versus live-only behavior, transcript, pending, stream, and generated-client effects.
- Enumeration of every inherited `RepairEmptyType` call site and which caller surfaces currently rely on the persisted repair versus the in-memory repaired bead.
- A mapping from `apiSessionResolveOptions` such as `materialize` and `allowClosed` to typed result kinds and adapter behavior.
- Collision fixtures for configured named identity versus live alias/session name, direct ID versus path alias, session name versus path alias, closed session name versus live template/config candidates, historical alias versus current alias, and dual alias/session-name demotion.
- Negative fixtures for ordinary config names, `template:<name>`, template basenames, exact agent names, historical aliases, missing targets, ambiguous aliases, config-orphan named sessions, and repairable empty-type session beads.
- No-delta proof for legacy mux and Huma rendering: status code, problem body shape, error code or prefix, `errors.Is` compatibility, and generated-client-visible schema.

**Required changes:**
- Enumerate inherited `RepairEmptyType` call sites and require parity fixtures for exact-ID lookup, session-name lookup, alias lookup, closed lookup where applicable, and candidate filtering; preserve successful selection for repairable session beads unless an owner-approved amendment changes it.
- State explicitly that the resolver helpers are shared by query, mutating-command, mail, extmsg, assignee, and worker-watch callers; require Slice 1 to keep non-adopter behavior identical or characterize those callers before shared-path extraction.
- Keep raw match vectors separate from adapter-selected outcomes in the contract. Represent ordinary config targets and historical aliases as explicit forbidden or diagnostic evidence rather than collapsing them into generic not-found when a surface needs to distinguish known-but-forbidden from unknown.
- Require a same-token collision matrix for the first adopter that proves precedence and diagnostics across all candidate families, including dual alias/session-name demotion.
- Add a `terminal_error` rule for config-orphan `rejected`: it must render as 404 and preserve the current `errors.Is` behavior for both `ErrSessionNotFound` and the rejected-by-config marker where used.
- Define path-alias compatibility fully, including `Title` source, named-bead exclusion, accepted and rejected states, most-recent `CreatedAt` choice, and deterministic tie handling for equal timestamps.
- Add exact route and renderer proof rows for legacy mux and Huma query handlers, including status, body shape, error compatibility, generated-client/dashboard obligations, and `repair-needed` behavior per endpoint.
