# Design Review Synthesis

## Overall Verdict: block

Nine of ten persona syntheses return `block`, so the global verdict is `block`
by worst-verdict-wins. The design has improved materially around behavior
evidence, typed Core participation, retired-source classification, doctor
coordination, and public-pack compatibility gates, but the implementation plan
still contains contradictions and unscheduled transition boundaries that can
load the wrong pack content, lose Gastown behavior, mutate operator data
outside the promised coordinator, or ship misleading operator guidance.

## Consensus Strengths

- Reviewers consistently praised the move from path/count checks toward typed
  participation records, behavior manifests, packcompat witnesses, and
  duplicate-definition gates.
- The two-pin public Gastown idea is seen as a plausible way to avoid
  bundled/public overlap, provided the compatibility and activation mechanics
  become explicit and executable.
- The retired-source classifier direction is broadly accepted as the right
  central control point for old Maintenance, Gastown, synthetic-cache, import,
  lock, and stale-directory handling.
- The doctor mutation coordinator concept is the right preservation model for
  comment-stable TOML edits, lock/install changes, Core repair, runtime-state
  migration, and failure recovery.
- Reviewers agree that behavior preservation has been framed correctly: old
  trigger conditions, recipients, channels, requester/detector semantics,
  authorship identity, runtime-state effects, and public replacement commits
  must have observable witnesses.
- The design recognizes that docs, wording matrices, examples, and generated
  reference/schema artifacts are release gates, not after-the-fact cleanup.

## Critical Findings

### [Blocker] Required Core Loading Still Has Contradictory Proof Semantics
**Sources:** Elias Sato; Marcus Driscoll; Tomas Park; Sofia Khoury
**Issue:** The newest contract requires pre-resolution Core integrity and
post-resolution typed `RequiredSystemPackParticipation`, but operative design
sections still permit path/provenance-only checks such as
`assertRequiredSystemPackProvenance` or expected-file comparisons. The loader
bypass inventory is also incomplete: concrete production surfaces in
`internal/dispatch`, `internal/doctor`, `internal/configedit`, controller
reload/no-refresh paths, import/cache/lock validation, and command helpers are
not all classified.
**Required change:** Rewrite the constructive loading sections around two fatal
gates: strict required-pack file-set integrity before resolution and typed
participation after resolution. Name the importable `internal/` loading
boundary or config-injection model, regenerate the production loader inventory,
classify each partial-read exception with tests, and add scanner tests that
cover non-test `cmd/gc` plus production `internal/` config-loading surfaces.

### [Blocker] Public Gastown Activation Is Not Scheduled Safely
**Sources:** Yuki Hayashi; Tomas Park; Nadia Volkov; Marcus Driscoll; Avery McAllister
**Issue:** The design describes compatibility and activation pins, but the
rollout still only clearly updates `PublicGastownPackVersion` for the
compatibility pin. That leaves a possible release state where bundled
Maintenance behavior is removed while the public replacement behavior is still
inactive or unadopted. Cache identity, synthetic alias retirement, durable
public refs, offline behavior, and old/new binary lock states are also not
specified enough to make the pin transition executable.
**Required change:** Add an explicit activation-pin adoption step or declare a
paired cross-repo activation/removal boundary. For both pins, name the commit,
durable public ref, registry/direct identity checks, remote-cache and
behavior-manifest digests, duplicate-definition matrix, old/new binary matrix,
offline/cache behavior, no-Maintenance packcompat mode, and rollback or
one-way-upgrade procedure.

### [Blocker] Public Gastown, Host Core, And Dog/Worker Routing Remain Unsettled
**Sources:** Avery McAllister; Ingrid Kovac; Felix Moreau; Marcus Driscoll; Nadia Volkov
**Issue:** The design does not choose a concrete public Gastown dependency
model: patch auto-included Core, explicitly import Core, or own Dog-related
assets. As a result, formula pools, `gc.routed_to`, mail/nudge targets,
warrant metadata, prompt examples, `[[patches.agent]] name = "dog"`, renamed
maintenance-worker behavior, and missing-host-Core diagnostics cannot be
implemented or tested safely. The role-neutrality lane also finds unresolved
Go and asset role surfaces such as tmux role maps, `DogTheme`, default mayor
scaffolding, prompt fallbacks, warmup defaults, sling formula-name heuristics,
API crew classification, provider-pack routes, and public-pack TOML defaults.
**Required change:** Choose the public Gastown to host Core model and define a
host-Core worker reference contract for route, pool, patch, prompt, metadata,
and mail/nudge surfaces under public-pack bindings. Add an asset-by-asset
role-surface table with owner, replacement mechanism, slice, and proof test;
resolve Dog/session hooks/theme/keybinding/agent-field ownership; and expand
role-token scanning to Go, TOML defaults, formula text, prompts, scripts,
overlays, metadata, comments, and generated docs with field-scoped allowlists.

### [Blocker] Retired Maintenance And Stale Gastown Sources Can Still Participate
**Sources:** Marcus Driscoll; Avery McAllister; Ritu Raman; Felix Moreau; Tomas Park
**Issue:** Removing Maintenance from required/bundled includes does not prove
that stale `.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, old
synthetic caches, orphaned locks, transitive `[imports.maintenance]`, local
`examples/gastown/packs/*` imports, public Maintenance aliases, or stale
prompt/template directories cannot still influence config loading. Some
registry/cache/helper dispositions remain implicit, including
`publicSubpathForPack`, `PublicRepository`, `normalizeRepository`,
`requiredBuiltinPackNames`, `RetiredBootstrapPacks`, and materializer pruning.
**Required change:** Centralize retired-source classification and require
call-site-specific behavior for config load, install/check, cached imports,
lock validation, doctor, packman, prompt/template discovery, and materializer
repair. Preserve operator-edited retired directories, but exclude them from
active content discovery; diagnose or migrate imports/locks before ordinary
remote fallback; and add stale-directory, public-alias, transitive-import, and
offline old-cache fixtures.

### [Blocker] `gc doctor --fix` Is Not Yet Coordinator-Owned Or Failure-Atomic
**Sources:** Sofia Khoury; Elias Sato; Marcus Driscoll; Yuki Hayashi; Felix Moreau
**Issue:** The design promises one staged mutation coordinator, but later text
still tells the Core Presence Doctor `Fix` to call
`MaterializeBuiltinPacks(cityPath)` directly, and the current doctor runner
allows each `CanFix()` check to mutate independently. Runtime-state migration,
generated/custom provenance, controller-active refusal, doctor-vs-doctor
exclusion, partial-publish recovery, air-gapped public-pin handling, lock-entry
disposition, and downgrade behavior are not yet executable contracts.
**Required change:** Replace direct per-check migration writes with fix intents
or one coordinator-owned writer. Define the coordinator API, staged layout,
preflight, validation hooks, publish records, partial-state convergence, active
controller refusal, doctor concurrency guard, generated-source provenance,
runtime-state copy/move/fallback policy, offline guidance, and failure
injection fixtures before any doctor implementation bead can proceed.

### [Blocker] Behavior Preservation Evidence Is Still Not A Blocking First Slice
**Sources:** Nadia Volkov; Avery McAllister; Tomas Park; Ingrid Kovac; Marcus Driscoll
**Issue:** Reviewers accept the behavior-manifest direction, but the manifest,
public Gastown replacement commit, exact consuming pin, old-to-final witness
mapping, and packcompat execution evidence are not yet concrete blocking
deliverables. Several observable behaviors remain under-specified: requesters,
detectors, notifications, mail/nudge channel and recipient semantics, trigger
conditions, authorship identity, route metadata, runtime-state mutation,
script branches, prompt fragments, Dog flows, Polecat and branch pruning,
commands, doctor checks, and provider/example behavior.
**Required change:** Make the behavior manifest and public replacement evidence
the first implementation gate. Include fields for old path, new path, public
landing commit, consuming pin, trigger, requester/detector semantics, channel,
recipient, author identity, old witness, final witness, semantic delta, and
removal approval. Require execution-level witnesses whenever the old behavior
had execution-level evidence, including packcompat behavioral-trigger fixtures.

### [Blocker] Bootstrap And Implicit-Import Semantics Are Ambiguous
**Sources:** Ritu Raman; Elias Sato; Marcus Driscoll; Tomas Park
**Issue:** `GC_BOOTSTRAP=skip` is simultaneously described as retired
production behavior and as a retained legacy materialization skip. Current code
can return before retired implicit-import cleanup and collision checks. The
transition for `bootstrapManagedImportNames`, legacy implicit imports named
`core` or `registry`, `RetiredBootstrapPacks`, the implicit-import-cache doctor,
post-embed `bootstrapAssets`, and fixture isolation is not sequenced precisely.
**Required change:** Choose the final `GC_BOOTSTRAP=skip` rule and make every
section and test match it. Define the implicit-import lifecycle, classify
legacy `core`/`registry` entries, keep production `bootstrapAssets` as a
private empty/erroring `fs.FS`, restrict fixtures to `_test.go`, inventory all
old `packs/core` and `BootstrapPacks` references, and state where real Core
content fidelity is tested after bootstrap fixture isolation.

### [Blocker] Docs, Terminology, Runtime-State Guidance, And Tutorials Are Not Slice-Aligned
**Sources:** Felix Moreau; Sofia Khoury; Avery McAllister; Yuki Hayashi
**Issue:** The wording matrix provenance, public Gastown companion artifact,
canonical system-pack reference page, docs navigation, runtime-state migration
table, moved-order naming policy, `[[patches.agent]]` semantics, and tutorial
proofs are still unresolved or late in the rollout. Operator-facing behavior
changes in doctor/import-state/Maintenance-retirement slices can ship before
the docs, generated references, CLI help, doctor output, examples, and
troubleshooting guidance become truthful.
**Required change:** Decide whether the wording matrix is hand-owned or
generated, name its schema and consumers in both repos, and re-ground
`docs/reference/system-packs.md` plus navigation against the actual baseline.
Move docs/golden updates into the first behavior-changing slice they describe
or enforce a real non-release gate. Add runtime-state migration tables,
whole-flow tutorial goldens, docs-site integrity tests, public-pack wording
digest checks, and case-aware manifest-derived docs lint.

### [Blocker] Examples And Slice Gates Still Permit Red Intermediate States
**Sources:** Tomas Park; Yuki Hayashi; Avery McAllister; Felix Moreau; Nadia Volkov
**Issue:** Slices 5-6 can leave `examples/gastown` importing local
`../maintenance` while Maintenance behavior is folded or removed, producing
duplicate definitions or unresolved imports. The large existing Gastown example
tests are not mapped to replacement packcompat or moved tests. `examples/dolt`
and provider-pack routes still carry Dog/dolt target assumptions. Broad suite
names and prose are not enough for loader, doctor, registry, cache,
runtime-state, docs, old/new binary, offline, and stale-path transitions.
**Required change:** Assign examples/gastown import rewiring and physical test
splitting to a named slice at or before Maintenance folding. Add a
command-level gate appendix with focused package targets, documented sharded
targets, old/new binary fixtures, offline/cache fixtures, stale-path scans,
docs lint, generator freshness checks, `make test-fast-parallel`, and
`go vet ./...`. Map every removed or rewritten example test to a replacement
behavior witness.

### [Major] Required Pack Integrity And Materialization Need Exact Active-Pack Semantics
**Sources:** Marcus Driscoll; Elias Sato; Ritu Raman
**Issue:** Expected-file-only or manifest-shape checks cannot prove required
Core is not carrying injected behavior. At the same time, retired directories
must be preserved without being refreshed, pruned, or loaded. Provider `bd` and
`dolt` continuity depends on old synthetic caches self-healing while retaining
byte/provenance behavior.
**Required change:** Define strict full-file-set validation for required packs
after repair and before behavior discovery, or an equally mechanical proof that
unexpected files cannot influence resolution. Define the active-pack filter for
materialization and add offline old-cache-to-new-binary tests for Core, `bd`,
and `dolt`.

### [Major] Generated Artifact And Freshness Contracts Are Too Vague
**Sources:** Tomas Park; Felix Moreau; Nadia Volkov
**Issue:** The design names generated behavior, role-surface, wording, and
old-test-to-new-test artifacts, but not all generator commands, schemas,
owners, provenance, digest/freshness tests, or conflict rules are defined.
**Required change:** Name generator entrypoints, input sources, output schema
paths, owners, digest checks, and CI freshness tests for each generated
artifact. If a file is hand-owned, remove the `generated` claim and test its
schema directly.

### [Minor] Current Persona-Synthesis Artifacts Still Land In The Wrong Attempt Directory
**Sources:** Global synthesizer assessment; Tomas Park; prior synthesis notes
**Issue:** The ten closed attempt-10 persona synthesis beads stamped
`gc.attempt=10` and were updated during this run, but their
`design_review.output_path` values point under
`.gc/design-reviews/ga-2404qu/attempt-1/persona-syntheses/`. The current
attempt directory contains `output.json`, metadata, `design-before.md`, and one
raw review, not the expected ten attempt-local persona synthesis files. This
global synthesis used the closed bead-declared outputs because they are
complete and metadata-linked to attempt 10, but the artifact identity is still
fragile.
**Required change:** Fix persona-synthesis output path calculation so attempt
N writes under `attempt-N/persona-syntheses`, or make global synthesis copy or
manifest the exact source files it consumed. Add a pre-global-synthesis guard
that validates every source path against the current attempt metadata before
the review can be used as approval evidence.

### [Minor] Stale Historical And Terminology Text Can Mislead Implementers
**Sources:** Nadia Volkov; Ritu Raman; Felix Moreau; Marcus Driscoll
**Issue:** Several lower design sections and docs comments still carry
superseded instructions, historical attempt/verdict wording, stale "Core and
maintenance" comments, and imprecise use of "Maintenance" versus store/Dolt
maintenance.
**Required change:** Remove or clearly supersede stale operational text. Use a
single terminology matrix to distinguish the retired Maintenance pack from
Core utility work and store/Dolt maintenance, and make stale comments/docs part
of the release gate.

## Disagreements

- Overall verdicts differ by lane: Nadia returns `approve-with-risks`, while
  the other nine persona syntheses return `block`. Worst-verdict-wins makes
  the global verdict `block`.
- Some raw reviewers approve with risks in individual lanes, especially where
  they read the newest contract text as authoritative. Blocking reviewers point
  to older operative sections, current code mechanics, missing slices, and
  unclassified call sites. My assessment: implementers will follow the whole
  plan, so contradictions in actionable sections remain blockers.
- Reviewers differ on whether the two-pin public Gastown model should survive
  or be replaced by a single atomic activation/removal boundary. My assessment:
  either can work, but the current design must choose a mechanism, name the
  activation step, and prove old/new binary plus duplicate-definition behavior.
- Reviewers differ on the exact mechanism for stale directories and runtime
  state: preserve-in-place with active-path exclusion, move/rename, copy with
  fallback, or rollback helper. My assessment: preservation is mandatory, but
  the mechanism can vary if it prevents active loading, supports rerun
  convergence, and has downgrade guidance.
- Source labels are inconsistent in some persona lanes, with Gemini-named
  artifacts self-identifying as DeepSeek. This is not a synthesis failure
  because the persona syntheses record the discrepancy and the substantive
  findings converge.

## Missing Evidence

- Final `RequiredSystemPackParticipation` type, producer, resolver integration,
  failure matrix, and whole-production config-load inventory.
- Exact compatibility and activation pins, durable public refs, digest checks,
  registry/direct equality proof, cache-cold fetch proof, duplicate-definition
  matrix, and no-Maintenance packcompat transcript.
- Public Gastown `pack.toml` dependency model and host-Core worker reference
  contract under normal public-pack binding, renamed-worker, and missing-Core
  cases.
- Complete behavior manifest with execution witnesses for requesters,
  detectors, notifications, channels, recipients, authorship identity,
  runtime-state mutation, script branches, prompt fragments, commands, doctor
  checks, and provider/example behavior.
- Retired-source classifier call-site table covering config load, install,
  check, cache, lock, doctor, packman, materializer, and prompt/template
  discovery.
- Doctor mutation coordinator API, runner integration, staged publish records,
  generated-source provenance, controller-active refusal, doctor concurrency
  fixture, runtime-state policy, partial-publish convergence tests, and offline
  guidance.
- Final `GC_BOOTSTRAP=skip` rule, implicit-import lifecycle, production
  `bootstrapAssets` contract, test-only fixture allowlist, and hidden dependency
  inventory.
- Runtime-state migration table for JSONL archives/export state, push-failure
  counters, spawn-storm state, order-tracking state, rollback, doctor behavior,
  and docs sequencing.
- Docs wording artifact provenance, public Gastown companion artifact, canonical
  system-pack reference baseline, docs navigation/index tests, generated
  reference/schema lint, CLI/help/doctor-output goldens, and tutorial
  transcripts.
- Machine-checkable old-test-to-new-test mapping for `examples/gastown` and
  `examples/dolt`, with command-level slice gates and generator freshness
  tests.
- Current-attempt artifact guard or explicit source manifest for persona
  syntheses.

## Recommended Changes

1. Reconcile the design into one actionable implementation contract: remove or
   rewrite stale lower sections on Core loading, doctor fixes, bootstrap,
   registry/cache cleanup, pin adoption, and docs sequencing.
2. Add the activation-pin rollout step or replace the two-pin model with an
   atomic activation/removal boundary, then attach duplicate-definition,
   old/new binary, offline/cache, rollback, and no-Maintenance packcompat gates.
3. Define the required Core loading boundary and typed participation model,
   then inventory and gate every production config-loading surface before
   implementation beads depend on it.
4. Choose the public Gastown to host Core/Dog model and write the host-Core
   worker reference contract, including route/pool/mail/nudge/patch/prompt
   behavior under public-pack bindings.
5. Make the behavior manifest, public Gastown replacement commit, consuming
   pin, and execution-level packcompat witness matrix a first-slice blocking
   deliverable.
6. Replace per-check doctor mutation with a staged coordinator contract and
   failure-injection test plan covering controller-active cities, concurrent
   doctors, partial publishes, custom forks, offline pins, and downgrade
   states.
7. Centralize retired-source handling across imports, locks, caches,
   materialization, config load, and prompt/template discovery, preserving old
   bytes while preventing active participation.
8. Resolve `GC_BOOTSTRAP=skip`, implicit-import cleanup, bootstrap fixture
   isolation, and hidden dependency inventory before removing bootstrap Core
   embeds or managed names.
9. Move docs, wording, runtime-state guidance, examples, generated references,
   and tutorial goldens into the slices that change the corresponding behavior
   or enforce a non-release gate.
10. Fix the design-review workflow artifact path bug, or have global synthesis
    record/copy the exact source persona syntheses it consumed before future
    approval iterations rely on the artifacts.
