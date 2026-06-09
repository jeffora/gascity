# Design Review Synthesis

## Overall Verdict: block

Attempt 14 is a clear block under worst-verdict-wins: nine persona syntheses
returned `block`, and the remaining behavior-preservation lane returned
`approve-with-risks`. The design has the right high-level direction, but the
current text still leaves load-bearing contracts unresolved across required
Core loading, doctor mutation safety, role neutrality, public Gastown ownership,
cache/version-skew behavior, and executable rollout gates.

## Consensus Strengths

- The host-Core/public-Gastown split is the right ownership direction: Core
  should own required SDK infrastructure while public Gastown owns
  Gastown-specific roles, prompts, formulas, orders, scripts, and docs.
- The generated artifact approach is strong in principle. Multiple personas
  praised the behavior manifest, role-surface table, wording matrix,
  test-migration map, public pin ledger, and packcompat witnesses as the right
  way to prevent narrative-only migration claims.
- Required system-pack participation is the right invariant. Reviewers agree
  Core must be materialized, validated, and proven to participate in the final
  resolved configuration rather than merely existing on disk.
- The doctor coordinator, retired-source classifier, and public pin ledger are
  promising mechanisms for safe migration, provided their exact mutation,
  lock, provenance, and rollback semantics are made executable.
- Removing production Core bytes from `internal/bootstrap` and replacing broad
  copied fixtures with synthetic test fixtures is directionally correct.
- The design correctly treats duplicate active definitions, stale local pack
  paths, and behavior loss during Maintenance retirement as release-blocking
  risks rather than cleanup work.

## Critical Findings

### [Blocker] Required Core participation and loud-failure semantics are not executable

**Sources:** 02-required-core-loading-invariant-reviewer, 03-doctor-migration-safety-reviewer, 06-pack-registry-and-cache-tester

**Issue:** The design promises required Core participation, but it does not
define a trustworthy resolver-produced provenance substrate with stable layer
IDs, import edges, file-set digests, and behavior for implicit includes.
Missing, corrupt, partial, stale, and unexpected-file Core states also have
inconsistent repair versus fail-closed behavior across normal loads,
controller reload, no-refresh paths, doctor diagnostics, and provider-pack
selection.

**Required change:** Specify typed required-pack participation as a
post-resolution assertion tied to validated file-set digests. Add a
loader-class by failure-class matrix for materialization, validation, repair,
read-only diagnostics, and fatal failure, and make every production config
resolution path either enforce it or carry a documented partial-read exception
with focused tests.

### [Blocker] Plain `gc doctor` can still mutate the state it is diagnosing

**Sources:** 03-doctor-migration-safety-reviewer, 02-required-core-loading-invariant-reviewer, 09-docs-dx-consistency-reviewer

**Issue:** The Core-presence doctor path is still described in terms that can
route through materializing runtime loaders. That can repair, quarantine, or
rewrite Core before reporting, bypassing the intended `--fix` coordinator,
advisory lock, controller-active refusal, and operator-facing diagnostic
contract.

**Required change:** Make plain doctor checks use validate-only APIs such as a
no-refresh runtime load or direct required-file-set validation. All mutation,
repair, import rewrite, prune, quarantine, lockfile, and migration-marker work
must happen only through an explicit `--fix` path governed by one coordinator
lock protocol and failure-injection tests.

### [Blocker] Role neutrality, dog ownership, and public Gastown boundaries remain unresolved

**Sources:** 04-zfc-role-neutrality-guardian, 08-gastown-pack-boundary-reviewer, 01-behavior-preservation-auditor

**Issue:** The design still lacks an enforceable disposition for active role
surfaces in production Go, Core/provider TOML, prompts, orders, formulas,
scripts, metadata, dashboard/API wire values, and public Gastown assets. Dog
prompt preservation is especially unresolved: current behavior relies on
same-name template fragment shadowing, while the proposed Core/public split
forbids the shadowing mechanism without replacing it with a concrete patch,
override, or ownership model.

**Required change:** Add a role-surface migration table with exact owners,
allowed fields, expiry rules, and tests for Go identifiers, TOML defaults,
step descriptions, mail/nudge targets, route metadata, prompt fragments,
provider-pack escalation recipients, `dog` pool behavior, and API/dashboard
`crew` vocabulary. Decide whether public Gastown imports Core, relies on
host-auto-included Core plus patch layering, or owns dog-related assets itself,
then prove that model with rendered prompt and packcompat tests.

### [Blocker] Retired Maintenance can still re-enter active resolution

**Sources:** 08-gastown-pack-boundary-reviewer, 06-pack-registry-and-cache-tester, 04-zfc-role-neutrality-guardian

**Issue:** Preserving stale `.gc/system/packs/maintenance` or local Gastown
bytes is safe only if production resolution, prompt discovery, cache
resolution, import checks, lock validation, and install paths cannot activate
them. The design still leaves room for city or pack TOML imports, transitive
imports, stale synthetic aliases, or prompt/template discovery to load retired
generated pack bytes.

**Required change:** Define retired-pack rejection as an active config
resolution invariant, not a doctor-only cleanup. `gc doctor --fix` should
rewrite or remove retired imports where safe, but normal production loading
must reject retired generated pack names and paths even when stale directories
remain on disk.

### [Blocker] Public pin, cache, offline, and old-binary behavior are under-specified

**Sources:** 05-rollout-version-skew-planner, 06-pack-registry-and-cache-tester, 10-test-slicing-coverage-verifier

**Issue:** The rollout cannot yet prove that old binaries load the pinned
public Gastown contents rather than embedded synthetic alias bytes. It also
does not define cache-key compatibility for old synthetic caches, ordinary
remote caches, subpaths, stale locks, offline upgrades, activation assets that
older binaries may not parse, or the existing shipped public pin
`sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`.

**Required change:** Split the rollout matrix by current/baseline pin,
compatibility pin, and activation pin. For each phase, state old-binary,
new-binary, rollback, stale synthetic cache, ordinary cache, and offline
behavior, and require observed content digests rather than load success alone.
Either keep activation old-binary-compatible or declare a named one-way
boundary with release notes, doctor diagnostics, manual recovery, and tests.

### [Blocker] Rollout gates and slice ordering still contradict each other

**Sources:** 10-test-slicing-coverage-verifier, 05-rollout-version-skew-planner, 03-doctor-migration-safety-reviewer

**Issue:** The slice-gate matrix requires process and integration coverage for
high-risk Core loading, doctor, and registry/cache cleanup slices, while the
numbered rollout omits those gates. Slice 5 also bundles activation-pin
flipping, Maintenance removal, asset moves, runtime-state retirement, and
doctor updates without commit-level ordering that keeps intermediate states
green and avoids duplicate or missing active definitions.

**Required change:** Make one checked-in per-slice gate artifact the binding
source of truth, including exact commands, package targets, `-run` filters,
fixtures, old/new binary expectations, and generated artifacts. Specify
commit-level ordering for the activation slice so each intermediate commit has
exactly one active copy of each behavior and passes the required fast gate.

### [Blocker] Behavior preservation evidence is still mostly promised, not produced

**Sources:** 01-behavior-preservation-auditor, 08-gastown-pack-boundary-reviewer, 10-test-slicing-coverage-verifier

**Issue:** Reviewers agree the generated manifest and packcompat plan are the
right approach, but the design still lacks the first concrete artifacts and
pilot rows for fragile behavior: shutdown-dance requester/detector semantics,
spawn-storm escalation, DOG_DONE nudges, JSONL/reaper notifications, Polecat
handoffs, branch pruning, review checks, commands, doctor scripts, prompt
fragments, session-live hooks, and Gastown archive authorship.

**Required change:** Make the first implementation slice generate and validate
the behavior manifest, independent old-tree baseline, role-surface table,
test-migration map, wording matrix, and public pin ledger before source moves,
role-neutral rewrites, public pin updates, registry/cache retirement, docs
claims, or Maintenance removal.

### [Major] Bootstrap fixture isolation still lacks a concrete seam

**Sources:** 07-bootstrap-fixture-isolation-reviewer, 06-pack-registry-and-cache-tester

**Issue:** The design is directionally correct on removing production Core
from `internal/bootstrap`, but it leaves unresolved how non-bootstrap tests
that currently depend on `bootstrap.BootstrapPacks`, `GC_BOOTSTRAP=skip`, or
embedded Core bytes should be migrated. It also contains contradictory fixture
language around preserving `packs/core` while rejecting lingering
`packs/core` references.

**Required change:** Limit bootstrap asset injection to unexported `_test.go`
helpers in `internal/bootstrap`; move other tests to cache/system-pack setup or
lower-level helpers. Delete the production `GC_BOOTSTRAP=skip` branch or
narrow it with tests proving it cannot bypass required Core materialization or
retired-entry pruning. Use visibly synthetic fixture paths such as
`packs/test-core`.

### [Major] Docs and operator wording gates are not release-ready

**Sources:** 09-docs-dx-consistency-reviewer, 05-rollout-version-skew-planner, 08-gastown-pack-boundary-reviewer

**Issue:** The docs inventory command is narrower than the stated contract,
the canonical `docs/reference/system-packs.md` deliverable is not established
in the current tree, and runtime-state troubleshooting guidance cannot be
correct until Core write paths, legacy read paths, doctor behavior, rollback,
and operator examples are decided. The wording matrix also lacks a concrete
schema, owner, generator, freshness test, public Gastown companion check, and
allowed-context rows.

**Required change:** Define a generated docs/wording linter covering Markdown,
MDX, JSON, TXT, docs navigation, generated references, schema files, examples,
scripts, CLI/help output, doctor output, public Gastown docs, and tutorial
transcripts. Create or rewrite the system-pack reference page, register it in
navigation, and sequence docs gates with every behavior-changing slice.

### [Major] Provider-pack continuity and role-cleaning conflict

**Sources:** 06-pack-registry-and-cache-tester, 04-zfc-role-neutrality-guardian, 01-behavior-preservation-auditor

**Issue:** The design wants Core repair to preserve `bd` and `dolt` bytes and
provenance, but it also places provider-pack role surfaces such as `dog`,
`mayor`, and `deacon` notification targets in the role-cleaning scope. Both
cannot be true unless the design names which provider assets are intentionally
rewritten or which role-bearing values are accepted as bounded compatibility
exceptions.

**Required change:** Add a provider-pack rewrite or compatibility-exception
table with old/new witnesses, expiry, provenance, and Core-repair isolation
tests proving untouched provider assets remain unchanged.

### [Minor] Several cross-lane decisions are still hidden in prose

**Sources:** 03-doctor-migration-safety-reviewer, 07-bootstrap-fixture-isolation-reviewer, 09-docs-dx-consistency-reviewer

**Issue:** Important operator and implementation details remain ambiguous:
migration-completed marker path/schema, lock hierarchy, multi-file publish
rollback versus recoverability, stale staging cleanup, moved-order qualifier
or alias behavior, public Gastown patch semantics, and schema/reference
regeneration ownership.

**Required change:** Promote these from prose to explicit tables with owners,
states, failure modes, and tests so implementation beads cannot make different
local choices.

## Disagreements

- Behavior preservation was the only non-blocking persona verdict:
  `approve-with-risks`. Its stricter concerns still align with the blockers
  from other lanes: generated artifacts, pilot behavior rows, and witness tests
  must precede destructive migration work.
- Several reviewers accept the high-level rollout shape, while rollout and
  test-slicing reviewers block on executable phase details. The synthesis
  position is that the concept is sound, but the design is not dispatchable
  until the per-slice gate artifact and commit-level activation ordering are
  binding.
- Reviewers differ on whether activation should remain compatible with
  `v1.2.1` or become a one-way boundary. Either can be acceptable, but the
  design must choose one and test the chosen behavior with observed public-pack
  content digests.
- Provider-pack continuity can be resolved either through narrow witnessed
  rewrites or formal compatibility exceptions. What is not acceptable is
  promising byte-for-byte continuity while also requiring role-cleaning of
  role-bearing provider assets.
- Retired bytes may be preserved on disk for rollback and operator safety.
  The consensus requirement is that preservation must not mean active
  importability, prompt discovery, cache selection, or lock acceptance.
- Reviewers differ on exact dog ownership: configurable Core worker, public
  Gastown-owned dog assets, or host-Core dog plus public Gastown patching. The
  design must pick one model and make prompt rendering, patch resolution,
  routing, and omitted-worker behavior follow from it.

## Missing Evidence

- Typed required-system-pack participation producer, schema, layer identity,
  digest source, implicit-include behavior, and materialized-but-not-resolved
  failure tests.
- Complete production loader inventory and scanner fixtures for direct
  `config.Load`, wrapper helpers, no-refresh paths, raw TOML decoding,
  manually assembled include lists, aliases, and exported config resolvers.
- Read-only doctor call path, unified lock protocol, failure-after-each-target
  publish tests, migration marker schema, and legacy direct-fix disablement
  gate.
- Exact public Gastown dependency model, dog prompt patch/override mechanism,
  rendered prompt goldens, and stale retired-import rejection fixtures.
- Phase-specific public pin ledger, current-pin baseline row, durable-ref CI
  rule, old-binary content-digest evidence, and offline/cache promotion
  contract.
- Generated behavior manifest, role-surface table, wording matrix,
  test-migration map, public pin ledger, and first pilot rows for the highest
  risk behavior moves.
- Checked-in per-slice gate artifact with exact commands, fixtures,
  old/new-binary expectations, generated artifact ownership, and slice 5
  commit ordering.
- Bootstrap hidden-dependency inventory covering non-bootstrap tests,
  `GC_BOOTSTRAP`, `internal/materialize`, `internal/doctor`,
  `cmd/gc` tests, source-path readers, and stale docs.
- Runtime-state migration table for JSONL archives, export state,
  push-failure counters, spawn-storm state, order tracking, read/write paths,
  doctor behavior, fix behavior, rollback, and operator commands.
- Public Gastown companion docs/comment lint and tutorial transcript tests
  proving retired local pack paths and stale ownership wording do not surface.

## Recommended Changes

1. Make generated evidence the first slice: behavior manifest, old-tree
   baseline, role-surface table, wording matrix, test-migration map, loader
   inventory, and public pin ledger with pilot rows.
2. Specify required-pack participation and failure semantics as typed,
   resolver-produced, post-resolution assertions with a loader/failure matrix.
3. Rewrite doctor semantics so plain diagnostics are read-only and every
   mutation goes through one coordinator lock and an explicit `--fix` path.
4. Resolve public Gastown/Core dependency and dog ownership before moving or
   deleting Maintenance assets.
5. Enforce retired-pack rejection in production resolution and discovery paths,
   while preserving bytes only as inactive rollback/operator evidence.
6. Split public pin rollout by current, compatibility, and activation phases,
   with old-binary content-digest evidence and explicit offline/cache behavior.
7. Convert rollout gates into a checked-in per-slice artifact and define
   activation-slice commit ordering that stays green at every intermediate
   commit.
8. Decide provider-pack role-cleaning versus compatibility exceptions, then
   add old/new witnesses and Core-repair isolation tests.
9. Finish bootstrap fixture isolation by removing production Core embeds,
   constraining fixture injection, and eliminating or tightly narrowing
   `GC_BOOTSTRAP=skip`.
10. Land the docs/wording/runtime-state contracts early enough that doctor
    output, tutorials, schema/reference docs, public Gastown comments, and
    troubleshooting guidance cannot ship in contradictory intermediate states.
