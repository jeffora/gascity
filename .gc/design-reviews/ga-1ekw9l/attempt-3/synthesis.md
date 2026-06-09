# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins yields `block`: eight of the ten available persona syntheses block, and the remaining two approve only with material risks. The implementation plan is now structurally schema-conforming, but it still lacks enforceable contracts for evidence, pin/cache authority, role neutrality, loader/bootstrap safety, runtime-state migration, and prerequisite-aware rollout. This attempt also has a workflow artifact defect: the current attempt's `persona-syntheses/` directory is empty while the persona synthesis beads stamped fresh output paths under `attempt-1`.

## Consensus Strengths

- Multiple reviewers agreed the Mayor implementation-plan schema is now structurally satisfied: front matter is present, the required top-level sections are in order, and prior attempt-history prose has been removed from the normative body.
- Reviewers consistently praised the direction of a single `internal/systempacks` loader boundary with strict required-pack materialization, typed participation validation, and fail-closed no-refresh behavior.
- The plan now distinguishes compatibility pins from activation pins and blocks destructive Gas City source deletion until public Gastown evidence exists.
- The proposed `FixIntent` and doctor mutation coordinator direction is sound: stage first, lock before mutation, validate provenance repeatedly, compare before rename, and record recovery state before publication.
- `internal/packsource`, generated behavior evidence, role-surface inventories, wording matrices, and exact public-pack packcompat checks are the right kinds of mechanisms if their contracts become concrete.
- The rollout plan is better sliced than earlier attempts and explicitly recognizes one-way boundaries around activation and committed mutation/runtime-state markers.

## Critical Findings

### [Blocker] Attempt 3 review artifacts are inconsistent

**Sources:** Camille Sato, Oleg Marchetti, Leah Okafor, workflow artifact inspection
**Actionability:** workflow-defect
**Issue:** The claimed global synthesis step is for attempt 3, and `.gc/design-reviews/ga-1ekw9l/attempt-3/persona-syntheses/` exists but contains no files. The ten persona synthesis beads closed with `gc.attempt=3`, but their `design_review.output_path` metadata points at `.gc/design-reviews/ga-1ekw9l/attempt-1/persona-syntheses/*.md`; those files were freshly overwritten during this attempt. Several raw review files are also missing from the attempt-3 `reviews/` directory, while the corresponding persona syntheses cite sources from the earlier output path. This makes the review chain malformed even though the available persona syntheses are readable.
**Required change:** Fix the design-review workflow's attempt metadata and output-path resolution so dynamic persona synthesis beads write to the live attempt directory, source only the live attempt's raw reviews, and fail or explicitly soft-fail missing required Claude/Codex inputs. Rerun the affected review iteration after the artifact routing bug is fixed.

### [Blocker] Behavior evidence, public-pack proof, and cache authority are not an executable acceptance surface

**Sources:** Oleg Marchetti, Lena Hoffmann, Owen Gallagher, Iris Kowalski, Camille Sato
**Actionability:** external-prerequisite
**Issue:** The plan still depends on artifacts and checks that are not yet concrete enough to prove behavior preservation or pin identity: behavior manifest schema/generator paths, source-kind coverage, witness-type matrix, Git historical baseline, exact public Gastown compatibility and activation commits, public pin ledger, packcompat transcripts, pack digests, cache provenance, and offline cache-hit/cache-miss tests. Reviewers also found escape paths through bundled synthetic aliases, stale cache entries, source+commit cache keys without subpath identity, source assertions used as behavior proof, and current-workspace-only deletion scans.
**Required change:** Produce or explicitly block on first-class evidence artifacts: `public-gastown-pins.yaml` schema and freshness check, behavior-preservation manifest schema/generator and sample rows, witness-kind matrix, Git delta/baseline deletion guard, exact-public-pin packcompat transcript, digest/provenance proof for promotion and read hits, subpath-aware `RepoCacheKey`, and offline cache-hit/miss tests that fail closed without synthetic or embedded fallback.

### [Blocker] Role neutrality and pack-boundary containment are still not enforceable

**Sources:** Anand Krishnaswamy, Owen Gallagher, Claire Dubois
**Actionability:** document-fixable
**Issue:** Active role-specific behavior remains in compiled or required-pack paths. Reviewers called out Go formula-name heuristics for `mol-polecat-*` and `mol-refinery-patrol`, under-specified maintenance-worker bindings, hardcoded `mayor/` and `deacon/` recipient targets in required provider-pack behavior, and no complete active-vs-historical role-surface manifest. `internal/packsource` is the right authority, but the plan does not yet make it the sole classifier for every behavior enumerator, cache/lock/materialization path, docs/generated-reference lint, stale-directory handling, and rollback state. Duplicate-active behavior is also not proven across compatibility pins, activation pins, stale generated directories, synthetic cache, ordinary remote cache, and old/new binary skew.
**Required change:** Amend the plan with a binding and role-surface contract: parse/resolve owners and precedence for `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and script environment injection; no Go fallback to concrete role names; recipient bindings for required provider-pack escalation targets; formula-declared metadata replacing formula-name heuristics; generated role-surface and retired-source manifests with owners, expiry, freshness tests, negative fixtures, and coverage across Go, TOML, shell, docs, generated assets, provider packs, and public Gastown files. Make `internal/packsource` mandatory for all active discovery and add zero-duplicate-active/zero-merge gates.

### [Blocker] Loader and bootstrap safety still lack decomposition-grade contracts

**Sources:** Camille Sato, Hiroshi Tanabe, Iris Kowalski
**Actionability:** document-fixable
**Issue:** The required loader design names good boundaries but not the typed runtime result that prevents stale last-known-good config from reaching behavior-changing callers. Provider-conditioned required-pack selection, the direct `config.Load*` bypass inventory, and API/controller guard points remain under-specified. Bootstrap removal has a sharper blocker: `internal/bootstrap/bootstrap.go` still embeds `packs/**`, and the plan does not name the concrete non-nil empty production `fs.FS` that safely replaces embedded Core after `internal/bootstrap/packs/core` is removed. `GC_BOOTSTRAP=skip` containment also needs a production-like regression, not only prose.
**Required change:** Define the `internal/systempacks` runtime-load result or equivalent guard with freshness, diagnostics, participation records, and allowed-use mode; enumerate behavior-changing caller gates; keep provider selection inside the systempacks boundary; seed the direct-loader allowlist with owner/reason/proof rows; specify the empty production bootstrap `fs.FS`, embed directive changes, `Stat`/`WalkDir`/`ReadFile` tests, import/path scanners, fixture-copy guards, and a normal-command `GC_BOOTSTRAP=skip` failure-closed regression.

### [Blocker] Runtime-state migration trigger, lock domain, and state table are contradictory or incomplete

**Sources:** Yelena Markovic, Leah Okafor, Iris Kowalski
**Actionability:** document-fixable
**Issue:** Runtime-state migration currently implies both controller-startup ownership and doctor-only mutation safety, while the mutation coordinator refuses when a controller for the same city is running. The plan also lacks a concrete migration table for JSONL archive state, `.git` refs/remotes, push cursors, escalation fields, spawn-storm ledgers, order skip/tracking aliases, explicit formula env state, and retained legacy paths. Post-marker old-binary writes, duplicate offsite pushes, half-copied archives, split-brain spawn-storm throttling, and divergent re-upgrade behavior are named but not specified.
**Required change:** Choose the migration owner and lock model. If migration runs on controller startup, define the startup-only coordinator exception and shared advisory lock; if doctor-only, require the controller to stop and make new controllers refuse un-migrated legacy state. Add an artifact-by-artifact migration table with old path, new Core path, legacy fallback, owner, read precedence, copy/merge/conflict rules, downgrade behavior, marker schema, digest semantics, staged git-aware archive copy, push reconciliation, spawn-storm read-union, order alias mapping, and failure-injection fixtures.

### [Major] Doctor mutation safety is directionally approved but not implementation-ready

**Sources:** Leah Okafor, Claire Dubois
**Actionability:** document-fixable
**Issue:** Reviewers approve the mutation-safety shape with risks, but still need concrete package/API ownership, legacy `Check.Fix(ctx)` migration policy, a scanner or guard proving writes cannot bypass the coordinator, recovery phases and commit point semantics, byte-preserving TOML strategy, generated-vs-fork content-digest provenance, and shared lock participation by controller, startup, pack install/update, doctor, scripts, and old binaries.
**Required change:** Add the coordinator package and exported `FixIntent`/runner API, list all callers and legacy fix dispositions, define recovery phase transitions and stale-record cleanup, state the TOML edit mechanism or refuse-only behavior, require content-digest provenance fixtures, and name lock ownership checks before each publish rename.

### [Major] Rollout and decomposition remain prerequisite-opaque

**Sources:** Iris Kowalski, Oleg Marchetti, Lena Hoffmann, Yelena Markovic
**Actionability:** document-fixable
**Issue:** `Open Questions: None` and `status: draft` conflict with unresolved prerequisite work: parent requirements are still questioned, public `gascity-packs` branch/PR and immutable commits are not named, AC6/AC7 artifacts are absent, and several rollout slices still bundle multiple failure domains. Slice 4 combines loader API, typed participation, scanner/allowlist, read-only diagnostics, mutation coordinator, and doctor paths; Slice 5 bundles activation, public Gastown consumption, no-Maintenance proof, duplicate behavior checks, and source deletion readiness.
**Required change:** Make prerequisite status explicit in the plan and later `tasks.md`: either create in-scope beads for public Gastown compatibility/activation artifacts, generated manifests, pin ledger, ownership rows, and proof transcripts, or mark them as external blockers with owners and downstream dependencies. Add a prerequisite/evidence slice, split broad rollout slices into smaller vertical boundaries, and attach exact commands, artifact paths, old/new binary expectations, cache/offline expectations, and rollback or one-way classifications to each slice.

### [Major] Operator docs and generated wording gates need executable matrices

**Sources:** Claire Dubois, Lena Hoffmann, Anand Krishnaswamy
**Actionability:** document-fixable
**Issue:** The docs/schema lane no longer blocks on artifact shape, but the wording scanner is not executable enough to distinguish retired standalone Maintenance-pack references from valid lowercase maintenance, Dolt/store-maintenance terminology, Core worker bindings, public Gastown wording, stale generated paths, and historical examples. Generated OpenAPI/dashboard/docs/schema/CLI/help/tutorial/doctor outputs need clear regeneration ownership and CI order before wording lint.
**Required change:** Add a terminology matrix with token classes, allowed/denied contexts, owners, examples, false-positive handling, and negative fixtures. Name the regeneration and freshness commands for OpenAPI, dashboard types, docs/schema outputs, CLI reference/help, tutorial transcripts, doctor output, and public Gastown companion docs. Add one operator recovery guide from stale/retired paths or missing pins through doctor diagnostics, public pin/cache verification, repair refusal/mutation, and manual recovery.

### [Minor] Several compatibility edges need final precision

**Sources:** Lena Hoffmann, Claire Dubois, Hiroshi Tanabe, Yelena Markovic
**Actionability:** document-fixable
**Issue:** Durable public refs must remain non-authoritative, old binary plus activation pin needs a hard downgrade boundary, `ensureBootstrapForDoctor` and implicit-import cache checks need explicit disposition, fixture-copy guards should derive from shipped Core content, and tmux integration cleanup must target isolated sockets only.
**Required change:** Add those narrow rules to the relevant sections and tests: immutable SHA plus digest always wins over durable refs; activation-pinned cities cannot safely downgrade to the old binary without lock rollback/manual recovery; bootstrap doctor compatibility paths are moved or removed; fixture prohibited lists are freshness-checked; tmux cleanup never uses a bare server kill.

## Disagreements

- Several raw model reviewers moved from `block` to `approve-with-risks` on structural schema and high-level design direction, while persona syntheses still blocked after weighing missing proof artifacts, stale workflow paths, and decomposition readiness. My assessment: the schema block is resolved, but the remaining blockers are substantive.
- Claude-oriented reviews often accepted missing generated artifacts as future prerequisite work; Codex and DeepSeek-style reviews treated missing artifacts as blockers for decomposition. My assessment: future work is acceptable only if it is explicit in the plan as a blocking prerequisite with owner, path, command, and dependency edge.
- Reviewers proposed different valid mechanisms in several places: provider selection algorithm versus validating all built-in providers; per-read full digest verification versus tamper-evident validation markers; controller-startup runtime migration versus doctor-only migration; source scanners versus runtime participation backstops. Any of these can pass if the plan chooses one and attaches tests.
- Leah Okafor and Claire Dubois returned `approve-with-risks`, while most other personas blocked. Their approvals should not lower the global verdict because their own required changes still feed the blocking runtime-state, rollout, and operator-safety findings.
- The workflow artifact defect is separate from the design verdict. It does not make the design content safe or unsafe by itself, but it does make this attempt's review chain unreliable until the attempt path and required-source behavior are fixed.

## Missing Evidence

- Attempt-3 persona synthesis files in `.gc/design-reviews/ga-1ekw9l/attempt-3/persona-syntheses/` with correct `design_review.output_path` metadata.
- Complete attempt-3 raw review source set or explicit soft-failure records for absent optional reviewers; required Claude/Codex absence must fail the persona synthesis.
- Behavior manifest schema, generator command, sample rows, source-kind extractor matrix, witness-kind matrix, owner, freshness tests, and Git historical baseline/deletion guard.
- Exact public Gastown compatibility commit, activation commit, pin ledger, packcompat transcript, behavior manifest digest, pack digest, and ordinary remote/cache proof.
- Offline public-pack cache-hit/cache-miss tests, digest mismatch diagnostics, synthetic alias rejection tests, subpath-aware `RepoCacheKey`, and atomic cache promotion protocol.
- Role-surface and retired-source manifests with active-vs-historical classification, allowlist expiry, negative fixtures, and coverage across Go, assets, required provider packs, generated artifacts, docs, scripts, prompts, examples, and tests.
- Runtime loader typed result or guard API, direct-loader bypass inventory, provider-pack selection matrix, API/controller behavior-readiness tests, and invalid Core diagnostic events.
- Empty production bootstrap `fs.FS`, embed directive update, deleted bootstrap Core import/path scanner, fixture-copy guard, and `GC_BOOTSTRAP=skip` production-like regression.
- Runtime-state migration trigger policy, shared lock domain, exact artifact/field/path table, marker schema, archive copy protocol, old-binary post-marker policy, push reconciliation, spawn-storm read-union, and order alias tests.
- Doctor mutation coordinator package/API, legacy fix migration policy, recovery state machine, TOML preservation/refusal strategy, content-digest fork provenance, and bypass guard tests.
- AC6 asset migration ledger, AC7 behavior-preservation harness, public-pack ownership plan, branch/PR/commit references, proof transcript paths, and per-slice gates.
- Terminology matrix, generated-artifact freshness commands, doctor output goldens, tutorial transcript goldens, and one operator recovery guide.

## Convergence Assessment

- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: Another design-doc edit can directly resolve several blocker and major findings: binding semantics, role-surface and packsource contracts, loader result shape, bootstrap empty-FS details, runtime-state trigger/lock policy, rollout prerequisite honesty, and docs wording matrices. However, the next iteration should not pretend external artifacts exist; it must either cite concrete public-pack/generated evidence or mark those items as external prerequisites. The review workflow artifact bug also needs non-design repair before the next review can be trusted.
- Next non-design work: fix dynamic attempt metadata/output path handling in the design-review workflow; generate or create placeholder-proof artifacts for public Gastown pins, behavior evidence, role/source inventories, packcompat transcripts, cache/offline tests, runtime-state fixtures, and docs/schema goldens before another approval-seeking review.

## Recommended Changes

1. Fix the review workflow artifact routing so attempt-3 persona syntheses are written under attempt-3 and source only attempt-3 raw reviews; rerun the malformed parts of this review iteration.
2. Update `implementation-plan.md` to mark external evidence prerequisites honestly and name owners, paths, commands, and downstream dependencies for public Gastown commits, behavior manifests, pin ledgers, and packcompat transcripts.
3. Define the Behavior Evidence contract in executable terms: source-kind extractors, Git baseline/deletion guard, witness-kind matrix, row schema, sample rows, semantic delta/removal authority, and exact-public-pin verification.
4. Specify public-pack pin/cache/offline authority: `public-gastown-pins.yaml`, compatibility versus activation pins, digest scope, synthetic alias rejection, subpath-aware `RepoCacheKey`, atomic promotion, and offline hit/miss diagnostics.
5. Make role neutrality enforceable with binding precedence, resolver ownership, no concrete-role Go fallbacks, provider-pack recipient bindings, formula-declared metadata, role-surface scanner, and SDK self-sufficiency tests.
6. Make `internal/packsource` the single retired-source authority across every active discovery and validation path, with stale-directory, zero-duplicate-active, zero-merge, old/new binary, cache, and rollback tests.
7. Define the required-systempacks loader result/guard, invalid Core behavior, provider-pack selection, direct-loader allowlist inventory, API/controller guard tests, and no-refresh read-only status behavior.
8. Specify bootstrap removal: non-nil empty production `fs.FS`, embed changes, import/path scanners, minimal fixture helpers, production-safe `GC_BOOTSTRAP=skip` behavior, and collision-gate sequencing.
9. Choose runtime-state migration ownership and lock semantics, then add the exact state table, marker schema, staged git-aware archive copy, post-marker old-write detection, push reconciliation, spawn-storm continuity, and order alias compatibility.
10. Add doctor coordinator implementation anchors, recovery phases, TOML preservation/refusal rules, content-digest fork classification, lock participation, and bypass guard tests.
11. Split rollout slices into smaller vertical tasks with exact proof commands, artifact paths, old/new binary expectations, cache/offline expectations, and rollback or one-way classifications.
12. Add executable operator docs gates: terminology matrix, generated artifact freshness commands, doctor output goldens, tutorial transcripts, upgrade recovery guide, and final vocabulary examples.
