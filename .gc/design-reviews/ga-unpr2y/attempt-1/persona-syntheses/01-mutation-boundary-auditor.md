# Elena Marchetti

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker for Slice 0 closure] Shrink-only enforcement is still not a forcing function unless validators fail on expired ledger/exception rows, rows whose retirement condition is met while the writer remains, and exception growth beyond the recorded baseline. All sources identify this as the difference between a temporary migration bypass and a permanent one.
- [Major] The static guard strategy is underspecified. The design needs to say exactly what `TestSessionBoundaryGuard` matches, what source roots and exclusions it scans, and how it avoids over-matching unrelated bead metadata writes while still catching direct session-owned metadata, status, type, close, reopen, create-time metadata, and patch-map writes.
- [Major] Generic bead mutation bridges are a runtime escape hatch unless they reject session-owned key families or route them through session-owned commands. API/Huma bead update surfaces, CLI bd-store bridge paths, `beads.UpdateOpts.Metadata`, close/reopen/status/type mutations, and dynamic metadata maps must be inventoried and covered by negative fixtures.
- [Major] Exported patch-map APIs remain a durable mutation-boundary risk. External callers can still range over exported `*Patch` constructors or construct transparent `MetadataPatch` values, so later slices need a committed end state where patch application is session-owned, preferably by unexporting constructors or making the representation opaque.
- [Major] `RepairEmptyType` and read-like resolver/helper paths are still write-on-touch risks. The ledger must enumerate these callers, assign an audited repair owner and owning slice, and ensure read-only classifiers/adopters cannot silently repair or backfill session-owned fields.
- [Major] The backlog lacks concrete call-site ownership for known direct lifecycle and identity writers. The design needs a slice-level map for external patch applications, inline rollback batches, identity assignment writes, identity retirement, and repair callers so each mutation-moving slice has a verifiable completion condition.
- [Major] Transitional coexistence needs an atomic key-family rule. Once a key family is adopted by a fenced command path, legacy blind writers for that same key family must be disabled or routed in the same slice to avoid split ownership and clobbering.

**Disagreements:**
- Claude and Codex both verdict `approve-with-risks`; DeepSeek verdict `iterate` with one blocker. I treat the required Claude/Codex agreement as enough for a persona-level `approve-with-risks`, but only because Slice 0 is explicitly non-mutating and must not close until the blocker-class validator and inventory gaps are fixed.
- Claude counts 14 `RepairEmptyType` call sites while DeepSeek counts 16. The exact count is unresolved; Slice 0 must produce the authoritative machine-readable inventory rather than relying on either review's manual count.
- Codex says no additional design-text change is required before design approval, while Claude and DeepSeek want explicit commitments for expiry enforcement, guard matching, and patch-map closure. I assess those commitments as required before closing Slice 0 and safer to record in the design or Slice 0 contract now.
- Claude emphasizes current external patch applications and unenforced expiry; Codex emphasizes generic API/CLI bridges and guard coverage; DeepSeek emphasizes concurrent legacy/new writers. These are compatible, not contradictory: they are different ways the same mutation boundary can leak.

**Missing evidence:**
- The actual `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, guard allowlist, and negative fixtures.
- Scanner source-set rules for production `cmd/` and `internal/` Go files, plus generated/test/fixture exclusions.
- A shared machine-readable session-owned key-family and top-level field list used by both static guards and runtime bridge denial.
- Negative fixtures for Huma bead update, bd-store bridge update, bd-store metadata writes, direct `UpdateOpts.Metadata`, create-time metadata, dynamic-key metadata writes, status/type changes, close/reopen, patch-constructor application, and repair helper calls.
- Proof that read-only classifiers and query adopters avoid mutating helpers such as `RepairEmptyType`, identity backfills, metadata-candidate helpers, and resolver paths that repair while resolving.
- Concrete slice ownership for external `ClosePatch` and `RetireNamedSessionPatch` applications, direct `session_name` assignment writers, inline wake rollback batches, lifecycle `SetMetadata*` writers, and all repair callers.
- CI details for expiry checks, retirement-condition checks, merge-base or baseline comparison, and active exception-count ratcheting.

**Required changes:**
- Add Slice 0 validator failures for expired ledger/exception rows, satisfied retirement conditions with remaining writers, missing source hits, active exception growth, and broadening an existing exception row without same-change retirement proof.
- Define `TestSessionBoundaryGuard` concretely: source roots, exclusions, manifest schemas, symbol/site matching rules, and how the key-family list is used as ownership metadata versus runtime bridge denial input.
- Fence generic mutation bridges with runtime rejection or session-command routing for session-owned key families and session bead top-level fields, with API/CLI parity fixtures for status codes, body shape, stdout/stderr, and exit codes where applicable.
- Ledger and assign owning slices for `RepairEmptyType`, side-effecting resolver/query helpers, external patch applications, direct lifecycle/identity metadata writers, identity assignment, inline rollback batches, create-time metadata, close/reopen, status, and type writes.
- Commit the patch-map end state: retire external patch application and make patch application session-owned through opaque/unexported APIs or an explicitly justified equivalent.
- Add an atomic key-family migration rule so a session-owned key family cannot be concurrently written by both legacy blind writers and new fenced command appliers.
- Give each mutation-moving backlog slice file- or symbol-level completion criteria tied to the Slice 0 inventory rows it retires or routes.
