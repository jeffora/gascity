# Elena Marchetti - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The current design now puts a hard non-mutating gate in front of behavior-moving work: Slice 0 is limited to inventories, schemas, validators, negative fixtures, and proof, and later slices must cite exact artifact rows before moving behavior (`internal/session/DESIGN.md:135`).
- The mutation ledger is broad enough in intent for this lane. It names exact path/function rows, session-bead targeting proof, literal keys or dynamic-key sources, top-level `Type`/`Status`/close/reopen/create-time fields, generic bridge reachability, exception expiry, and persistence-error handling (`internal/session/DESIGN.md:295`).
- The guard scope now covers the important escape hatches: `beads.UpdateOpts.Metadata`, patch constructors, dynamic metadata maps, create-time metadata, `Store.Close`, `Store.Reopen`, `Store.Update`, patch-returning helpers, repair helpers, runtime identity backfills, and generic API/CLI bead update routes (`internal/session/DESIGN.md:319`, `internal/session/DESIGN.md:334`).
- `RepairEmptyType` is finally treated as a mutating repair, not as harmless read cleanup. The design requires read-only classifiers to return `repair-needed` and gives actual repair writes an audited owner that propagates persistence errors (`internal/session/DESIGN.md:246`, `internal/session/DESIGN.md:351`).
- Worker-boundary coexistence is explicit enough to avoid accidental bypass precedent: API manager construction and session-resolution direct-create paths are exceptions, while new lifecycle API/CLI mutations default to `worker.Handle` or an exact expiring exception row (`internal/session/DESIGN.md:398`).

**Critical risks:**
- [Major] The approval still depends on Slice 0 being truly source-complete. The current tree has several broad production mutation bridges that can touch session beads: Huma bead update/close/reopen/delete routes call `store.Update`, `store.Close`, and `store.Reopen` with arbitrary IDs and metadata (`internal/api/huma_handlers_beads.go:357`, `internal/api/huma_handlers_beads.go:448`), and the bd-store bridge exposes create/update/close/reopen/set-metadata operations (`cmd/gc/cmd_bd_store_bridge.go:150`, `cmd/gc/cmd_bd_store_bridge.go:174`). The design covers this category, but Slice 0 must prove the scanner includes these bridges and does not let them hide behind broad API, CLI, or store-level exemptions.
- [Major] The phrase "new external" in the shrink-only guard could be implemented too narrowly. Existing external writers and patch applications are the real risk during migration: `cmd/gc/session_wake.go` applies `PreWakePatch` outside `internal/session`, `cmd/gc/session_beads.go` applies retirement patches and top-level status updates, and `internal/api/session_resolution.go` applies `RetireNamedSessionPatch` directly (`cmd/gc/session_wake.go:60`, `cmd/gc/session_beads.go:444`, `internal/api/session_resolution.go:171`). The guard needs a committed baseline and must fail on active exception growth, expired rows, missing retirement proof, and edits that broaden existing rows, not only on newly introduced call sites.
- [Minor] Runtime bridge denial is named, but the exact user-visible failure contract for bridge rejection is deferred to later artifacts. Generic Huma and CLI bridge attempts to mutate session-owned key families should have explicit response/exit-code parity rows, or implementers may reject correctly but break API/CLI compatibility.

**Missing evidence:**
- The actual `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, guard allowlist, and negative fixtures do not exist in this design artifact yet. That is acceptable for a design approval only because Slice 0 is the first executable work and is non-mutating.
- The scanner source set is not shown. Slice 0 needs an explicit inclusion rule for production non-test, non-generated Go under `cmd/` and `internal/`, plus named exclusions for generated files and fixtures.
- There is no proof yet that bridge denial and static scanning share the same session-owned key-family list. The design requires it, but the required fixture set should include dynamic metadata keys, top-level `Type`/`Status`, create-time metadata, close/reopen, and patch-map application.
- There is no proof yet that all `RepairEmptyType` callers are inventoried, including resolver helpers outside the first API query classifier path.

**Required changes:**
- No additional design-text change is required for this persona before design approval. The document now states the right mutation-boundary contract.
- Before closing Slice 0, add exact inventory and negative-fixture rows for the Huma bead bridge, the bd-store CLI bridge, `RepairEmptyType`, external `MetadataPatch` application, direct `SetMetadata*`, `beads.UpdateOpts.Metadata`, create-time metadata, close/reopen/status/type mutation, and direct `session.Manager.Create*` bypasses.
- Make shrink-only enforcement concrete in `SLICE0_CONTRACT.yaml`: fixed baseline source, no active exception growth, expiry failures, retirement-condition failures, missing source-hit failures, and same-change updates when a touched production writer changes.
- Require bridge-denial fixtures to assert both domain safety and wire/output parity: Huma status/body shape, CLI stdout/stderr/exit code, and generated-client/dashboard obligations when applicable.

**Questions:**
- For generic bead bridges, should a session-owned key-family write be rejected with a stable 4xx/CLI error, or should any operation be rerouted through a session-owned command when one exists?
- What is the authoritative shrink-only baseline: the committed Slice 0 inventory, the merge base against the target branch, or both?
- Are low-level store implementations exempt as infrastructure only, with guard responsibility placed on bridge/caller rows, or should the ledger also include store-method rows for every externally reachable mutation primitive?
