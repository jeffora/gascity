# Ravi Krishnamurthy

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Major] The design is acceptable only as a non-mutating evidence-preflight, not as approval for behavior-moving slices. Claude and Codex both approve with risks because Slice 0 can inventory and validate without changing callers, materialization, repair behavior, or command APIs. DeepSeek agrees the strangler shape is strong but blocks broader decomposition because mutating-slice routing and coexistence protections remain unresolved.
- [Major] Split-brain close paths are not fully sequenced. Claude and DeepSeek both identify that the design scopes close alignment to production CLI callers while reconciler drained-close, failed-create/sweep closes, and parallel failed-create close still write raw close metadata. Slice 4 would add another writer unless the contract enumerates and converges all production close writers.
- [Major] The worker-boundary migration and session mutation-boundary migration overlap on shared files without an ordering rule. Claude calls out `internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and `cmd/gc/session_beads.go`; Codex requires concrete exception/retirement rows for wake and drain; DeepSeek says deferred routing lets future slices violate either `worker.Handle` boundaries or session layering. The design needs an authoritative sequence for shared files.
- [Major] Existing legacy raw writers can coexist indefinitely with new command-validated writers. Claude asks for per-field ownership transfer, Codex asks for legacy writer retirement conditions before wake/drain delegation, and DeepSeek escalates this to a cross-process split-brain risk because in-process locks do not coordinate the CLI and reconciler daemon. A shrink-only guard is not enough once a field is migrated behind a validated command.
- [Major] The "read-only" classifier path is not proven write-free. Claude and DeepSeek both identify `session.RepairEmptyType` writes on paths that could adopt target classification, while Codex flags side-effectful API resolver helpers that materialize, retire, reassign, kill runtimes, or update extmsg bindings. Slice 0 must quarantine these paths so read-only classification cannot inherit hidden writes.
- [Major] Per-slice rollback and independent shippability are underspecified. Claude says slices 4 and 6 collide on `cmd/gc/session_reconciler.go`; Codex asks for data-direction rollback rules; DeepSeek says only the read-only classifier slice is independently revertible unless the overlaps are resolved. The current rollback metadata is too declarative to prove safe partial adoption.
- [Major] Slice 0 artifact contracts need exact schemas, paths, owners, validators, and negative fixtures. Codex emphasizes that named artifacts such as `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and `COMMAND_APPLIERS.yaml` must be made concrete before later slices can cite them as evidence. Claude agrees the exception ledgers need owner and retirement rules.

**Disagreements:**
- Verdict severity differs. Claude and Codex return `approve-with-risks`; DeepSeek returns `block`. My assessment is `approve-with-risks` for the current Slice 0 evidence-preflight only. Any behavior-moving wake, drain, close, runtime-start, identity-retirement, or reconciler-fact slice should be treated as blocked until the required migration sequencing and writer-ownership changes land.
- DeepSeek requires a physical cross-process concurrency fence before partial adoption. Claude and Codex focus first on per-field exclusive ownership, retirement conditions, characterization tests, and rollback guards. My assessment is that the design must at least forbid remaining production raw writers for migrated fields; if coexistence is intentionally retained, then DeepSeek's physical conditional-write fence becomes mandatory.
- DeepSeek mandates all mutating session operations route through `worker.Handle` for production CLI code. Claude and Codex allow a store-level route if it is explicitly documented as a worker-boundary exception with retirement conditions and parity tests. My assessment is that the design must choose one route per operation before delegation rather than defer the decision to future slice work.
- Codex frames the artifact problem as the top risk; Claude and DeepSeek focus more on concrete close and worker-boundary collisions. My assessment is that both are the same migration failure mode: the artifacts are only useful if they force closure on the known collisions.

**Missing evidence:**
- Complete close-writer inventory and migration plan covering CLI `worker.Handle.CloseDetailed`, reconciler drained-close, failed-create/sweep closes, and parallel failed-create close.
- Cross-migration ordering for shared files, especially `internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and `cmd/gc/session_beads.go`.
- Per-field ownership-transfer rule proving that once a field moves behind a session command, remaining production raw writers for that field are retired or fenced.
- Canonical Slice 0 artifact locations, schema IDs, validator names, negative fixtures, proof commands, and workflow-close metadata.
- Route-inventory rows for side-effectful API resolution helpers and CLI wake, marking which are read-only classifier candidates and which are forbidden inputs to the first classifier.
- Per-slice rollback rows that specify files/imports/types allowed to remain, files/imports/types that must revert, and tests proving only one writer remains for each touched key family.
- Decision on whether wake and drain become `worker.Handle` methods or remain store-level session commands with explicit worker-boundary exceptions.
- Treatment of `session.RepairEmptyType` on read paths: whether it is quarantined, split into an audited repair command, or left as legacy behavior outside the read-only classifier slice.

**Required changes:**
- Update the close slice contract to enumerate all production close writers and require Slice 4 to characterize, converge, and retire raw close patches rather than only aligning production CLI callers.
- State the cross-migration sequence for every shared `cmd/gc` and `internal/api` file, including a "do not double-migrate" rule for files carrying both worker-boundary and mutation-boundary exceptions.
- Add a per-field ownership-transfer rule: once a field's writes move behind a session command, the guard must forbid remaining production raw writers for that field, or the design must specify a concrete cross-process conditional-write fence.
- Make `SLICE0_CONTRACT.yaml` or equivalent the canonical source for artifact paths, schemas, owners, validators, negative fixtures, proof commands, and close metadata; validators should fail when prose names an artifact absent from the contract.
- Add explicit route-inventory rows for `resolveConfiguredNamedSessionIDWithContext`, `materializeNamedSessionWithContext`, `retireContinuityIneligibleNamedSessionIdentifiers`, `reassignContinuityIneligibleNamedSessionState`, `resolveLiveSessionByPathAlias`, and CLI wake.
- Before any wake/drain behavior-moving bead, require the command-applier row to choose the boundary route, name legacy writer retirement conditions, list rollback files, and provide local/API parity tests.
- Add per-slice rollback data-direction rules and explicitly resolve the Slice 4 versus Slice 6 overlap on `cmd/gc/session_reconciler.go`.
- Resolve the read-only classifier repair paradox by quarantining silent `RepairEmptyType` writes or routing them through a separate audited repair command outside the read-only slice.
