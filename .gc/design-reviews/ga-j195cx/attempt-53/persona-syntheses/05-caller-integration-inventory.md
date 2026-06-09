# Yuki Patel

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] Accepted-artifact continuation across host downgrade is not yet a crisp contract. Claude flags that `[daemon] formula_v2 = false` appears to allow artifact-stamped graph roots to continue graph-specific writes, while operators may expect graph behavior to stop. Codex identifies the implementer-level conflict: `ValidateAcceptedArtifact` can fail on host-capability disagreement even though same-identity in-flight continuation is supposed to remain valid after downgrade.
- [Blocker] The design needs an operator-visible recovery and diagnostics surface for in-flight graph workflows. Claude requires plain-language semantics for new compiles, artifact-stamped roots, legacy-only roots, fanout, retry-eval, scope-check, workflow-finalize, and convergence while the flag is disabled. Codex requires the `gc formula repair-root-artifact` path to be scheduled with ownership, docs, dry-run JSON, idempotency tests, zero-write failure tests, and behavior for missing source, downgraded host, unsupported artifact schema, and already-stamped roots.
- [Blocker] Caller migration coverage is not mechanically enforceable enough. Claude shows that generic rows hide duplicate raw filters in `internal/api/orders_feed.go`, `internal/sling/sling.go`, `internal/sling/sling_attachment.go`, and `internal/graphroute/graphroute.go`; the migration table and grep-derived caller manifest need one row per filter or predicate site, not one row per file.
- [Major] Workflow-root predicate normalization can silently change persisted-data behavior. Claude identifies divergent current semantics across `sourceworkflow`, `sling`, `sling_attachment`, and `graphroute` around exact match versus trim/case-fold and OR versus AND logic. The shared predicates must state which destination predicate each caller uses, and the rollout needs either a production-store scan for non-byte-exact values or explicit legacy fallback semantics.
- [Major] Host-capability provenance is specified as load-bearing but not fully constructible. Codex notes that `HostCapabilities` carries `SourceKind` and `ConfigGeneration`, but the proposed public constructor only accepts an enabled boolean and source string. Every config edge needs an implementable adapter that preserves omitted default versus explicit false, deprecated promotion, and per-tick generation.
- [Major] Migration cleanup is implied rather than scheduled. Claude requires a phase-3 deletion list and name-based CI guard for legacy helpers and globals such as `declaresGraphV2Contract`, `isGraphWorkflow`, `requiresExplicitGraphContract`, `metadataRequiresGraphContract`, `formulaV2Enabled`, `IsFormulaV2Enabled`, and `SetFormulaV2Enabled`, with similar treatment for dead convergence formula code if confirmed unused.
- [Major] Fanout and convergence appear to use different host-capability lifetimes. Claude flags that fanout reuses the parent operation snapshot while convergence can revalidate against current host when identity changes. The design must align the rule or explain the asymmetry and give operators remediation for stranded convergence.
- [Major] The external `bd` story is incomplete. Claude notes that `bd` direct users may still run the frozen beads compiler against requires-only formulas, so the compatibility matrix must state whether legacy `bd` consumers must dual-declare `contract = "graph.v2"` permanently or whether upstream beads will learn `[requires]`.
- [Blocker] Stale documentation can preserve the old fallback model. Claude specifically calls out `engdocs/proposals/formula-migration.md`, which still points readers at `GC_NATIVE_FORMULA=false -> Store.MolCook`; the design should require superseding or updating it in the same PR stack and include it in stale-guidance checks.

**Disagreements:**
- There is no verdict disagreement: both source reviews return `approve-with-risks`, and this synthesis adopts that verdict.
- Claude treats several caller-inventory and operator-semantics gaps as blocker-level risks; Codex labels its overlapping concerns major. My assessment is `approve-with-risks` rather than `block` because the runtime direction is accepted by both reviewers, but the overlapping host-downgrade, repair, and caller-manifest issues are required fixes before the migration plan is relied on.
- Claude emphasizes stale docs, external `bd` direct-user compatibility, predicate normalization, duplicate filter enumeration, and legacy helper deletion. Codex emphasizes API shape for host-capability provenance, accepted-artifact validation modes, and scheduling the repair command. These are complementary gaps, not conflicting recommendations.
- Claude questions whether convergence retry across downgrade should fail closed when runtime vars change; Codex leaves the implementation shape open between an `ArtifactReuseIntent` and write intents carrying both current-host and accepted-host capability. The design needs to choose one explicit model.

**Missing evidence:**
- No Gemini review artifact was present for this persona.
- No before/after table maps each current workflow-root or graph-root predicate to its exact replacement semantics.
- No fixture covers the full in-flight matrix for legacy-only roots and artifact-stamped roots across flag-on, flag-off, mid-fanout, and mid-convergence states.
- No accepted-artifact validation fixture proves that a same-identity persisted artifact can continue graph-specific writes after host downgrade while identity changes recompile against current host and fail closed when unsupported.
- No concrete config-edge adapter signature shows how `SourceKind` and `ConfigGeneration` are derived from `city.toml`, API requests, order scan ticks, convergence retries, and controller ticks.
- No operator-visible artifact or command output is specified for listing graph workflow roots that continue writing while `formula_v2 = false`.
- No evidence states whether `internal/convergence/formula.go:ValidateForConvergence` is dead code to remove or a still-needed migration target.
- No commitment states whether upstream `bd` will gain `[requires]` support or whether dual declaration is permanent for packs consumed by legacy `bd`.

**Required changes:**
- Split accepted-artifact validation into explicit current-host validation for new or changed compiles and persisted-artifact reuse for existing roots whose formula identity, vars/options hash, content hash, and artifact schema still match. Add tests for downgrade reuse, identity changes, and zero-write fail-closed behavior.
- State the operator-visible meaning of `[daemon] formula_v2 = false`: new graph compiles fail closed; artifact-stamped roots may continue only writes authorized by the persisted artifact; legacy-only graph roots without artifacts fail closed for graph-specific writes; and a CLI/API/dashboard surface lists affected in-flight roots.
- Add `gc formula repair-root-artifact` to the rollout matrix with owner package, command contract, generated docs/help, dry-run JSON schema, idempotency tests, failure diagnostics, and behavior for missing source, downgraded host, unsupported artifact schema, and already-stamped roots.
- Replace or supplement `HostCapabilitiesFromFormulaV2(enabled bool, source string)` with an edge adapter that accepts or derives `SourceKind` and `ConfigGeneration`, and require every production host-capability entry point to use it.
- Expand the caller inventory and executable migration table to enumerate each duplicate raw filter or predicate site by file and line, including the duplicate `orders_feed.go` filters and post-filters, sling predicates, sling attachment checks, and graphroute checks. Require the grep-derived caller manifest to track each occurrence.
- Document persisted metadata normalization risk and add a preflight scan or migration fixture for non-byte-exact `gc.formula_contract` values before alias removal begins.
- Add a phase-3 deletion list and name-based CI guard for legacy formula-v2 helper identifiers and any confirmed-dead convergence formula entry points.
- Resolve the fanout versus convergence host-capability lifetime rule by aligning behavior or documenting the deliberate asymmetry with operator remediation.
- Supersede or update `engdocs/proposals/formula-migration.md` in the same PR stack, remove stale `GC_NATIVE_FORMULA=false -> Store.MolCook` runtime rollback guidance, and include the path in stale-guidance CI checks.
- Add `bd` direct-user rows to the compatibility and external-support matrices, stating whether upstream beads will support `[requires]` or whether legacy `bd` consumers must keep dual declarations.
