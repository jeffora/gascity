# Amara Diallo - DeepSeek V4 Flash

**Verdict:** approve-with-risks

**Top strengths:**
- **Exquisite Precedence Specification:** Step 1-8 resolver precedence in the contract perfectly captures the complex, real-world resolver sequence from `internal/api/session_resolution.go`, including critical deconfigured config-orphan rejection (Step 5) and path-alias Title matching with deterministic newest-wins tiebreaking (Step 6).
- **Separated, Highly-Expressive Schema:** Replacing the flat candidate `kind` with a multi-dimensional schema containing `match_vectors[]`, `bead_state`, and `config_state` is a triumph. It beautifully preserves multi-dimensional matches (e.g., dual alias/session-name) without conflating token match facts with bead statuses or leaking policy into the classifier.
- **Strict Side-Effect-Free Isolation:** Restricting the classifier to be completely read-only and side-effect-free (moving `RepairEmptyType` out of the read path to a deferred repair command via the `repair-needed` result kind) perfectly guarantees that the classifier can be safely run without triggering unintended database side-effects.

**Critical risks:**
- **[Major] Precedence Rule Ambiguity for Dual-Match Demotion across Multiple Candidates:**
  While the design properly specifies the `match_vectors[]` schema, it is silent on the precise multi-bead demotion algorithm. Under the current `resolve.go` behavior, a dual bead (matching both `session_name` and `alias`) is demoted from alias matching *only if* another separate bead has a `session_name` match. If the classifier simply outputs raw candidates with their match vectors, the surface adapter must execute a cross-candidate pass to apply this demotion rule. The contract should explicitly clarify that the classifier outputs the raw, unfiltered match vectors, and the first-adopter API query adapter owns the responsibility of applying this demotion logic.
- **[Major] Semantic Mapping of the `repair-needed` Result Kind to Read-Path Success:**
  The `result_kind` enum defines `repair-needed` as a top-level kind alongside `selected`. For the first-adopter API query surface, if the classifier returns `repair-needed`, does this block target selection (returning a 404/Not Found) or does it select the bead successfully?
  To avoid a severe behavioral regression, the design must explicitly state that the API query-side resolver must treat `repair-needed` as functionally equivalent to `selected` for target resolution purposes, with the DB repair write deferred to the separate audited repair command. If `repair-needed` is treated as a rejection or failure, query-side API requests targeting unrepaired empty-type session beads will fail.
- **[Minor] Tie-breaker Collision on Identical CreatedAt Timestamps:**
  Step 6 resolves duplicate path-aliases by selecting the newest bead by `CreatedAt`. While rare, if two concurrent active pool sessions are created within the exact same timestamp resolution, the tiebreaker is undefined. The design should specify a secondary, deterministic tiebreaker (such as alphabetical order of bead IDs) to guarantee 100% deterministic outputs.
- **[Minor] Historical Alias Query Exclusion across Mutating Surfaces:**
  Step 7 explicitly notes that historical alias metadata is not a lookup source. While this is correct for read-only query-side lookups, mutating surfaces (such as mail and extmsg) remain characterization-only and are excluded from this slice. There is a minor risk that once those surfaces are delegated in future slices, their need to resolve historical aliases might force retrofitting of the `match_vectors[]` schema or result in backward-compatibility issues.

**Missing evidence:**
- **Query Selection Logic for `repair-needed`:** The document lacks an explicit description or test citation of how the first-adopter API query adapter processes the `repair-needed` result kind (i.e., proving that it maps to successful selection and 200 OK while scheduling repair).
- **Secondary Tie-breaker Specification:** There is no discussion of the tie-breaking behavior for identical `CreatedAt` timestamps on duplicate path-aliases.

**Required changes:**
- **Clarify the `repair-needed` Selection Rule:** Under the `RepairEmptyType` section or the first-adopter precedence rules, explicitly add a statement: *"The first-adopter API query adapter must map a `repair-needed` result kind to successful target selection, matching the existing behavior where lookup succeeds while the repair command is asynchronously triggered."*
- **Add a Secondary Tie-breaker for Path-Aliases:** In step 6, specify: *"In the event of duplicate titles with identical `CreatedAt` timestamps, alphabetical order of bead IDs must be used as a deterministic secondary tiebreaker."*

**Questions:**
- Will the separate audited repair command be triggered synchronously or asynchronously upon detecting a `repair-needed` result kind, and what ensures that concurrent query lookups do not spawn duplicate repair commands before the write commits?
