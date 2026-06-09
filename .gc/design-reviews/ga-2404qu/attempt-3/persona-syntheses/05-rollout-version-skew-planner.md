# Yuki Hayashi

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Strength] The broad rollout strategy is directionally sound: land `gascity-packs` first, keep `PublicGastownPackVersion` as an immutable `sha:` pin, add packcompat gates before deleting in-tree sources, stage the work into slices, and treat the offline-to-network change as an explicit product decision.
- [Blocker] The detailed rollout order violates the replacement-before-removal invariant. The design can fold/remove Maintenance or Gastown-owned behavior before `PublicGastownPackVersion` is updated to the replacement commit, so fresh Gastown init may still pin the old public pack while Gas City behavior has already moved or disappeared.
- [Blocker] The pin/alias cutover is not safely sequenced. Slice 6 combines removing Gastown/Maintenance from the built-in registry, retiring public synthetic aliases, bumping `PublicGastownPackVersion`, and validating stale cache rejection. The new pin, new Core, ordinary network fetch path, and no-synthetic-alias path are not proven together before the release commits to them.
- [Blocker] Synthetic cache validation has a first-run upgrade failure mode. Removing Gastown and Maintenance from `All()` changes both `SyntheticContentHash()` and the allowed file set used by `ValidateSyntheticRepo`; existing bundled synthetic caches can fail with unexpected-file or hash errors unless the design chooses a namespace bump, cache cleanup/migration, or deliberate legacy validation window.
- [Blocker] `requiredBuiltinPackNames` still includes `maintenance`, and no rollout slice explicitly changes it. If Maintenance files move before this required-pack list changes, materialization can still try to include a pack that no longer exists; if it changes atomically with registry removal, it inherits the slice-6 consistency risk.
- [Major] Public pin reachability is underspecified. A `sha:` pin is immutable but only deployable if the object remains fetchable from a retained public ref; CI must prove the exact commit is reachable through the ordinary remote-pack install path, materializes to the pinned HEAD, and does not come from a bundled synthetic marker.
- [Major] The packcompat gate is not anchored to the exact new-pin/new-Core pair early enough. The planned `TestPinnedPublicGastownBehavior` would use the current `PublicGastownPackVersion` until the constant changes, so slices before the cutover can validate the old pin rather than the candidate replacement commit.
- [Major] Old-binary/new-public-pack compatibility is asserted but not gated. The matrix says the candidate public pack must work with older Gas City binaries, but no CI job installs the candidate pack with the last released `gc` binary and proves fresh init, existing locked cities, formulas/orders, and no-Maintenance operation.
- [Major] Rollback after doctor/import-state mutation is not operationally defined. Once `gc doctor --fix` removes Maintenance imports or rewrites Gastown imports to a new public commit, downgrading can leave cities unresolvable offline or silently tied to stale embedded content unless the migration is declared forward-only or has tested recovery steps.
- [Major] The code-path transition matrix is missing. Phase-sensitive paths include `All()`, `requiredBuiltinPackNames`, `publicSubpathForPack`, `IsSource`/`NameForSource`, `SyntheticContentHash`, `ValidateSyntheticRepo`, `MaterializeBuiltinPacks`, `pruneStaleGeneratedPackFiles`, `RepoCacheKey`, `legacyPublicPackForSource`, `defaultWave1PublicPackImports`, `GastownCity`, and `builtinPackIncludes`.
- [Major] The post-migration public Gastown import graph is unspecified. The design must say what replaces `../maintenance`: explicit Core import, host auto-inclusion plus patching, duplicated/moved assets, or another dependency mechanism.
- [Major] `examples/gastown/` has no clear target state. It must either become a public-import-only example with updated tests/docs or be removed with tests such as registry identity assertions updated in the same slice.

**Disagreements:**
- Claude rated this lane `approve-with-risks`; Codex and DeepSeek V4 Flash rated it `block`. Assessment: choose `block` because the unresolved pin/alias ordering, cache validation, required-pack transition, and rollback story can leave fresh or upgraded cities unresolvable.
- Claude recommends splitting slice 6 into pin bump then alias retirement; Codex allows either a proven atomic cross-repo release slice or explicit ordering; DeepSeek emphasizes that slice 6 already combines too many state changes. Assessment: require a tested sequence, with a split cutover preferred because it proves production fetchability before removing the fallback.
- Reviewers differ on offline support after alias removal. Assessment: offline public Gastown init is not required by this lane if the product decision is network-required init, but the failure path must be tested and actionable.
- DeepSeek focuses more heavily on `ValidateSyntheticRepo`, `SyntheticCacheNamespace`, `RepoCacheKey`, and `requiredBuiltinPackNames` than Claude/Codex. Assessment: accept those as required changes because they are concrete upgrade breakpoints in the rollout path.

**Missing evidence:**
- A phase-by-phase matrix for fresh init, existing cities, manual imports, doctor fix, rollback, and air-gapped/network-failure behavior.
- A pin-promotion procedure with concrete commands/artifacts: public repo and subpath, candidate commit SHA, retained ref/tag, behavior inventory artifact, `packs.lock` result, ordinary remote cache path, and Gas City PR changing `PublicGastownPackVersion`.
- CI proof that the candidate pin is `sha:<40hex>`, fetchable from the public repository, materializes to the exact commit, and cannot be satisfied by `.gc-bundled-pack-cache.toml`.
- A gate that runs the candidate public pack with the last released `gc` binary and proves old-binary/new-pack compatibility.
- A tested order for public pin bump versus synthetic alias removal, including unreachable remote/commit behavior and stale synthetic cache behavior.
- A strategy for synthetic cache transition: namespace bump, explicit stale-cache cleanup/migration, or legacy validation tolerance.
- A specific slice assignment for changing `requiredBuiltinPackNames`, `publicSubpathForPack("maintenance")`, `All()`, registry identity tests, and synthetic alias behavior.
- A rollback or forward-only migration contract for cities whose `city.toml` and `packs.lock` were mutated by `gc doctor --fix`.
- A specification for the public Gastown pack's post-Maintenance import graph and moved asset ownership.
- A target-state decision for `examples/gastown/`.

**Required changes:**
- Reorder or pair the public-pack release so `PublicGastownPackVersion` is updated and packcompat validates the exact candidate commit before any Gas City slice removes, generalizes, or stops including old Gastown/Maintenance behavior. If no deployable midpoint exists, mark it as one paired cross-repo release slice with explicit gates.
- Split the pin/alias cutover or otherwise prove it atomically. Preferred sequence: publish and retain the public commit, run packcompat against the exact new pin and extracted Core, bump `PublicGastownPackVersion`, verify production fetchability, then retire public synthetic aliases in a later separately deployable step.
- Add an immutable-pin release gate: reject non-`sha:<40hex>` pins, fetch the exact object from `gascity-packs`, assert the materialized cache HEAD equals the pin, and assert no bundled synthetic cache marker is used for public Gastown.
- Add an old-binary/new-pack compatibility gate using the last released `gc` binary against the candidate public pack, covering fresh init, existing locks, formula/order composition, and no-Maintenance operation.
- Add a code-path transition matrix covering `All()`, `requiredBuiltinPackNames`, `publicSubpathForPack`, `IsSource`/`NameForSource`, `SyntheticContentHash`, `ValidateSyntheticRepo`, `MaterializeBuiltinPacks`, `RepoCacheKey`, `legacyPublicPackForSource`, `defaultWave1PublicPackImports`, `GastownCity`, and `builtinPackIncludes`.
- Resolve synthetic cache migration explicitly. Account for both content-hash rotation and allowed-path shrinkage; choose namespace bump, cache cleanup/migration, or legacy validation tolerance.
- Explicitly assign the `requiredBuiltinPackNames` change to a rollout slice and add tests for the transition where Maintenance is no longer in `All()` while stale `.gc/system/packs/maintenance/` directories still exist.
- Update `publicSubpathForPack` in step with registry removal and test that `publicSubpathForPack("maintenance")` returns `("", false)` and Maintenance public sources are not treated as synthetic after retirement.
- Specify rollback behavior after `gc doctor --fix`. Either provide tested downgrade recovery steps or declare the migration forward-only with clear diagnostics and release notes.
- Specify the public Gastown import graph after `../maintenance` is retired and test that it does not rely on an unresolved Maintenance import.
- Decide whether `examples/gastown/` survives; update its configs, registry tests, and docs in the same slice as that decision.
