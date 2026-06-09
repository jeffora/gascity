# Elena Marchetti - Codex

**Verdict:** block

Reviewed `internal/session/DESIGN.md`, which matches `.gc/design-reviews/ga-unpr2y/attempt-4/design-before.md`, plus `internal/session/REQUIREMENTS.md` and the scoped session instructions.

**Top strengths:**
- The design now gives `internal/session` clear ownership of lifecycle and identity mutation, including exact key families and dynamic-key treatment for session beads.
- The static guard direction is the right shape: AST/symbol based, shrink-only allowlist, and additive to the existing worker-boundary guard.
- The slice sequencing correctly says implementation is blocked when source proof, scenario mapping, or writer retirement conditions are stale.

**Critical risks:**
- [Blocker] The "Canonical Production Writer Inventory" is still not source-complete, so the proposed guard cannot be generated or reviewed from the design. Current production code outside `internal/session` still creates or mutates session-owned lifecycle/identity fields in call sites that are not named as inventory rows: `cmd/gc/adoption_barrier.go` builds `session_name`, `state`, `generation`, `continuation_epoch`, and `instance_token` metadata before `store.Create`; `cmd/gc/session_name_lookup.go` creates session beads and later sets `session_name`; `cmd/gc/cmd_stop.go` writes `sleep_reason`; `cmd/gc/session_reconcile.go` clears `last_woke_at`; `cmd/gc/cmd_wait.go` writes `wait_hold`, `sleep_intent`, and sometimes `sleep_reason`; `cmd/gc/session_beads.go` sets pool `session_name` after create. Some broad rows partially imply these areas, but the design itself requires exception rows to appear in the static guard allowlist with the same ID, owner slice, and expiry. Those concrete rows do not exist.
- [Major] Several exception rows are too broad to be guard contracts. `W-011` says `cmd/gc/soft_reload.go`, `cmd/gc/cmd_nudge.go`, and `cmd/gc/cmd_bd_store_bridge.go` "if they write owned keys"; `W-021` covers generic bridge writes as "unknown"; `W-009` buckets large reconciler files. A failing-build guard needs exact call sites, key sets or dynamic-key policy, bead-type discrimination proof, owner slice, and retirement condition. Otherwise the allowlist becomes a permanent bypass.
- [Major] The static guard spec does not explicitly require dataflow coverage for metadata variables used in session bead creation/update. For example `adoption_barrier.go` assembles owned keys in `meta` and later passes `Metadata: meta` to `store.Create`; `session_name_lookup.go` similarly builds metadata before create and sets the generated pool `session_name` after creation. A guard limited to same-expression map literals and helper calls will miss the highest-risk create paths unless variable metadata maps on session beads are either traced or always require explicit allowlist rows.

**Missing evidence:**
- A source-verified writer inventory produced from the current checkout, not just path-family buckets.
- The actual guard fixture shape: owned-key taxonomy, allowlist rows, expiry/retirement fields, and examples proving it catches `Metadata: meta` on `Type: sessionBeadType`.
- Ownership and retirement decisions for adoption barrier creation, pool session-name materialization, city-stop `sleep_reason`, wait-hold writes, and reconciler `last_woke_at` repair.
- Proof that broad non-session stores such as nudge, wait, mail, extmsg, convoy, and order writes are discriminated by bead type before being exempted from session-owned key rules.

**Required changes:**
- Add concrete inventory rows for every direct production session-owned writer found above, with file/function, exact owned keys or dynamic-key class, target bead type proof, owner slice, exception status, and retirement condition.
- Replace broad "if they write owned keys" and "unknown" rows with either concrete allowlist entries or a rule that the guard fails until a matching design row is added.
- Specify the guard must catch session bead creates/updates when metadata is assembled in variables, not only inline map literals.
- Make doctor/migration/repair exceptions structurally bounded: row ID, caller symbol, allowed keys, required trace/log evidence, and expiry. Ordinary production callers must not be able to self-label as repair by writing from a generic helper.

**Questions:**
- Is `runAdoptionBarrier` intended to remain a bounded migration/repair exception, or should it become part of the runtime-start/session-create command slice?
- Should `gc stop` marking `sleep_reason=city-stop` be modeled as a session hold/sleep command, a stop-specific session command, or a temporary repair exception?
- Which slice owns pool session-name finalization after bead creation: runtime start, target identity, or worker-boundary migration?
