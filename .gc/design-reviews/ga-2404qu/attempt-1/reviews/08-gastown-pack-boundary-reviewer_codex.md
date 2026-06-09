# Avery McAllister

**Verdict:** approve-with-risks

**Top strengths:**
- The target boundary is explicit: public Gastown depends on host Core by contract, never by importing Core or Maintenance, and Maintenance is retired rather than recreated as another system pack.
- The design requires generated behavior rows, a public Gastown asset ledger, exact pinned-checkout packcompat, and no-Maintenance production-loader mode before source deletion. That is the right evidence shape for preserving Polecat, branch pruning, detector/requester, review workflow, commands, doctor checks, overlays, prompts, warrants, and notifications.
- The rollout order is materially safer than a flag-day move: public Gastown lands first, Gas City adopts a compatibility pin, Core is extracted, then activation and Maintenance removal happen on the same candidate tree with old/new binary and rollback evidence.

**Critical risks:**
- [Major] The Pack Ownership section still leaves a test-coverage loophole. It says pack-specific tests from `examples/gastown/gastown_test.go` can move to `gascity-packs` "or" be replaced by Gas City tests that assert init/import wiring only (design.md:2777-2784). Later sections correctly require old-test to new-test mapping, behavior-preservation tests, and a failure if any `examples/gastown/gastown_test.go` or `maintenance_scripts_test.go` assertion disappears without a manifest row (design.md:2916-2919, design.md:3335-3351). The older "or wiring only" wording is still dangerous: it permits exactly the red-flag outcome where Gas City deletes local behavior tests before equivalent public-pack coverage exists. Make the later contract the only contract.
- [Major] Public Gastown routes to the host Core maintenance worker need an explicit replacement rule. The design correctly says active public Gastown must use symbolic `target_binding = "core.maintenance_worker"` for patches, and old `[[patches.agent]] name = "dog"` must disappear by activation (design.md:1788-1845, design.md:2169-2175). Current public-pack worktrees still contain `[[patches.agent]] name = "dog"`, comments that the host supplies an implicit maintenance/core layer, `pool = "dog"`, and prompt/formula warrant metadata like `gc.routed_to="{{binding_prefix}}dog"`. The design has generic witness rows for `pool = "dog"` and hardcoded worker targets (design.md:2672-2677), but it should say directly that public Gastown's active prompts, formulas, orders, and role instructions may not target `{{binding_prefix}}dog` unless `dog` is a public-Gastown-owned agent. If the target is host Core, it must resolve through the host-Core binding and be tested with a renamed Core maintenance worker.
- [Minor] The final ownership prose lists roles, formulas, orders, scripts, prompts, overlays, branch pruning, and Polecat, but commands and doctor checks are mostly covered only in later witness-row language (design.md:2657-2661). Since this lane asks whether public Gastown owns commands and doctor checks after the split, the Pack Ownership summary should explicitly include them to avoid implementers treating command/doctor migration as secondary cleanup.

**Missing evidence:**
- A concrete sample from `test-migration.generated.yaml` showing an old `examples/gastown/gastown_test.go` subtest mapped to a public `gascity-packs/gastown` test and then consumed by Gas City's packcompat gate.
- A public Gastown route matrix for deacon, witness, boot, digest, shutdown-dance, branch pruning, review checks, and dog/maintenance-worker interactions showing final owner, final route target, whether the target is public Gastown or host Core, and the renamed-worker fixture that proves it.
- A wording-lint fixture that fails on public Gastown comments like "implicit maintenance/core utility layer" and on active public-pack imports or docs that imply a standalone Maintenance layer.

**Required changes:**
- Replace the Pack Ownership test sentence with: public-pack behavior tests must move to or be recreated in `gascity-packs/gastown`; Gas City may retain only additional init/import wiring tests after the generated test-migration matrix proves equivalent public-pack coverage or approved removals.
- Add an explicit public-Gastown host-Core route rule: any active public Gastown patch, formula pool, `gc.routed_to`, warrant metadata, mail/nudge target, prompt instruction, or command that targets the Core maintenance worker must use the configured host-Core binding, not `{{binding_prefix}}dog` or `name = "dog"`. Packcompat must run this with default, renamed, and omitted maintenance-worker fixtures.
- Add commands and doctor checks to the final Pack Ownership summary, not just to witness rows, with required final paths or `blocked:path-unresolved` rows before source moves.

**Questions:**
- Should public Gastown retain any public-owned `dog` agent, or is every `dog` surface after activation strictly host-Core compatibility data? The route rules differ depending on that answer.
- Which public Gastown test package is the canonical home for moved Gas City behavior tests, and will Gas City's `test/packcompat` invoke it or only consume its generated manifest and pinned assets?
