# Marcus Driscoll

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash (current dependency output is the `_gemini.md` artifact; an older `_deepseek.md` artifact remains in the directory but was not used)

**Consensus findings:**
- [Major] Public Gastown cache identity still needs a canonical migration contract. Reviewers converge on source normalization, ordinary remote cache keys, old synthetic aliases, existing `packs.lock` entries, and online/offline upgrade behavior as mandatory design surface. Without a source-normalization and cache-migration matrix, stale aliases or lookalike sources can collide with ordinary cache keys or fail unpredictably.
- [Major] Legacy five-pack `SyntheticContentHash` transition needs an explicit stale-cache test. Removing Maintenance/Gastown from the bundled set changes the combined hash used by synthetic cache validation. The design should pin the new hash domain to `{core, bd, dolt}` and prove an old five-pack synthetic cache is treated stale and rematerialized offline without changing `bd`/`dolt` bytes or provenance.
- [Major] Required-pack materialization needs a concrete isolation and concurrency model. Claude asks whether `MaterializeRequiredPacks` stages per pack or mutates the shared tree in place; DeepSeek flags parallel startup repair/quarantine write collisions. The design must make mutation serialization, staging, prune/quarantine scope, and provider-pack isolation testable.
- [Major] Legacy synthetic-cache-shaped directories can reach ordinary remote/git validation after aliases retire. DeepSeek flags `.git`-less legacy directories as a crash/error surface, and Codex requires cache key acceptance/rejection cases. The loader must detect legacy synthetic cache markers before running ordinary git validation and either ignore, prune, promote after digest validation, or emit a typed diagnostic.
- [Major] Provider continuity gates are correct but scattered. The design needs generated slice gates tying `bd`/`dolt` untouched-file digest proof, Core repair isolation, provider exception rows, and role-cleaning witnesses to exact commands so implementation cannot satisfy a later slice with only smoke tests.
- [Major] `dolt` provider continuity needs named witnesses for moved scripts and notifications. Claude identifies `examples/dolt/port_resolve_test.go`, the Core move of `dolt-target.sh`, `port_resolve.sh` sourcing, `examples/dolt/pack.toml`, distinct escalation vs summary/DOG_DONE recipients, `--from controller`, and nudge failure handling as continuity-critical details.
- [Minor] Old Core path and retired alias rejection should be hard negatives, not soft diagnostics. `internal/bootstrap/packs/core`, retired Maintenance/Gastown aliases, and old public synthetic aliases need `IsSource`, `NameForSource`, lock/install, and `publicSubpathForPack` tests proving they cannot be silently selected.

**Disagreements:**
- There is no current verdict disagreement: Claude, Codex, and DeepSeek V4 Flash all return `approve-with-risks`. DeepSeek labels legacy combined-hash stale serving as a blocker-level risk inside an overall approve-with-risks verdict. Assessment: keep the persona verdict at `approve-with-risks`, but require the combined-hash stale-cache test before the registry/cache cleanup slice lands.
- Reviewers differ on legacy synthetic-cache promotion. Codex wants the promotion mention removed or made an explicit operator/helper path that never runs during normal init/load/doctor; DeepSeek wants automatic promotion for valid offline upgrades; Claude allows promotion, dual-read, or typed diagnostics. Assessment: the design must choose one online/offline behavior rather than leave promotion optional.
- The process-serialization mechanism is not settled. DeepSeek proposes a short advisory lock; Claude focuses on per-pack staging and repair isolation. Assessment: the design needs crash-safe mutation serialization without introducing stale runtime-state artifacts.
- `dolt` role-cleaning and notification details are emphasized mainly by Claude. Assessment: they still belong in the provider continuity gate because they protect store-critical behavior and cross-pack script contracts.

**Missing evidence:**
- A public Gastown source-normalization matrix with accepted and rejected source spellings, canonical cache key inputs, lock/install behavior, and lookalike-source cases.
- Exact migration behavior for existing `packs.lock` entries and legacy synthetic cache directories in online, offline, no-cache, valid legacy-cache, corrupted legacy-cache, and `.git`-less legacy-directory states.
- A generated `slice-gates.generated.yaml` row set that binds provider byte-continuity, Core repair isolation, and role-cleaning provenance to concrete commands.
- A decision on whether legacy synthetic-cache promotion is removed, explicit/manual, or automatic under offline/network-failure conditions, including helper ownership and digest-validation inputs.
- A test proving the old five-pack combined-hash synthetic cache is treated stale and rematerialized with the new `{core, bd, dolt}` domain.
- The `MaterializeRequiredPacks` staging/locking/quarantine contract and tests proving Core repair leaves `bd`/`dolt` bytes, modes, locks, installed metadata, cache digests, and provenance unchanged.
- `.git` preflight or synthetic-manifest detection behavior before ordinary remote git validation.
- Cross-pack script-continuity evidence for `dolt-target.sh`, `port_resolve.sh`, `examples/dolt/port_resolve_test.go`, and `examples/dolt/pack.toml`.
- Explicit `dolt` exception rows for `pool = "dog"`, distinct escalation and summary recipients, unconditional `DOG_DONE`, `--from controller`, delivery failure handling, and health JSON consumers.
- Hard negative fixtures for `internal/bootstrap/packs/core`, retired Maintenance/Gastown aliases, public Gastown synthetic aliases, and `publicSubpathForPack`.

**Required changes:**
- Add a canonical public-source normalization and cache-key test matrix covering `RepoCacheKey`, `IsSource`, `NameForSource`, install-lock generation, and materialization.
- Define the public Gastown cache migration contract: old-key lookup, new-key lookup, legacy cache promotion or rejection, offline no-cache behavior, corrupted legacy-cache behavior, `.git`-less legacy directory handling, typed diagnostics, and stale old five-pack cache behavior.
- Pin the `SyntheticContentHash` role division: combined hash is cache identity only; per-pack file-set manifest is the tamper authority; the new domain is `{core, bd, dolt}`. Add an offline stale-cache rematerialization test.
- Specify `MaterializeRequiredPacks` mutation semantics. If it mutates on normal runtime paths, add crash-safe serialization plus failure-injection tests; in all cases, add Core-repair isolation fixtures proving provider packs are unchanged except for explicit provider rewrite ledger entries.
- Copy the `bd`/`dolt` untouched-file digest proof, provider matrix, Core repair isolation witnesses, and role-cleaning old/new witnesses into the final slice-gate artifact contract.
- Detect `.git`-less legacy synthetic cache directories before running ordinary git commands, then safely ignore, prune, promote after digest validation, or report a typed migration diagnostic.
- Name `examples/dolt/port_resolve_test.go` in the test-migration map, define the replacement content/sourcing assertion at the Core path, and update or remove the co-location fallback in `dolt-target.sh` for the new Core depth.
- Resolve the `dolt` notification/binding contract: keep escalation and summary/DOG_DONE as distinct configurable recipients, preserve `--from controller` authorship where required, stop masking rewritten nudge failures, and decide whether `pool = "dog"` is stable, configurable, or replaced by Core bindings.
- Convert the old Core source path and retired aliases into hard negative tests for `IsSource`, `NameForSource`, lock/install, and `publicSubpathForPack`.
