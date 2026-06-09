# Iris Kowalski - Claude

**Verdict:** block

> Lane: independently deployable slices, decomposition readiness, prerequisite
> honesty, exact gates, cross-repo sequencing and test coverage. Reviewed against
> the current `implementation-plan.md` (528 lines,
> `updated_at: 2026-06-09T01:20:00Z`) — §"Rollout And Recovery" (453–518),
> §"Testing" (393–451), §"Open Questions" (520–528) — and the parent
> `requirements.md`. I verified prerequisite existence rather than trust the
> prose. The design improved real things since iteration 1, but the lead
> prerequisite-honesty Blocker stands on unchanged facts.
>
> Output written to the live iteration-3 reviews dir (`attempt-3/`) beside the
> Codex sibling; my routed bead `ga-9ep31n` carries `gc.attempt=1` while its
> iteration is 3 (logical `ga-s86v4z` is `attempt=3`), so the literal
> `attempt-${gc.attempt}` path would overwrite the unrelated iteration-1 review.
> Reported via `design_review.output_path`.

**Schema conformance:** Structurally conforms (front matter, eight ordered
sections, `Open Questions: None`). But "Decomposition Readiness" in the schema
requires open questions to be genuinely `None`; see Blocker 1 — the artifact
satisfies the letter while the parent requirements remain `status: questions`.

**Top strengths:**
- Slicing discipline on the dangerous edges: source deletion (Slice 7) and docs
  (Slice 7) are separated from activation (Slice 5), and the one-way upgrade
  boundaries are explicitly flagged — Slice 4 reverts "only if no mutation
  coordinator state has committed" (477–479) and Slice 6 is one-way "after
  activation has shipped" (493–496). Every slice names a concrete rollback.
- Gates exceed fast-unit for high-risk work (answering red flag #3):
  loader/doctor/runtime-state slices "also run the sharded process and
  integration targets … `make test-cmd-gc-process-parallel` and
  `make test-integration-shards-parallel`" (447–451), and packcompat exercises
  production loaders from the pinned public checkout via the ordinary remote-pack
  path (162–166), not copied fixtures or `config.Load` shortcuts.
- The release compatibility matrix (504–512) and the runtime-state safety
  contract (old-binary post-marker write detection + version-skew diagnostic,
  266–273) give the cross-binary story real shape it lacked before.

**Critical risks:**

- **[Blocker] Not prerequisite-honest / not decomposition-ready.** The plan
  asserts `Open Questions: None` (522) at `status: draft` (front matter:8) while
  the parent `requirements.md` is `status: questions` (requirements:7) with five
  unresolved questions (requirements:120–134), and the AC6/AC7 artifacts the
  rollout gates depend on do not exist — I confirmed `test/packcompat` is absent,
  `plans/core-gastown-pack-migration/artifacts/` is absent, and there is no
  asset-migration ledger, `behavior-preservation.yaml`, or pin ledger anywhere in
  the plan tree (`tasks.md` is also absent — decomposition has not occurred).
  Reclassifying these as "external prerequisites … not open design questions"
  (524–528) is a categorization, not a resolution: requirements Q1 (who owns the
  ledger) and Q2 (what command produces the manifest and where it lives) are
  unresolved ownership/interface **design** decisions that gate Slice 1 and every
  behavior-evidence-dependent slice. This is red flag #2 verbatim — "status says
  decomposition-ready while required generators or public pack commits do not
  exist." It was the iteration-1 Blocker and the underlying facts are unchanged.

- **[Major] Slice 1's cross-repo prerequisite is not executable as beads.** It
  says "Land the `gascity-packs` branch with moved … assets,
  behavior-preservation manifest, pin ledger, replacement tests, ownership rows,
  and host-Core/no-Maintenance proof" (455–459; expanded 104–128) but never names
  the owning repo plan, the branch/PR, the compatibility and activation commit
  SHAs, the artifact paths, the owner, the proof commands, or the dependency
  edges into Gas City slices. A decomposer cannot create or order Slice-1 beads,
  and Gas City Slices 2/5 dangle on an unidentified external commit.

- **[Major] Slices 4 and 5 batch multiple failure domains.** Slice 4 bundles
  `internal/systempacks` + pre/post-resolution validation + typed participation +
  the loader scanner + allowlist generation + Core doctor + pre-resolution
  recovery + version-skew diagnostics + the mutation coordinator (474–479). Slice
  5 bundles the activation pin + removing Maintenance from
  `requiredBuiltinPackNames` + moving Core-owned Maintenance assets + consuming
  Gastown assets from the public pack + the packcompat mode change "in the same
  candidate branch" (481–489). The activation flip (pin + required-pack change) is
  irreducibly atomic, but the mechanically independent parts — the systempacks
  API vs the scanner/allowlist vs the coordinator (Slice 4); the Core-owned asset
  moves (Slice 5) — should be separable sub-slices with stable intermediate
  contracts, or the atomicity must be justified (red flag #1, narrowed: deletion
  and docs are correctly excluded from these batches).

- **[Major] Runtime-state migration has a safety contract but no slice
  placement.** The JSONL/ledger/refs/escalation migration is specified (266–273,
  377–381) yet no Slice step executes it, and the sequencing between
  controller-startup migration and the coordinator's "refuse if a controller is
  running" rule (257) is unresolved — as is the in-flight-session policy
  (requirements Q4). A decomposer cannot place this work or its gates.

- **[Major] Proof gates are not bound to each slice's acceptance.** §Testing
  lists test groups and §Rollout lists slices, but no slice states which exact
  commands, artifact IDs, old/new-binary checks, offline/cache checks, and
  rollback proofs gate ITS merge. This invites later beads to under-test
  loader/doctor/runtime-state/cross-repo moves (red flag #3 at slice granularity).

**Missing evidence:**
- Resolutions (or honest reinstatement as Open Questions) for requirements Q1–Q5
  (ledger ownership, manifest-producing command + location, version-skew window,
  in-flight-session policy, missing-Core repair command).
- The `gascity-packs` repo plan, branch/PR, compatibility + activation commit
  SHAs, artifact paths, owner, and proof commands referenced by Slices 1/2/5.
- A per-slice acceptance map: exact gate commands and proofs that block each
  slice's merge.
- The slice that executes runtime-state migration and its ordering relative to
  controller liveness.
- Any decomposition artifact (`tasks.md`) demonstrating the slices actually cut
  into independently mergeable beads with revert/one-way boundaries.

**Required changes:**
- Either resolve requirements Q1–Q2 (ledger + manifest ownership and the
  producing command/location) in-plan, or move them back into Open Questions and
  set the plan to `blocked:prerequisite`; do not assert `Open Questions: None`
  while they are open and the artifacts are absent.
- Make Slice 1 executable: name the owning repo plan/branch, the compatibility
  and activation SHAs, artifact paths, owner, proof commands, and the dependency
  edges into Gas City slices.
- Split Slices 4 and 5 into sub-slices with stable intermediate contracts;
  isolate the irreducible activation flip from the mechanically independent
  moves.
- Place runtime-state migration in the timeline, resolve the
  controller-liveness/coordinator ordering, and decide the in-flight-session
  policy.
- Bind exact proof gates (commands, artifact IDs, old/new-binary, offline/cache,
  rollback) to each slice's acceptance.

**Questions:**
- Are the AC6/AC7 ledger and manifest Gas City beads, a separate `gascity-packs`
  plan, or external prerequisites — and who owns each? Until this is answered,
  Slice 1 cannot be decomposed.
- What are the compatibility and activation commit SHAs, and does a real
  `gascity-packs` branch exist today?
- Which slice runs the runtime-state migration, and does it run before the
  controller is live (avoiding the coordinator's running-controller refusal)?
