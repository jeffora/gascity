# Sarah Chen

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] API lifecycle routing is not specified at the level needed to preserve the projection contract. The reviews converge on the need for a per-handler inventory that distinguishes legacy API handlers, Huma handlers, `worker.Handle`, `session.Manager`, direct `session.*` package calls, and any approved session command factory. Without current and target routes per operation, implementation beads can move behavior while preserving neither response parity nor the worker-boundary migration.
- [Blocker] The proposed static guard does not cover the actual bypass classes. Claude and Codex both flag guard precision gaps, and DeepSeek gives concrete examples: manager methods such as `mgr.CloseDetailed`/`mgr.Suspend`, package-level mutators such as `session.WakeSession`, direct store writes, and patch-map extension. A guard that only scans raw store calls will miss several current API mutation paths.
- [Major] CLI parity proof is too weak for target-resolution and command-conversion work. Claude and Codex both identify existing CLI-output oracles that the design does not cite, including `cmd/gc/session_model_phase0_cli_surface_spec_test.go`, `cmd/gc/session_action_json_test.go`, and relevant `cmd_session_*` tests. Broad regex gates are useful smoke coverage, but they do not prove stdout, stderr, JSON shape, and exit-code compatibility for each touched command.
- [Major] API exception and command inventories are incomplete. Claude flags that several "root-documented" API exception labels are not actually documented in root `AGENTS.md`. Codex requires explicit classification for current Huma command handlers such as stop, kill, respond, rename, permission-mode, and template-override paths. DeepSeek adds that legacy and Huma handlers for suspend, close, and wake currently use different routing patterns.
- [Major] Read-side lifecycle/projection adoption evidence is overstated. Claude verifies that `SESSION-LIFE-008` is a small snippet denylist, not a general guard. DeepSeek identifies additional raw lifecycle reads in Huma wake, `handler_status.go`, and Huma query paths that are not assigned to an adoption slice.
- [Major] Async API command invariants are named but not enumerated. Claude and Codex both call out missing proof for 202 responses, `RequestID` echo, async-create rejection cases, late-result/timeout behavior, and event/result semantics before slices that touch async create or close.
- [Major] Named-session and identity mutation details need ownership before conversion. DeepSeek identifies `session_resolution.go` extending `RetireNamedSessionPatch` with `alias_history`, and the Huma close handler embedding an always-on named-session policy. These must be assigned to the session command boundary or documented as temporary guarded exceptions.

**Disagreements:**
- Claude and Codex verdicts are `approve-with-risks`; DeepSeek verdict is `block`. Assessment: choose `block` because all three sources identify the same routing, parity, and guard gaps, and DeepSeek demonstrates that the guard as written would miss real current bypass paths.
- Claude treats the worker-boundary design direction as compatible if proof is tightened, while DeepSeek says the API routing matrix is missing enough to block. Assessment: the direction can stay, but design approval should block until every API lifecycle handler has current and target routing rows.
- Codex focuses on `apiClient()` route behavior for mutable CLI commands, while Claude focuses on `worker.Handle` versus session command factory. Assessment: these are complementary boundaries. The design needs both a CLI remote/local route matrix and a local/API command-boundary decision.
- Claude describes `SESSION-LIFE-008` as a narrow proof claim that should be corrected; DeepSeek names additional unguarded read sites. Assessment: downgrade the current proof language and add the named sites to the inventory and guard plan.
- DeepSeek raises concrete source-level issues around `alias_history` and Huma close policy that Claude and Codex do not emphasize. Assessment: treat them as required inventory items for the slices touching identity materialization and close, not as separate architecture pivots.

**Missing evidence:**
- Per-handler API routing inventory for legacy and Huma lifecycle/materialization handlers, including current route, target route, response shape, side effects, rollback route, and retirement slice.
- Per-command CLI route matrix covering supervisor API via `apiClient()`, standalone controller API, `GC_NO_API`, non-loopback read-only API, controller-down fallback, API error/no-fallback, local worker-boundary fallback, stdout, stderr, JSON, and exit code.
- Source-verified classification for all Huma session command endpoints, including create, patch/presentation, permission mode, submit/message async, stop, kill, respond, suspend, close, wake, rename, delete/prune/repair, and template overrides.
- Root or scoped `AGENTS.md` migration-note updates for any API exception rows beyond `session_manager.go` and `session_resolution.go`.
- Concrete guard artifact with scan roots, allowlist schema, fixtures, CI/pre-commit integration, owner, and shrink-only retirement conditions.
- Guard coverage for manager method calls and package-level mutators from outside `internal/session`, especially external `session.WakeSession` callers.
- Named CLI-output oracles for Slice 1 and later command-conversion slices.
- Named async API oracles for 202 response behavior, `RequestID` echo, async-create 400 cases, late-result/timeout behavior, and event/result semantics.
- Read-side adoption inventory for Huma wake's closed-status check, `handler_status.go` lifecycle derivations, Huma query read sides, and Huma close's named-session policy metadata reads.
- Decision on whether `RetireNamedSessionPatch` owns `alias_history` clearing or caller-side patch extension remains temporarily allowed.

**Required changes:**
- Add a per-operation API/CLI routing matrix before behavior-moving slices are generated or approved. It must distinguish CLI `apiClient()` projection, local fallback, API legacy handlers, API Huma handlers, `worker.Handle`, session command factory, `session.Manager`, and direct package-function paths.
- Split broad API exception rows into endpoint-level inventory rows with exact status codes, problem-detail bodies, request ID behavior, async event behavior, response bodies, generated schema/client impact, side effects, current routing, target routing, rollback route, and retirement slice.
- Add the existing CLI oracle files to Slice 1 and command-conversion proof lists, and require each adopting CLI surface to name stdout, stderr, JSON, and exit-code expectations.
- Define the approved session command boundary for API and local CLI fallback, including methods, typed result structs, conflict/error types, and adapter mappings to current Huma problem details and CLI output.
- Expand the static guard specification to cover direct store writes, manager method calls, package-level session mutators, direct command construction, dynamic metadata batches, and patch-map extension outside approved allowlists.
- Reconcile API exception labels with root `AGENTS.md`: either document W-016/W-025/W-029/W-030-style exceptions with owner and expiry in Slice 0, or relabel them as new exceptions requiring documentation before the guard lands.
- Re-describe `SESSION-LIFE-008` as the narrow snippet guard it is today, then add read-side guard or adoption obligations for the additional API/CLI raw lifecycle reads.
- Enumerate async session-command invariants and cite or require current oracles for slices touching async create, close, or command result delivery.
- Add external `session.WakeSession` callers, Huma wake's raw closed-status check, `handler_status.go` raw lifecycle derivations, Huma query read sides, `session_resolution.go` patch extension, and Huma close's named-session policy check to the slice inventory with ownership and proof requirements.
- Fold `alias_history` clearing into `RetireNamedSessionPatch`, or document the caller-side extension as a temporary guarded exception with a shrink-only allowlist row.
