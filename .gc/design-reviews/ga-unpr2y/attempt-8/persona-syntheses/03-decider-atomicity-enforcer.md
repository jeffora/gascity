# Takeshi Yamamoto

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The pure-decider boundary is contradicted by the current baseline code and is not guarded in Slice 0. Claude and DeepSeek both identify `ProjectLifecycle` falling back to `time.Now().UTC()` when `LifecycleInput.Now` is zero, while the design says deciders must receive time as an injected fact. Codex also flags this as a major risk. The guard must reject direct clock reads in decider code and the existing fallback needs an explicit migration or removal before new deciders inherit it.
- [Blocker] The command-applier contract is still mostly prose. All reviewers agree runtime-start, close, wake, drain, and identity-retirement command rows need machine-checkable preconditions, validation points, write primitives, fence markers, stale-success handling, partial-state matrices, provider side-effect ordering, repair authority, and race tests before production callers delegate to them.
- [Blocker] Runtime-start still contains an unsafe commit escape hatch. Claude found the normative "Commit runtime start" row still allows a "legacy no-token compatibility path", contradicting the design's own rule that no-token runtime identity backfills are compatibility repair only. This licenses an unfenced commit path over a successor attempt.
- [Blocker] Crash recovery for runtime-start is not keyed only on durable facts. Claude shows that "crash before provider start" and "provider start succeeds, commit fails" can be durably indistinguishable if runtime identity is only in process memory. Recovery must first observe/adopt/stop a runtime by a durably derivable attempt identity and only roll back on complete-missing observation.
- [Blocker] Close semantics are under-specified and divergent. DeepSeek identifies two different close paths (`CloseDetailed` and `closeBead`) with different side effects, and both Claude and Codex call for exact close/retire failure-order tests. Without a close failure-point matrix and unified command, provider stop success plus close persistence failure can leave open beads, stranded work, retained wake/hold overrides, or resurrected sessions.
- [Blocker] Lock-free `instance_token` backfills remain an unprotected race. DeepSeek flags current backfills in `internal/session/chat.go` that can generate competing tokens and blindly overwrite each other. The design says backfills need conditional coverage or repair matrices, but does not specify how they are locked, moved into a validated batch, or retired.
- [Major] Store primitive and batch atomicity proof is missing. Claude and Codex both note that the primary bd store is documented last-writer-wins and that `SetMetadataBatch` may apply sequentially or partially. Any lifecycle slice using these paths needs backend-specific proof or repair tests for every visible partial state.
- [Major] Some prevention wording overstates what blind writes can guarantee. Claude notes that the design should phrase stale-token safety as detect-and-converge repair, not prevention, and every fenced commit batch should restamp the fence token so mixed states remain attributable.
- [Major] Slice 0 artifacts do not exist yet. Codex treats this as acceptable only for approving Slice 0 implementation; DeepSeek treats it as a blocker for any later slice. The lane agrees no mutation-owning slice should proceed until Slice 0 files, guards, ledgers, and tests physically exist and pass.

**Disagreements:**
- Codex returns `approve-with-risks`, while Claude and DeepSeek return `block`. My assessment: the persona verdict is `block` because the risks are not only missing implementation tests; they are contradictions or gaps in the normative design rows that implementation beads will consume.
- DeepSeek calls the `time` import guard structurally impossible because deciders need `time.Time` types, while Claude frames the issue as rejecting wall-clock reads and Codex asks for a decider guard. My assessment: the guard should allow time types but reject direct clock calls such as `time.Now()` inside the pure decider file set.
- Claude focuses on the runtime-start commit/recovery loopholes; DeepSeek focuses on dual close paths and token backfills; Codex focuses on per-slice command ledgers and backend proof. These are complementary atomicity gaps and should all be resolved in the same command-applier contract framework.
- The reviewers differ on whether a conditional store primitive is expected. The common requirement is explicit: either prove a real conditional/atomic primitive for the exact backend path or standardize on tokened blind writes with deterministic repair and tests for every partial state.

**Missing evidence:**
- A machine-checkable command-applier ledger or equivalent artifact for runtime-start, close, wake, drain, identity retirement, and token backfill.
- Tests or proof for stale success, duplicate command, stale token, successor attempt, skipped event emission, provider-start success plus commit failure, provider-stop success plus close failure, and every visible partial metadata batch.
- A store-backend support matrix showing which runtime command paths rely on `Update`, `Close`, `SetMetadataBatch`, `Tx`, or a future conditional helper, and whether each is atomic, partial, or last-writer-wins.
- A durable mechanism for associating a prepared `instance_token` with a live provider runtime after the process that started it crashes.
- A decider-purity file set and guard that rejects store/config/runtime/event/work-query access and direct wall-clock reads while allowing typed time values.
- Evidence that all callers pass a non-zero `LifecycleInput.Now`, or a migration plan that makes zero `Now` invalid without breaking required compatibility.
- A close failure-point matrix covering crash before stop, stop failure, stop success plus close commit failure, stale close intent, stranded assigned work, wake/hold cleanup, named identity retirement, and event emission failure.

**Required changes:**
- Delete the "legacy no-token compatibility path" from the Commit runtime start row. Route no-token identity backfills only through audited repair/backfill flows with explicit preconditions and race tests.
- Rewrite runtime-start recovery using durable facts only. Require prepare-time facts to make the attempted runtime identity discoverable, then adopt-or-stop any matching live runtime before rollback; rollback only after complete-missing observation.
- Add a Slice 0 decider-purity guard and proof command, and remove or explicitly migrate the `ProjectLifecycle` zero-`Now` fallback before new command deciders depend on it.
- Materialize command-applier rows as a machine-checkable ledger or tests before each command slice delegates production callers. Prose in `DESIGN.md` is not enough.
- Create a close command design with one owner path, one failure-point matrix, and explicit classification of re-derivable versus operator-initiated closes. Non-re-derivable closes need a durable closing intent before provider stop.
- Unify or retire divergent close implementations so close always handles wait cancellation, wake/hold override cleanup, named identity retirement, bead close, work release, and repairable partial states consistently.
- Protect or eliminate lock-free `instance_token` backfills by moving them under the session mutation boundary, into a validated prepare/wake batch, or into an audited repair flow.
- Add backend primitive proof for each lifecycle command path. If `BdStore.SetMetadataBatch` remains in use, assume partial writes unless proven otherwise and test repair for every lifecycle-visible subset.
- Rephrase stale-token guarantees from prevention to detect-and-converge, and require fenced commit batches to reassert their attempt token or phase marker in the same batch.
