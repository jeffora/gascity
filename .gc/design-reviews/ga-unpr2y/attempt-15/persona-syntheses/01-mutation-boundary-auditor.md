# Elena Marchetti

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Generic mutation bridges remain an unbounded runtime escape hatch. A path-level exception for API/CLI bead update surfaces cannot constrain dynamic metadata keys, target bead IDs, or top-level `Type`/`Status` writes. These bridges need runtime rejection for session-owned key families or routing through session-owned command APIs, with negative fixtures for lifecycle and identity keys.
- [Major] The guard and ledger scope are still incomplete. The design must cover `beads.UpdateOpts.Metadata`, create-time metadata for session beads or repairable candidates, direct `Close`/`Reopen`, top-level status/type writes, helper-returned metadata maps, exported repair helpers, patch constructors, and dynamic-key sources. Without this, callers can bypass `SetMetadata*` scanning by using adjacent bead-store mutation APIs.
- [Major] Shrink-only and expiry rules do not yet force shrinkage. The reviews agree that exception rows, allowlists, and ledger entries need CI-visible constraints: expired rows must fail, rows whose retirement condition is met must fail while the writer remains, and the active exception set must not grow relative to the merge-base or recorded baseline.
- [Major] Read-like flows still have mutating repair behavior. `session.RepairEmptyType` and side-effecting resolver/helper paths must be named in the ledger, assigned to an owning slice, and fenced so read-only classifier adoption cannot call helpers that repair or backfill session-owned fields.
- [Major] Transparent exported patch maps remain a durable boundary risk. Existing external applications of `MetadataPatch` constructors should have a committed retirement path; the end state should make patch application session-owned, preferably by unexporting constructors or making the patch representation opaque after the migration slices absorb call sites.
- [Major] Transitional coexistence can create multi-writer races. Once a key family is adopted by a fenced command applier, legacy direct writers for that same key family must be disabled atomically across production surfaces to avoid blind-write clobbering.
- [Major] The static analysis strategy is underspecified. The Slice 0 guard needs schema-validated inventory files and a type-directed scanner, or an equally concrete strategy, that distinguishes session-bead writes from unrelated bead-store writes without relying on brittle path or string matching.

**Disagreements:**
- Claude and DeepSeek rated the current design `approve-with-risks`; Codex rated it `block`. The disagreement is mainly about whether Slice 0's planned inventory and validators are enough to approve the design now. I assess Codex's block as controlling because generic bead mutation bridges can mutate session-owned fields at runtime without adding a new guarded call site, so a static shrink-only guard can pass while the ownership boundary remains porous.
- Claude emphasizes the existing external patch-application baseline and unenforced expiry as the dominant risk. Codex emphasizes dynamic bridges, omitted `UpdateOpts.Metadata` coverage, and side-effecting read helpers. These are compatible findings: together they show both current-source escape hatches and runtime mutation surfaces need to be closed or explicitly fenced.
- DeepSeek views deferring exact call-site mapping to Slice 0 as a robust stale-design mitigation. Claude notes the current backlog still lacks reviewable call-site-to-slice mappings. I assess the deferral as acceptable only if Slice 0 is a hard prerequisite and produces machine-checkable inventory IDs that later slices must cite before behavior-moving work starts.

**Missing evidence:**
- Exact scanner scope, including production roots, test/generated exclusions, and how `cmd/` bridges versus `internal/` session-owned code are classified.
- A machine-readable session-owned key-family and top-level field list shared by the static guard and runtime bridge denial logic.
- Negative fixtures for API bead update, bd-store bridge update, bd-store bridge metadata writes, direct `UpdateOpts.Metadata`, `CreateOpts.Metadata`, lifecycle status/type writes, and dynamic-key attempts against session-owned key families.
- Proof that read-only target classifiers and query adopters avoid `ResolveSessionID*`, `ListAllSessionBeads`, metadata-candidate helpers, `RepairEmptyType`, or any other helper that repairs or backfills during lookup.
- CI enforcement details for shrink-only behavior, expiry dates, retirement conditions, and exception-count baselines.
- Concrete schema and validation rules for `SESSION_BOUNDARY_SYMBOLS.yaml`, `SCENARIO_PARITY.yaml`, and `WORKER_BOUNDARY_EXCEPTIONS.yaml`.
- A slice-level ownership map for known direct lifecycle/identity writers, external patch applications, `RepairEmptyType` callers, inline rollback batches, and identity assignment writes.

**Required changes:**
- Add runtime bridge fencing or session-command routing for generic bead mutation surfaces. Session-owned metadata key families and session bead `Type`/`Status` must not be allowlisted by path alone.
- Expand the mutation ledger and guard contract to cover all bead-store mutation forms that can carry session-owned data, including direct `UpdateOpts.Metadata`, create-time metadata, status/type/close/reopen mutations, helper-returned metadata maps, exported patch constructors, and exported repair helpers.
- Make exception expiry and retirement conditions build-enforced, and add CI merge-base or baseline comparison so shrink-only cannot be bypassed by simply adding or extending exception rows.
- Ledger and assign ownership for `RepairEmptyType` and all side-effecting resolver/helper call paths; add a guard rule that read-only adopters fail CI if they call mutating repair/backfill helpers.
- Commit the end state for patch maps: retire external patch-apply sites in the migration slices and make patch application session-owned, either through opaque/unexported patch APIs or an explicitly justified equivalent.
- Add atomic key-family transition rules so a key family cannot be concurrently written by both legacy blind writers and new fenced command appliers.
- Define the Slice 0 linter strategy and inventory schemas concretely enough that reviewers can tell how session-bead writes are distinguished from unrelated bead-store writes and how dynamic keys are constrained.
