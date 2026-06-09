# Amara Diallo - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now pins the first adopter to read-only API query lookup and explicitly leaves mutating API commands, CLI, mail, extmsg, assignee normalization, nudge, attach, pool resume, and sling as characterization-only until each surface has its own row.
- The first-adopter precedence matches the live resolver shape: reject `template:`, exact ID, configured named-session handling, ordinary live `session_name`/alias, config-orphan rejection, path alias by `Title`, and allow-closed fallback. This is the right way to avoid replacing current behavior with a generic target bus.
- The design separates raw classification from operation policy. `match_vectors`, `config_state`, diagnostics, and wrapped terminal errors give adapters enough room to preserve surface-specific behavior instead of reinterpreting classifier output differently per caller.

**Critical risks:**
- [Major] The typed candidate taxonomy omits ordinary config targets and historical aliases as explicit vector kinds. The requirements say ordinary config names and `template:<name>` are factory/config targets, not live session targets, and existing tests prove historical aliases are not lookup sources. If the classifier collapses ordinary config and historical-alias evidence into plain `not-found`, later mail, CLI, assignee, and extmsg adapters can no longer distinguish "known but forbidden" from "unknown", which is exactly where misrouting tends to happen.
- [Major] `repair-needed` is correctly introduced, but first-adopter behavior is still underspecified. Current resolver helpers call `RepairEmptyType` during exact-ID and metadata lookup, then continue selection even if the persistence update failed because the in-memory bead is patched. The design says parity tests must prove whether selection is preserved after separate repair or intentionally changed with approval, but it should require each query endpoint to state its actual adapter outcome for `repair-needed`.
- [Major] The proof requirement says "positive and negative fixtures for every precedence row", but it does not require a single same-token collision matrix that spans all candidate families. Target classification needs cross-family fixtures where the same token appears as direct bead ID, configured named identity, live `session_name`, live alias, path `Title`, closed `session_name`, closed alias, ordinary config name, `template:` target, and historical alias.
- [Minor] The design preserves Huma/legacy error rendering in prose, but the first-adopter proof should name the concrete route set and response fields. `writeResolveError` and `humaResolveError` must stay compatible not only by status code, but also by problem body/code prefix and generated-client-visible shape where applicable.

**Missing evidence:**
- Materialized `TARGET_CLASSIFICATION_CONTRACT.yaml` rows for the first adopter.
- A route inventory naming the exact read-only API query endpoints that delegate first, including allow-closed versus live-only behavior.
- Collision fixtures for configured named identity versus live alias/session_name, direct ID versus path alias, session_name versus path alias, closed session_name versus live template/config candidates, and historical alias versus current alias.
- Explicit negative fixtures for ordinary config names, `template:<name>`, template basenames, exact agent names, historical aliases, missing targets, ambiguous alias/session-name matches, config-orphan named sessions, and repairable empty-type session beads.
- No-delta assertions for legacy mux and Huma error rendering, including 409 ambiguity/configured-name conflict, 404 not-found/config-orphan rejection, and 500 store errors.

**Required changes:**
- Add explicit target kinds for `ordinary-config-target` and `historical-alias`, or add an explicit negative-candidate mechanism that records those facts without making them selectable. Do not let them disappear into generic `not-found` when a surface needs to reject a known-but-forbidden target.
- Require every delegated surface row to declare how it maps `selected`, `not-found`, `ambiguous`, `rejected`, `repair-needed`, and `store-error` to its current behavior. For the first adopter, state whether `repair-needed` runs a separate repair and then selects, returns a specific error, or changes behavior with owner approval.
- Make the first-adopter fixture set a same-token matrix, not independent one-off examples. The matrix should prove precedence and diagnostics when multiple candidate families exist for the same token.
- In `TARGET_CLASSIFICATION_CONTRACT.yaml`, keep raw match vectors and adapter-selected outcomes separate. The raw classifier can collect evidence, but only the surface adapter may choose the authoritative target or materialization rule.
- Add exact route and renderer proof rows for legacy mux and Huma query handlers, including status, error body, `errors.Is` compatibility where callers depend on it, and generated-client/dashboard obligations where visible.

**Questions:**
- Should the first API-query classifier inspect ordinary config targets only to report a rejected/known-negative candidate, or should that be deferred until CLI/mail/extmsg surfaces delegate?
- For repairable empty-type session beads, is the intended first-adopter behavior "repair command then select", "return repair-needed", or "select read-only without writing"? The design allows all three unless the endpoint row chooses.
- Are historical aliases intentionally never collected, or should the classifier record them as forbidden candidates for diagnostics while preserving the no-lookup rule?
