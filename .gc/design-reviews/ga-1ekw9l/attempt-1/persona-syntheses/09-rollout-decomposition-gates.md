# Iris Kowalski

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Full behavior-changing implementation decomposition is not ready. All three reviewers agree the plan may proceed only with prerequisite-producing work until AC6, AC7, and AC14-AC17 proof artifacts exist, validate, and are cited by immutable commit or checked path. Codex found no local `support/` or `artifacts/` directory; Claude and DeepSeek both flag that the readiness boundary still needs a named executable enforcer before bead creation.
- [Blocker] AC17 remains paper-only without a concrete command, pre-commit hook, or CI job. The plan says the acceptance-proof matrix gates decomposition, but it does not name the executable target that validates `support/acceptance-proof-matrix.yaml` and blocks behavior-changing beads when required evidence is missing.
- [Blocker] Slice 2's start gate omits the Slice 1b compatibility-pin dependency. The table allows Slice 2 to start when AC15/AC16 cache schema proof exists, but Slice 2 consumes a public compatibility commit that is produced by Slice 1b. Starting Slice 2 without an immutable Slice 1b pin would schedule work against an unavailable prerequisite.
- [Blocker] Cross-repo prerequisites are not executable enough for rollout. Codex requires an external-prerequisite ledger with owner, repository, branch or PR, immutable commit, artifact path, digest, validation command, consuming slice, and missing-artifact policy. DeepSeek adds that the no-Maintenance activation transcript can deadlock across `gascity-packs` and Gas City unless the plan defines a staging/bootstrap workflow for candidate loader changes before immutable public commits exist. Claude also asks for a source of truth for commit availability.
- [Major] Several rollout slices are too broad to be independently deployable without further splitting or exact file/gate proof. Codex treats Slices 2, 3, 5b, and 6 as multi-failure-domain landings; DeepSeek specifically requires splitting 5b and 7; Claude calls 5b and 7 dense candidates for further splitting if their merge gates prove heavy. The shared risk is that broad slices hide rollback and sequencing decisions inside large beads.
- [Major] Rollback, downgrade, and version-skew behavior is not operational. Reviewers ask for state markers, schema versions, downgrade behavior, recovery commands, and tests covering old/new binaries against old/new lock schemas, cache layouts, mutated `city.toml`, offline exact cache hits, stale synthetic aliases, interrupted doctor migration, and post-activation downgrade.
- [Minor] The artifact lacks structured slice-level readiness state. Front matter and "Open Questions: None" can be misread as full decomposition readiness even though the plan intends prerequisite-only decomposition now and behavior-changing slices only after gates pass.

**Disagreements:**
- Claude blocks on two narrow exact-gate defects plus the requirements' explicit approval constraint, while Codex blocks on broader decomposition granularity, absent proof artifacts, and non-executable external prerequisites. Assessment: the narrower and broader findings are compatible; full implementation decomposition remains blocked, while prerequisite-producing beads are acceptable.
- Claude says the slice architecture is strong and treats 5b/7 density as minor unless gates prove heavy. Codex says Slices 2, 3, 5b, and 6 are already too large; DeepSeek requires splitting 5b and 7. Assessment: the plan should either split the named dense slices or give each a concrete file set, proof command, prerequisite ledger entry, and rollback boundary strong enough to be independently revertible.
- DeepSeek uniquely identifies a cross-repository staging deadlock between no-Maintenance transcript generation and Gas City loader support. Assessment: this is credible rollout signal even without full consensus because both other reviewers also found external prerequisite handling insufficient.
- DeepSeek uniquely requires explicit sharded process and integration coverage for Slices 5b and 6. Assessment: because those slices change required builtin packs, cache keys, alias handling, and runtime state, the extra coverage should be carried forward.

**Missing evidence:**
- Checked-in support artifacts such as `support/acceptance-proof-matrix.yaml`, validators, generators, schemas, and passing proof outputs.
- A named executable readiness gate, such as a make target, pre-commit check, or CI job, that validates the acceptance-proof matrix before bead creation.
- An external-prerequisite ledger for `gascity-packs` commits, generated manifests, pin ledgers, ownership rows, packcompat transcripts, artifact digests, validation commands, owners, and failure policies.
- Bootstrap or staging commands that prove no-Maintenance loading against candidate Gas City loader changes before final immutable public-pack commits are consumed.
- Exact criteria for whether doctor-mutated manifests remain readable by old binaries or require documented downgrade limits.
- Programmatic downgrade tests for a legacy `gc` binary against doctor-mutated `city.toml` and changed lock/cache state.
- Per-slice file sets, dependency edges, start gates, merge gates, and rollback or one-way-boundary predicates for the densest slices.
- Explicit confirmation that Slice 2 disables unsafe legacy import rewrites until the Slice 4b mutation coordinator exists.
- Sharded process and integration coverage requirements for Slices 5b and 6.

**Required changes:**
- State that the current plan is approved only for prerequisite-producing decomposition, or keep it blocked until the AC6, AC7, and AC14-AC17 artifacts exist and pass. Any task plan must limit the next wave to artifact, generator, validator, harness, and public-pack prerequisite beads.
- Add and name the executable AC17 gate that validates the acceptance-proof matrix and blocks behavior-changing bead creation when evidence is absent.
- Amend Slice 2's start gate to require an immutable Slice 1b compatibility pin in addition to AC15/AC16 cache schema proof, and state that unsafe legacy import rewrites are disabled at Slice 2 because the mutation coordinator is not available until Slice 4b.
- Add the external-prerequisite ledger with immutable commit and artifact references, validation commands, owners, consuming slices, and blocked-behavior policy for missing rows.
- Define a cross-repo bootstrap/staging workflow for validating no-Maintenance loading before final public-pack activation commits are immutable.
- Split or materially tighten Slices 2, 3, 5b, 6, and 7 so each deployable unit has exact files, gates, prerequisites, and rollback or one-way-boundary predicates.
- Expand the rollback/version-skew matrix and add downgrade tests for doctor-mutated manifests, runtime-state migration, lock/cache schema transitions, offline cache behavior, stale alias handling, and post-activation downgrade.
- Require `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` for Slices 5b and 6 as well as the loader, doctor, and runtime-state slices.
- Add explicit slice-level phase tags such as "decomposable now" and "blocked until gate" so automated or human decomposition cannot confuse prerequisite readiness with full implementation approval.
