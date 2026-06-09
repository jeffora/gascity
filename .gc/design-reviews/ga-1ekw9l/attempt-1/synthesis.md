# Design Review Synthesis

## Overall Verdict: block

All persona syntheses are present and every persona reached a `block` verdict, so worst-verdict-wins yields a global `block`. The most universal issue is schema conformance: the reviewed `implementation-plan.md` has valid front matter, but the body is not the required Mayor implementation-plan shape because it uses `## Proposed Design`, omits `## Data And State`, uses `## Rollout`, and keeps long attempt-history strata in `## Summary`. Beyond that output-shape mismatch, reviewers found real design risks around loader authority, public-pack proof, doctor/runtime-state mutation safety, role neutrality, pack-boundary containment, bootstrap sequencing, and operator release gates.

## Consensus Strengths

- Reviewers generally agree the target architecture is directionally right: Core should be mandatory infrastructure owned by Gas City, while Gastown behavior moves to an explicit public pack pinned by immutable public-pack evidence.
- Multiple personas praised the intent to fail closed on required Core problems, eliminate direct loader bypasses, and make stale Maintenance/Gastown state diagnosable instead of silently active.
- The symbolic `core.maintenance_worker` binding, single retired-source classifier, generated role-surface/wording inventories, and fixture-only bootstrap end state were all viewed as promising mechanisms once made concrete and enforceable.
- The doctor mutation plan has the right safety instincts: lock-first mutation, byte-preserving edits, staged publication, custom/fork preservation, and failure-injection tests. It is blocked because those instincts are not yet specified as an enforceable transaction and recovery protocol.
- Reviewers consistently favored proof-driven migration: generated manifests, exact pinned public-pack compatibility checks, old-to-new witness maps, docs/schema goldens, and process/integration gates rather than prose-only equivalence claims.

## Critical Findings

### [Blocker] Implementation plan schema is invalid and the normative contract is fragmented

**Sources:** Camille Sato, Oleg Marchetti, Leah Okafor, Anand Krishnaswamy, Lena Hoffmann, Owen Gallagher, Hiroshi Tanabe, Yelena Markovic, Iris Kowalski, Claire Dubois

**Actionability:** document-fixable

**Issue:** The reviewed artifact does not conform to `gc.mayor.implementation-plan.v1`. It has valid front matter, but the required body is not present in order: `## Proposed Design` should be `## Proposed Implementation`, `## Data And State` is missing, `## Rollout` should be `## Rollout And Recovery`, and `## Summary` contains a long review-resolution ledger rather than a concise implementation summary. Because durable state, proof artifacts, rollback, and current contracts are split across attempt-history sections, implementers cannot tell which rules are authoritative.

**Required change:** Rewrite `plans/core-gastown-pack-migration/implementation-plan.md` into the exact schema: `Summary`, `Current System`, `Proposed Implementation`, `Data And State`, `Testing`, `Rollout And Recovery`, and `Open Questions`. Move attempt notes, review responses, and superseded ledgers into workflow artifacts or omit them from the normative plan. Put all durable state, generated artifacts, migration markers, cache/lock state, rollback rules, and proof gates in the required sections.

### [Blocker] Required Core loader and live-reload authority is not enforceable

**Sources:** Camille Sato, Hiroshi Tanabe, Iris Kowalski, Anand Krishnaswamy

**Actionability:** document-fixable

**Issue:** The plan says required Core must fail closed, but it does not define a single canonical production loader surface, no-refresh reload failure semantics, or the API/controller behavior after invalid Core is discovered. Direct `config.Load*` bypass containment also lacks an enforceable AST/import-aware scanner or runtime participation backstop for aliases, wrappers, method values, generated/API paths, and partial reads. The bootstrap migration can also create a collision gap if `core` and `registry` leave `bootstrapManagedImportNames` before replacement system-pack collision gates are live.

**Required change:** Choose and consistently name the canonical runtime loader APIs and call-site inventory. Specify live-reload behavior for missing, corrupt, stale, shadowed, or participation-missing Core, including last-known-good/degraded behavior, dispatch/API pause or refusal behavior, event/diagnostic output, and proving tests. Add a loader bypass scanner or runtime validated-participation stamp that covers CLI, controller, API, generated paths, partial reads, aliases, wrappers, and method/function values. Keep legacy bootstrap collision protection until the replacement system-pack gates are proven live.

### [Blocker] Behavior-evidence, public-pack, pin, and cache gates are not a machine-checkable acceptance surface

**Sources:** Oleg Marchetti, Lena Hoffmann, Iris Kowalski, Claire Dubois, Camille Sato

**Actionability:** mixed

**Issue:** The plan requires public Gastown behavior preservation, but reviewers found no single authoritative manifest contract with generator entrypoint, schema, row identity, digest rules, evidence classes, old/new witness mapping, source-deletion gate, and exact-public-pin packcompat proof. Public-pack pin and cache authority are also underspecified: legacy synthetic cache promotion could launder stale bytes, ordinary cache reads may trust key identity without digest verification, `RepoCacheKey` subpath identity is inconsistent, and compatibility versus activation pins are conflated in rollback.

**Required change:** Define one Behavior Evidence contract with generated manifest schema, generator command, output paths, row fields, owner, freshness tests, evidence-class enum, old-test migration map, source-deletion gate, and packcompat proof that reads the exact pinned public Gastown checkout or remote-cache path. Define `public-gastown-pins.yaml`, compatibility and activation pin semantics, digest verification for promotion and read hits, `RepoCacheKey` source+commit+subpath identity, and stale/mismatched cache failure behavior. If required public-pack commits or generated artifacts do not yet exist, mark them as external prerequisites rather than pretending the plan is decomposable.

### [Blocker] Doctor and runtime-state migration can leave cities half-migrated or silently drop legacy writes

**Sources:** Leah Okafor, Yelena Markovic, Iris Kowalski

**Actionability:** mixed

**Issue:** The mutation story still relies on per-file atomic renames for a multi-file city migration and does not define durable recovery before the first live publish. The current doctor runner may still allow direct `Check.Fix(ctx)` writes outside a coordinator. Runtime-state migration has an additional safety gap: old-binary writes after the migration marker can update pending archive push state, refs/remotes, escalation fields, or spawn-storm ledgers, while the new binary treats them as retained rollback state.

**Required change:** Replace "cross-file atomicity" with a transaction or convergence protocol that writes durable recovery state before mutation, defines publish order and commit point, and supports deterministic rerun or rollback after process death. Specify the `FixIntent`/mutation coordinator API, how legacy `Fix(ctx)` checks are adapted or refused, and the mutual-exclusion contract for doctor, controller, startup, pack install/update, and legacy fix paths. Put the runtime-state migration table in `Data And State`, including marker path/schema, shared lock, staged archive copy, archive digest semantics, old-binary post-marker write detection, push-cursor reconciliation, rollback, and re-upgrade behavior.

### [Blocker] Role neutrality and pack-boundary containment are not yet a single enforced contract

**Sources:** Anand Krishnaswamy, Owen Gallagher, Claire Dubois

**Actionability:** document-fixable

**Issue:** The plan still leaves role-bearing behavior and retired-source recognition spread across Go, Core assets, provider packs, public Gastown assets, API/dashboard generated artifacts, scripts, prompts, docs, and tests. `internal/packsource` is intended to be the single retired-source classifier, but the final body does not make it authoritative for all load, install, cache, lock, materialization, discovery, docs-lint, generated-reference-lint, and public-source normalization paths. Active Dog/Gastown role names, tmux role themes, sling formula-prefix heuristics, API crew classifications, scaffold/warmup fallbacks, and mail/nudge escalation recipients still need neutral replacements or narrow expiring compatibility rows.

**Required change:** Promote the role-neutrality and pack-boundary contract into the canonical plan body. Define the generated role-surface manifest, retired-source classifier API, scanner commands, row schemas, owner, freshness tests, allowlist expiry, and negative fixtures. Specify parser/resolver owners for `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER`. Replace active role-specific Core behavior with neutral names/mechanisms or prove retained names are non-behavioral compatibility aliases. Make `internal/packsource` the only authority for retired Maintenance/Gastown source classification across every active discovery and validation path.

### [Major] Bootstrap fixture isolation and skip semantics need stronger proof

**Sources:** Hiroshi Tanabe, Iris Kowalski, Camille Sato

**Actionability:** document-fixable

**Issue:** The desired empty-bootstrap end state is sound, but the plan does not fully inventory `GC_BOOTSTRAP` production and test harness dependencies, including doctor checks and command/testscript defaults. The fixture isolation story also needs guards against copied production Core content under allowed test paths, plus separate import and literal path guards for old bootstrap Core paths.

**Required change:** Add a repo-wide `GC_BOOTSTRAP` audit with dispositions for production code, doctor checks, command tests, testscript defaults, and helper paths. Define the empty `fs.FS` implementation and prove `Stat`, `WalkDir`, and `ReadFile` behavior. Add normal-command regression tests showing `GC_BOOTSTRAP=skip` cannot bypass Core materialization, strict validation, collision checks, typed participation, retired-source classification, provider materialization, or doctor cleanup. Split old bootstrap verification into Go import guards and string/path guards with exact fixture allowlists.

### [Major] Operator docs, schemas, and generated references are not release-gated with behavior changes

**Sources:** Claire Dubois, Iris Kowalski, Anand Krishnaswamy

**Actionability:** document-fixable

**Issue:** Operator-facing changes can ship before same-slice docs, schema, CLI/help, doctor-output, tutorial-transcript, OpenAPI/dashboard generated reference, and wording-matrix freshness gates are required. The current inventory command is too narrow and cannot distinguish retired Maintenance pack terminology from valid lowercase maintenance, Dolt/store maintenance keys, historical exceptions, or public-pack companion docs.

**Required change:** Define a generated wording/docs scanner contract covering Markdown/MDX, JSON, TXT, TS, OpenAPI, dashboard generated files, docs/schema outputs, public Gastown docs, generated help, CLI examples, scripts, prompts, tutorial transcripts, and doctor outputs. Add slice-by-slice release gates for public pins, doctor behavior, diagnostics, examples, Maintenance removal, and moved orders. Include `docs/tutorials/05-formulas.md` and `docs/tutorials/07-orders.md`, per-order dispositions for `mol-dog-*` and `prune-branches`, and strict allowlist metadata with owner, justification, expiry, and negative fixtures.

### [Major] Decomposition slices and external prerequisites are too broad or unresolved

**Sources:** Iris Kowalski, Oleg Marchetti, Owen Gallagher, Anand Krishnaswamy

**Actionability:** mixed

**Issue:** The first slice combines external `gascity-packs` ownership work, asset moves, behavior manifests, old-test mapping, compatibility pins, activation pins, and Gas City migration gates into one large critical-path phase. Cross-pack ownership remains unresolved for behavior-bearing surfaces such as `mol-review-quorum`, provider overlays, Dog prompt fragments, review checks, shutdown-dance examples, role-theme/tmux APIs, and moved orders. The plan also claims no open questions while these ownership choices affect proof artifacts, witnesses, aliases, and public commits.

**Required change:** Either decompose the public `gascity-packs` work into concrete smaller prerequisite tasks with paths, branches/PRs, artifacts, approval gates, and downstream dependencies, or mark it as an external prerequisite that blocks Gas City decomposition until the compatibility and activation artifacts exist. Resolve cross-pack ownership rows before source moves or public pin consumption. Add per-slice revert posture, one-way boundaries, and process/integration shard gates for high-risk concurrency, cache, doctor, runtime-state, and loader behavior.

### [Minor] Several narrow contracts need final cleanup before approval

**Sources:** Lena Hoffmann, Claire Dubois, Hiroshi Tanabe, Yelena Markovic

**Actionability:** document-fixable

**Issue:** The durable public ref is underdefined and must remain non-authoritative. `ensureBootstrapForDoctor` has no explicit adaptation/removal schedule. Order skip/tracking compatibility is still conditional rather than listing stable keys and aliases. Docs-lint allowlist discipline is weaker than the role scanner's.

**Required change:** State that durable public refs only keep SHAs fetchable and never replace immutable SHA plus digest validation. Schedule `ensureBootstrapForDoctor` cleanup. Name stable order keys and compatibility aliases, then prove existing skip/tracking beads still suppress the intended order. Apply owner, justification, expiry, and negative-fixture discipline to wording allowlists.

## Disagreements

- Several individual model lanes returned `approve-with-risks` while the persona synthesis blocked. My assessment agrees with the persona syntheses: in each case, the optional or more optimistic reviewer assumed contracts that are not yet present in the artifact, while another required reviewer identified schema, recovery, or proof gaps that affect decomposition safety.
- Reviewers disagreed on exact mechanisms in several areas: WAL versus convergence scan for doctor recovery, per-hit digest checks versus validation markers for cache reads, AST/token scanners versus equivalent fail-closed scanners, and public-cache promotion allowed with prior authenticated digest versus forbidden offline. These are implementation choices; the plan must state one enforceable mechanism and the tests that prove it.
- Some reviewers treated missing generated artifacts as document gaps, while others treated them as external prerequisites. My assessment is mixed: the plan can fix many blockers by naming contracts, gates, and prerequisite status, but actual public-pack commits, generated manifests, inventories, and proof transcripts may still be required before a later review can approve.
- Claude-oriented reviews often credited the design direction as strong, while Codex and DeepSeek/Gemini reviews tended to block on authority, freshness, and rollback enforcement. I give more weight to the blocking interpretation because this plan is intended to feed task decomposition, where ambiguous authority and missing artifact contracts turn into unsafe work items.

## Missing Evidence

- A schema-valid `implementation-plan.md` with one canonical implementation contract and no attempt-history strata in the normative body.
- Exact runtime loader API names, no-refresh/live-reload failure semantics, operator-visible diagnostics, and tests proving invalid Core cannot silently continue as current validated behavior.
- A loader bypass scanner or runtime participation backstop covering aliases, wrappers, method/function values, generated/API paths, and partial reads.
- Behavior manifest schema, generator command, paths, row identity, digest rules, evidence classes, approval authority, old/new witness map, source-deletion gate, and exact-public-pin packcompat proof.
- Public-pack pin ledger, compatibility/activation pin distinction, cache promotion validation authority, read-hit digest/provenance contract, subpath-aware `RepoCacheKey`, and stale/mismatched cache failure tests.
- Doctor mutation coordinator API, legacy `Fix(ctx)` adaptation policy, cross-process mutual exclusion, durable recovery state, failure-injection tests after each publish step, and byte-preserving TOML edit strategy.
- Runtime-state marker schema/path, archive staged-copy protocol, digest semantics, old-binary post-marker write detection, push-cursor reconciliation, rollback/re-upgrade behavior, and order skip/tracking aliases.
- `internal/packsource` classifier API, state enum, scanner/freshness tests, migration inventory for retired-pack literals, duplicate-active matrix, and stale-directory negative fixtures.
- Role-surface manifest and reconciliation rows for retained Dog/Mayor/Deacon/Polecat/Refinery/crew/Gastown strings across Go, assets, provider packs, public Gastown, API/OpenAPI/dashboard/generated TypeScript, docs, scripts, prompts, examples, and tests.
- Complete `GC_BOOTSTRAP` audit, empty bootstrap FS implementation, fixture allowlist/minimality tests, import and literal guards for old bootstrap Core paths, and normal-command `GC_BOOTSTRAP=skip` regression proof.
- Docs/schema scanner roots and generated outputs, slice release gates, tutorial and doctor-output goldens, operator recovery guide, and narrow expiring wording allowlists.
- External `gascity-packs` prerequisite plan or actual artifacts for public Gastown compatibility and activation commits, moved assets, manifests, pins, and ownership decisions.

## Convergence Assessment

- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: Another design-doc edit can directly resolve the universal schema blocker and several authority/contract blockers by producing a schema-shaped implementation plan with one canonical contract, explicit prerequisites, exact gates, and no stale attempt-history strata. However, the next iteration should not pretend prose alone proves external public-pack artifacts, generated inventories, or migration tests exist; it should either cite concrete artifacts or mark those items as external prerequisites.
- Next non-design work: public `gascity-packs` compatibility/activation artifacts, generated behavior and role-surface manifests, loader/retired-source inventories, packcompat transcripts, docs/schema goldens, and process/integration proof tests may be required before a later review can move from `block` to approval.

## Recommended Changes

1. Rewrite `implementation-plan.md` to the exact Mayor implementation-plan schema and remove attempt-history sections from the normative artifact.
2. Consolidate all durable state into `Data And State`: public pins, caches, manifests, inventories, lockfiles, migration markers, runtime archives, recovery records, rollback/downgrade state, wording matrices, and compatibility aliases.
3. Define the required Core loader contract end to end: canonical APIs, no-refresh/live-reload failure behavior, diagnostics, API/controller behavior, bypass scanner or participation stamp, and bootstrap collision-gate sequencing.
4. Replace scattered behavior-preservation prose with one Behavior Evidence contract and exact-public-pin packcompat gate, including witness classes and rules for behavior without old execution tests.
5. Specify public-pack pin and cache authority: compatibility versus activation pins, digest validation for promotion/read hits, subpath-aware cache keys, durable public ref limits, rollback matrix, and stale-cache failure behavior.
6. Specify doctor and runtime-state safety as a transaction/convergence protocol with a mutation coordinator API, durable recovery state, shared locks, staged archive copy, post-marker legacy-write detection, push-cursor reconciliation, and failure-injection tests.
7. Make role neutrality and pack-boundary containment enforceable through `internal/packsource`, generated role-surface and retired-source inventories, negative fixtures, compatibility alias rules, and concrete replacements for tmux themes, sling branch heuristics, crew classifications, prompt/scaffold defaults, and escalation recipients.
8. Resolve `GC_BOOTSTRAP` and bootstrap fixture cleanup with a complete dependency audit, exact fixture allowlists, empty FS tests, import/path guards, and a normal-command skip regression test.
9. Add operator-facing release gates for docs, schemas, OpenAPI/dashboard generated references, CLI/help, doctor output, tutorial transcripts, wording matrices, and recovery guidance in every slice that changes user-visible behavior.
10. Decide whether public `gascity-packs` work is part of this decomposition or an external prerequisite; either way, provide concrete artifact paths, ownership rows, proof commands, and downstream dependencies before claiming `Open Questions: None`.
