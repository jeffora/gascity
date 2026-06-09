# Ravi Krishnamurthy — DeepSeek V4 Flash (Independent Review, Attempt 6)

**Verdict:** pass

**Persona:** migration sequencing, legacy-new coexistence, rollback slices, worker-boundary collision, cross-document consistency.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 6, matching `.gc/design-reviews/ga-unpr2y/attempt-6/design-before.md` with Attempt 6 revisions), `internal/session/REQUIREMENTS.md`, `cmd/gc/cmd_session_wake.go`, `cmd/gc/session_lifecycle_parallel.go`, `internal/api/session_resolution.go`, and attempt-6 cross-persona reviews (Amara Diallo, Amara Osei).

---

## Top Strengths

- **Structured Caller Routing & Boundary Mapping (`DESIGN.md:264–281`)**: The newly added `Caller Routing And Command Construction Gate` is an exemplary, first-principles resolution to the risk of boundary collision. Rather than asserting "compatibility" in high-level prose, the design now explicitly maps each caller surface (production CLI, legacy API handlers, Huma API handlers, reconciler) to its required target end route and guard requirement. Specifying new static-guard exceptions (`W-025` through `W-028`) to track progress ensures that the transition is completely auditable and structured.
- **Durable Per-Key Owner Matrix & Rollback Contracts (`DESIGN.md:896–910`)**: Mandating that each implementation bead carry a per-key owner matrix and explicit rollback rule as its "rollback contract" completely closes the gap around split-brain metadata writes. This provides operational teeth to the "one writer per key during bake" rule, ensuring that rollback transitions are fully documented and verified.
- **Deprecated and Fenced Package-Level Mutators**: Resolving the bypass risks from Attempt 5 by adding `W-023` (for `session.WakeSession`), `W-024` (for `RepairEmptyType`), and `W-027`/`W-028` (for `Manager` direct calls and patch builders) ensures that the AST-based static guard catches every possible bypass during the transition.
- **Strict Read-Only Target Classification (`DESIGN.md:170–174`)**: Decoupling `RepairEmptyType` from target classification is a major win for architectural coherence and concurrency safety. Keeping the classifier strictly read-only prevents race-prone side effects on read-only endpoints, satisfying TR-002.

---

## Remaining Risks & Recommendations (Minor)

### 1. Reconciler Churn-Clears for `session_key`
While `cmd/gc/session_reconcile.go` is now correctly listed as an inventory source path, the implementation team must ensure that the gate for Slice 3 (Runtime start) explicitly divides and retires the individual `session_key`, `instance_token`, and `continuation_reset_*` clears in the reconciler loop. Any untracked clear-path during the bake window could violate the "one writer per key" invariant.

### 2. Intra-File Revert Order for `session_resolution.go`
Because `internal/api/session_resolution.go` is touched by three separate slices (classification in Slice 1, materialization in Slice 3, and retirement in Slice 4), the implementation team must ensure that the AST guard allowlist for this file is shrunk progressively. Reverting Slice 4 must not accidentally re-expose or restore the Slice 3 direct materialization exception.

---

## Answers to Persona Questions

1. **How does the plan sequence this extraction with the in-flight worker-boundary migration on overlapping cmd/gc and internal/api call sites?**
   - The extraction is sequenced via the `Caller Routing And Command Construction Gate` mapping. The in-flight worker-boundary remains the canonical production boundary during transition. The API and CLI layers are systematically wrapped in adapters that route through the allowed boundaries (`worker.Handle` or explicit command factories) while the static guard uses shrink-only allowlist exceptions (`W-001` through `W-028`) to prevent regressions.
2. **During partial adoption, what prevents legacy patch-map callers and new command callers from split-brain writes to the same metadata fields?**
   - The combination of: (a) the per-key owner matrix which defines the single command slice owning each metadata family, (b) the shrink-only allowlist enforced by the build-failing static guard, and (c) the mandatory "rollback contract" attached to each implementation bead. Any un-inventoried direct write or double-write will immediately fail the build.
3. **Which single slice is independently shippable and revertible, and what proves it does not silently require the next slice?**
   - Slice 1 (Target Classification) is independently shippable and revertible because it operates as a strictly read-only candidate resolver, utilizing compatibility adapters over the old resolver oracle. Reverting requires only a single configuration/adapter toggle with zero metadata migration. Slice 0 (Transition reducer baseline) is also introduced as an independent baseline ensuring transition validation is established before extracting downstream commands.

---

## Questions for the Author

1. Will the AST static guard run as a pre-commit hook or as part of `make test` to ensure that developer-local violations are caught before code is pushed?
2. For Slice 3, does the team plan to wrap the reconciler's `session_key` clear-paths into a dedicated, locked sub-command or transaction to prevent concurrency races with in-flight start-prepares?
