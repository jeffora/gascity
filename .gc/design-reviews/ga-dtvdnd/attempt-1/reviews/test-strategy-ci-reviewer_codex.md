# Mira Acharya - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The requirements artifact now maps every acceptance criterion to a verification class, and the major migration invariants are expressed as validators, command checks, golden outputs, active execution evidence, release gates, or manual audit gates rather than reviewer-only policy.
- AC8 is strong for role-neutrality testing: it covers Go code, Core assets, generated/materialized metadata, rendered output, route and notification targets, allowlists, positive controls, negative controls, and rot-guard behavior.
- AC13 directly addresses the legacy-test transfer risk by requiring a frozen baseline, active `go test -json` evidence, negative/regression controls, and failure on skipped, empty, or no-op witnesses.

**Critical risks:**
- [Major] AC2 and AC9 do not define the minimum controller-only operation denominator. "Normal SDK infrastructure operations," "controller-only operation," and "no-executor controller-only test" can be satisfied by a shallow config-load check unless the requirements name the operations that must work with Gastown roles absent and the Core maintenance executor renamed, omitted, undefined, or disabled.
- [Major] AC7 does not require per-witness negative or mutation controls for behavior preservation. It requires executable checks that the external Gastown pack loads, resolves, renders, triggers, routes, notifies, runs scripts, and exercises recovery paths, but a witness could still be too broad to fail when a moved prompt variable, route, notification target, script side effect, formula trigger, or recovery branch is removed.
- [Minor] AC17 requires `acceptance-proof-matrix.yaml`, a named validation command, and at least one executable gate, but the requirements do not define the minimum command/gate shape that would be acceptable. That is probably fine for requirements, but it leaves a residual risk that the implementation plan names a weak ad hoc command instead of a stable local or CI entry point.

**Missing evidence:**
- A concrete controller-only/no-executor operation matrix for AC2 and AC9, covering the actual infrastructure surfaces that must remain deterministic without any Gastown role.
- Per-witness negative-control expectations for AC7 behavior-preservation rows, especially prompt variables, routes, notification targets, script side effects, formula/order triggers, and failure/recovery branches.
- A minimum acceptable acceptance-proof gate shape for AC17, such as a stable validator command plus the local/pre-commit/CI surface that must invoke it before decomposition or implementation approval.

**Required changes:**
- Amend AC2 and AC9 to list the minimum controller-owned operations that must pass with only Core present and with the optional Core maintenance executor renamed, omitted, undefined, or disabled. At minimum, the requirements should state whether this includes config resolution, bead/task operations, hook/dispatch/sling setup, formula or order expansion, doctor/import-state diagnostics, health/reconciler diagnostics, and city-state mutation safeguards.
- Amend AC7 to require negative or mutation controls for each behavior-row or call-site witness, with active execution evidence proving the witness is not skipped, empty, no-op, or only checking file presence.
- Tighten AC17 by defining the minimum acceptance-proof gate shape: the proof matrix must expose a stable validation command and that command must be invoked by a named class of gate before decomposition or implementation approval.

**Questions:**
- Which controller-only operations are non-negotiable for this migration: config resolution, bead CRUD/query/hook, dispatch/sling setup, formula/order expansion, doctor/import-state diagnostics, health/reconciler diagnostics, city-state mutation safeguards, or all of those?
- Should the requirements mandate an exact proof-matrix command name now, or is it sufficient to require the support artifact to define the command before decomposition?
