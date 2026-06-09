# Design Review Synthesis

## Overall Verdict: approve-with-risks

Attempt 13 reviewed the same design text as Attempt 12. The architecture is
directionally sound and the host-Core/public-Gastown model is now coherent, but
the document still contains contradictory rollout, doctor, provider-pack, and
fixture contracts that must be cleaned up before decomposition.

## Consensus Strengths

- The host-Core/public-Gastown split is the right ownership model: Core owns
  required infrastructure and public Gastown owns Gastown-specific behavior.
- The `core.maintenance_worker` binding removes the previous hard dependency
  on a concrete `dog` agent while preserving an override point for explicit
  configurations.
- Required Core loading now has the right shape: deny-by-default system-pack
  materialization, strict integrity checks, and typed
  `RequiredSystemPackParticipation` proof after normal resolution.
- The retired-source classifier and public pin ledger are strong mechanisms for
  avoiding stale local paths, duplicate active definitions, and synthetic-cache
  shadowing.
- The generated behavior manifest and witness floor are the right gates for the
  user invariant that Gastown must not lose behavior.

## Critical Findings

### [Blocker] Slice 5 still has a circular no-Maintenance gate

**Sources:** rollout-version-skew-planner, test-slicing-coverage-verifier

**Issue:** The rollout text says to update `PublicGastownPackVersion` to the
activation commit and run the no-Maintenance production-loader gate before
removing Maintenance from required packs. That intermediate state can neither
load without Maintenance nor avoid duplicate active definitions.

**Required change:** Make activation pin adoption and Maintenance removal one
candidate slice. Run the no-Maintenance production-loader gate on that combined
tree before merge.

### [Blocker] Provider pack de-roling contradicts byte-continuity

**Sources:** behavior-preservation-auditor, pack-registry-and-cache-tester

**Issue:** Required provider packs are in the role-surface cleanup scope, but
the design also says `bd` and `dolt` bytes remain unchanged except for manifest
metadata. Existing provider formulas contain hardcoded role targets, so both
requirements cannot hold.

**Required change:** Permit narrow provider-pack asset rewrites for role
cleaning and target-binding conversion, with behavior witnesses and provenance
records. Keep byte-continuity only for assets not intentionally rewritten.

### [Blocker] Runtime-state copy semantics still create a conflict loop

**Sources:** doctor-migration-safety-reviewer, rollout-version-skew-planner

**Issue:** The design copies legacy Maintenance state to Core but then refuses
automatic fixes when both sides exist and digests differ. Healthy upgraded
cities will diverge as Core writes live state, causing repeated false conflicts.

**Required change:** Record a migration-completed/publish marker. After it
proves Core was initialized from the legacy source, legacy divergence is ignored
as retained rollback state instead of treated as a conflict.

### [Major] Required-pack tamper recovery can deadlock

**Sources:** pack-registry-and-cache-tester, doctor-migration-safety-reviewer

**Issue:** Strict file-set validation fails on unexpected files, but repair
semantics only cover missing or corrupt expected files. An injected unexpected
file could make every subsequent startup fail.

**Required change:** Define required-pack repair as prune-or-quarantine of
unexpected effective files before validation, preserving operator-edited files
in a recovery location when needed.

### [Major] Doctor mutation protocol has ordering and rollback gaps

**Sources:** doctor-migration-safety-reviewer

**Issue:** The coordinator acquires its advisory lock after preflight, creating
a stale-preflight race. It also lands after public-pin adoption in the rollout,
leaving an unsafe legacy fix window.

**Required change:** Acquire the lock before preflight or require post-lock
digest revalidation before staging. Land the safe coordinator before any slice
that exposes operators to import rewrites, or freeze legacy auto-fixes until it
lands.

### [Major] Offline/cache semantics need one explicit rule

**Sources:** behavior-preservation-auditor, doctor-migration-safety-reviewer,
rollout-version-skew-planner, pack-registry-and-cache-tester

**Issue:** The design forbids synthetic fallback for public Gastown, but also
requires offline upgrade and hermetic CI behavior. Fresh remote fetches cannot
be mandatory for every test or repair path.

**Required change:** Treat reachability as satisfied by either the network
source or a validated ordinary remote-cache entry for the exact pinned commit.
Add a cache-promotion path for valid legacy synthetic public Gastown caches, and
make `test/packcompat` runnable from a local fixture/cache.

### [Major] Bootstrap fixture and hidden-dependency wording still need cleanup

**Sources:** bootstrap-fixture-isolation-reviewer, test-slicing-coverage-verifier

**Issue:** The design still leaves room for copied `testdata/packs/core`
fixtures, empty embed leftovers, and unlisted production path reads such as
`cmd/gc/prompt_test.go`.

**Required change:** Require deletion of the `//go:embed packs/**` directive
and `embeddedBootstrapPacks`, use synthetic inline fixture paths such as
`packs/test-core`, and keep the hidden-dependency inventory explicit.

## Disagreements

- Some reviewers described stale system-pack directories as preserved in place;
  others recommended renaming them to a retired backup namespace. The stronger
  operational contract is preserve bytes but isolate from active discovery
  paths, using either a retired namespace during doctor fix or a classifier that
  proves active discovery cannot see them.
- The rollout reviewer called the runtime-state migration destructive, while
  the current design already moved toward copy semantics. The remaining problem
  is not the copy itself; it is the post-copy conflict rule and rollback marker.
- Provider pack continuity was praised by some reviewers and challenged by
  others. The reconciliation is targeted, witnessed rewrites for role-cleaning
  only, with byte-continuity for all untouched provider assets.

## Missing Evidence

- A concrete provider-pack role-cleaning exception table for `bd` and `dolt`.
- A migration-completed record and test proving repeated `gc doctor --fix` does
  not report false conflicts after Core state diverges.
- A hermetic `test/packcompat` mode using a local fixture or pre-populated
  ordinary remote cache.
- Specific process/integration shard names for the highest-risk rollout slices.
- An explicit unexpected-file quarantine/prune contract for required packs.

## Recommended Changes

1. Consolidate the latest resolution contracts into the main design sections and
   remove stale contradictory wording instead of appending another attempt
   appendix.
2. Rewrite Slice 5 so activation pin adoption and Maintenance removal are one
   candidate change with the no-Maintenance gate run on that candidate.
3. Add a provider-pack role-cleaning exception to the byte-continuity rule.
4. Update runtime-state migration to copy, record completion, and ignore
   expected post-migration divergence from retained legacy state.
5. Define unexpected-file pruning/quarantine for required Core/provider repair.
6. Clarify offline success through validated ordinary remote cache or local
   packcompat fixture, not network-only reachability.
7. Add exact sharded test gates and hidden-dependency cleanup items before
   engineering decomposition.
