# 09 Rollout Decomposition Gates - Codex Review

Persona: Iris Kowalski

Verdict: approve with one required decomposition change. The rollout is much
closer to taskable than a typical migration plan: it names slices, gates,
rollback boundaries, public-pack prerequisites, and broad test commands. The
remaining risk is that the missing proof/generator work is described as a
policy but not yet cut as its own concrete pre-implementation slice.

## Findings

### Required: add a concrete proof-foundation slice before Slice 1a

The plan correctly states that, before full Gas City implementation
decomposition, only external-prerequisite and proof-producing beads may be
created: public `gascity-packs` ownership/pin work, support-artifact generators
and validators, non-mutating scans, packcompat harness work, and docs/schema
golden generation (`implementation-plan.md:692`). It also correctly says source
deletion, Maintenance removal, activation-pin consumption, runtime-state
mutation, automatic doctor repair, and behavior-changing loader cutover wait
until AC6, AC7, AC14, AC15, AC16, and AC17 exist and pass
(`implementation-plan.md:696`).

The slice plan then starts at Slice 1a, but Slice 1a already assumes proof
infrastructure exists: the slice-to-gate table says Slice 1a may start when the
AC6 ledger generator exists (`implementation-plan.md:797`). Similar prerequisites
exist for AC15/AC16 cache schema proof before Slice 2
(`implementation-plan.md:800`) and AC3/AC11 evidence before Slice 4a
(`implementation-plan.md:802`). The support directory currently contains only
`maintenance-asset-classification.md`, not the binding support artifacts named
in the plan. That means the first real decomposition step is missing: who builds
the generators, validators, schemas, proof matrix, and deterministic harnesses
that make the later slices legal.

Required change: add an explicit "Slice 0: proof foundation" or equivalent
pre-implementation task group. It should create and validate the missing support
artifacts and gates before runnable implementation slices are created:
`asset-migration-ledger.yaml`, `behavior-preservation-manifest.yaml`,
`pack-resolution-matrix.yaml`, `source-consumer-closure.yaml`,
`role-neutrality-scan.yaml`, `migration-diagnostics.schema.json`,
`coverage-transfer.yaml`, `public-gastown-pin-ledger.yaml`,
`version-skew-matrix.yaml`, and `acceptance-proof-matrix.yaml`
(`implementation-plan.md:554`). It should also name the first executable
commands or make targets that enforce AC17 before any task plan can schedule
Slices 1a-7 as runnable work.

For `tasks.md`, create only Slice 0 and external `gascity-packs` prerequisite
beads as runnable work at first. Later implementation slices can be recorded as
blocked/deferred placeholders only if their dependencies explicitly point at the
validated AC17 proof matrix and the required public-pack commit/digest/transcript
beads. That prevents the main failure mode for this migration: a task graph that
contains the right prose gates but still makes pin changes, source deletion,
doctor mutation, docs updates, and activation work ready too early.

## What Is Solid

- The implementation slices are directionally good and independently reviewable:
  public ownership/proof, compatibility pin, Core extraction, systempacks loader,
  doctorfix, runtime-state migration, activation candidate, Maintenance fold,
  registry/cache cleanup, and stale source/docs deletion are separate enough to
  reason about (`implementation-plan.md:724`, `implementation-plan.md:739`,
  `implementation-plan.md:747`, `implementation-plan.md:752`,
  `implementation-plan.md:757`, `implementation-plan.md:761`,
  `implementation-plan.md:766`, `implementation-plan.md:772`,
  `implementation-plan.md:780`, `implementation-plan.md:787`).
- The plan names rollback or recovery behavior for each major slice instead of
  treating the migration as a single irreversible batch (`implementation-plan.md:724`,
  `implementation-plan.md:729`, `implementation-plan.md:739`,
  `implementation-plan.md:747`, `implementation-plan.md:752`,
  `implementation-plan.md:757`, `implementation-plan.md:772`,
  `implementation-plan.md:780`, `implementation-plan.md:787`).
- Test depth is not limited to fast unit tests. The plan calls out focused unit
  packages, packcompat, packlint, doctor/runtime-state tests, docs/golden
  freshness, `make test-fast-parallel`, `go vet ./...`, and higher-risk process
  and integration shard targets (`implementation-plan.md:603`,
  `implementation-plan.md:639`, `implementation-plan.md:657`,
  `implementation-plan.md:672`, `implementation-plan.md:681`).

## Residual Risk

When tasks are created, every runnable bead should state its first allowed gate,
its exact affected files, its rollback boundary, and whether it is pre-gate
proof work or post-gate behavior-changing work. Any bead that mutates runtime
state, removes a source root, updates `PublicGastownPackVersion`, changes
required-pack loading, or edits doctor fix behavior should be impossible to run
until the proof-foundation slice and public-pack prerequisite beads have closed.
