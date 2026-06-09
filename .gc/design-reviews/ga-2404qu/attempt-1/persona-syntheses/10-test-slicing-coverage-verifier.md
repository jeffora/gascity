# Tomas Park

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash (present in the `_gemini.md` artifact; no separate `_deepseek.md` artifact was present)

**Consensus findings:**
- [Blocker] Core role-asset deletion and de-referencing are not assigned to a named green slice. Claude and DeepSeek both identify that moving `mol-polecat-*` out of Core while Core formulas such as `mol-do-work` still reference those formulas can break composition in an intermediate commit. Codex does not call out that asset set directly, but its request for exact generated slice rows and forbidden intermediate states depends on the same missing executable slice contract.
- [Blocker] The rollout authorities are not reconciled into one deterministic gate plan. Claude finds three decompositions plus a future `slice-gates.generated.yaml` artifact with no crosswalk; DeepSeek flags Slice 4 prose omitting process and integration shards required by the matrix; Codex says generated gate artifacts still need concrete command/test entrypoints and sample rows before implementation can execute safely.
- [Blocker] The stale synthetic-cache rule is contradictory. DeepSeek reports that prose Slice 2 requires stale retired-alias cache rejection while the slice-gate matrix says stale synthetic cache is ignored. Those behaviors are mutually exclusive and would produce divergent tests and implementation.
- [Major] The Maintenance-to-Core fold can create duplicate active definitions. Claude and DeepSeek both find that adding same-named `dog` assets to Core while the Maintenance pack remains active violates the stated duplicate-free loader contract unless the fold is a per-asset atomic move or an activation/removal atomic commit.
- [Major] First-pass evidence generation is floating. The design depends on `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, `loader-inventory.generated.yaml`, `slice-gates.generated.yaml`, `test-migration.generated.yaml`, `public-gastown-pins.yaml`, and related artifacts before destructive moves, but does not assign their generation, schemas, freshness tests, and row IDs to a concrete initial slice.
- [Major] Behavior-manifest completeness is self-certified. Claude and DeepSeek both note that generated-vs-checked-in freshness cannot detect discovery false-negatives if both files come from the same walk. Codex's request for concrete generator/validator entrypoints reinforces that the evidence path must be mechanically testable, not only described.
- [Minor] The `examples/` transition window is asserted rather than demonstrated. Claude and DeepSeek both question how `go test ./examples/...` remains green while example cities are rewired in slices 2-4 but stale local packs and local tests are removed later.
- [Minor] External `gascity-packs` gates are not executable enough. Codex accepts the intended coverage but asks for exact pack-local commands and artifacts for formula/order composition, prompt-template resolution, script execution, retired-path scans, and old-test mapping.

**Disagreements:**
- Verdict split: Claude and DeepSeek block; Codex gives approve-with-risks. Assessment: this persona blocks because the unresolved ordering can produce broken intermediate commits, which is directly in the test-slicing lane.
- Codex treats placeholder generated-artifact commands as minor execution risk, while Claude and DeepSeek treat missing crosswalks and slice ownership as blockers. Assessment: exact commands and row IDs are required before implementation beads can claim green slices.
- The `mol-polecat-*` / `mol-do-work` gap is foregrounded by Claude and DeepSeek but not Codex. Assessment: accept it as a blocker because it is a concrete behavior-composition failure mode.
- DeepSeek alone identifies the stale-cache rejection-versus-ignore contradiction. Assessment: keep it blocking until the design proves the two terms are not describing the same fixture, because mutually exclusive cache semantics would invalidate slice gates.

**Missing evidence:**
- A named slice for deleting or rewriting in-tree Core role assets, including `mol-polecat-*`, `gc-dispatch`, `mol-do-work`, and any sibling formula or skill that references Gastown-bound assets.
- Expiry slices for temporary role-token allowlist rows that cover role tokens currently present in Core.
- A canonical crosswalk mapping prose rollout slices 1-7, the slice-gate matrix, activation steps A-F, and `slice-gates.generated.yaml` rows, with one representation named authoritative.
- Exact generated gate rows showing commands, package targets, `-run` filters, artifact inputs, public-pin phase, old/new binary fixtures, cache/offline fixtures, forbidden states, and required sharded targets.
- Concrete generator and validator entrypoints for `slice-gates.generated.yaml`, `behavior-manifest.generated.yaml`, `test-migration.generated.yaml`, `public-gastown-pins.yaml`, and scanner artifacts.
- A resolved stale synthetic-cache behavior: reject or ignore, consistently across prose, generated gates, tests, and implementation.
- A per-commit duplicate-free assertion for the Maintenance fold candidate tree through the production loader.
- An independent VCS-level completeness check that reconciles manifest rows against moved, deleted, and added pack assets so generator discovery false-negatives fail CI.
- Exact `gascity-packs` commands for the candidate public Gastown slice.
- A demonstrated mechanism that keeps `go test ./examples/...` green during slices 2-4.

**Required changes:**
- Assign the in-tree Core role-asset removal to a named slice. Re-home `mol-polecat-*` in public Gastown before or in the same commit that removes them from Core, de-reference Core formulas in that same commit, and state the expiry slice for each temporary role-token allowlist row.
- Add a canonical slice/commit crosswalk covering prose slices 1-7, matrix rows, activation steps A-F, and `slice-gates.generated.yaml`. State which source is authoritative and require implementation beads to cite generated row IDs.
- Align Slice 4 prose gates with the matrix by explicitly requiring `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` where controller reload, doctor repair, and config-loading behavior can regress.
- Reconcile stale synthetic-cache semantics everywhere. Pick rejection or ignored behavior and update prose, generated gates, tests, and implementation expectations to match.
- Define the Maintenance fold as duplicate-free at every mergeable state. Use per-asset atomic moves or one atomic activation/Maintenance-removal commit, and add a production-loader test proving zero duplicate active definitions on the candidate tree.
- Add an explicit Slice 0 or equivalent first Gas City slice for evidence generation, schema validation, freshness tests, pilot rows, and row-id availability before any destructive source move, public pin update, role rewrite, or test removal.
- Add an independent VCS-level behavior-manifest completeness gate so discovery false-negatives fail CI instead of passing generated-vs-checked-in freshness.
- Fill in concrete generator/validator and pack-local gate commands, including the external `gascity-packs` formula/order composition, prompt-template resolution, script execution, retired-path scan, and old-test mapping checks.
- Specify how the example tree remains green during slices 2-4, either by removing or rewiring stale local packs and tests in the same slice as the city rewire or by documenting a concrete skip/delete strategy.
