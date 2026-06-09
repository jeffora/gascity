# Lena Hoffmann - Claude

**Verdict:** block

I reviewed `plans/core-gastown-pack-migration/implementation-plan.md` strictly in
my lane — public Gastown pin integrity, immutable content hash, `RepoCacheKey`
identity, synthetic-alias retirement, and offline/rollback behavior — and
grounded the claims against shipped code in
`internal/builtinpacks/registry.go`, `internal/config/pack_include.go`,
`internal/config/public_packs.go`, and `internal/packman/cache.go`. The plan
states the correct end state, but its slice ordering consumes the public pin
several slices before it retires the mechanism that currently makes that pin
resolve from embedded binary bytes. That is a pin-integrity defect in my exact
lane, so I block.

The artifact otherwise conforms to `implementation-plan.schema.md`: front matter
has `phase: implementation-plan`, a `requirements_file`, `status: draft`, no
`design_file`; the seven required sections are present and in order. (Schema nit
below: the plan is self-described as prerequisite-blocked, which reads closer to
`blocked` than `draft`.)

**Top strengths:**

- **The pin is a real immutable SHA and mutable-ref drift is explicitly
  forbidden.** `internal/config/public_packs.go` ships
  `PublicGastownPackVersion = "sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b"`
  with `PublicGastownPackSource = "…/gascity-packs.git//gastown"` (subpath
  included), and the plan states "Durable public refs may keep SHAs fetchable,
  but they never replace immutable SHA plus digest validation." Red flag 1
  (mutable branch/tag pinning) is well defended.
- **The offline / no-silent-fallback intent is explicit and correct.** The
  testing matrix includes "fail-closed cache miss when network access is
  unavailable," activation-pin packcompat "fails if any assertion resolves from
  `examples/gastown`, `.gc/system/packs/gastown`, `.gc/system/packs/maintenance`,
  or a synthetic public alias," and recovery "must not silently reselect retired
  embedded behavior." This is the right target for red flag 3.
- **Two-pin model + per-slice rollback + release compatibility matrix** give a
  concrete deployable/rollback narrative across the gascity-packs-landing →
  activation-pin window (lane question 3): compatibility pin coexists with the
  current loader, activation pin is candidate-branch only, and each slice names
  a rollback.

**Critical risks:**

- **[Blocker] The public pin is consumed (Slice 2) four-to-five slices before
  the synthetic alias that currently makes it resolve from embedded binary bytes
  is retired (Slice 6).** Shipped code:
  `NameForSource("…/gascity-packs.git//gastown")` matches a synthetic layout
  because `publicSubpathForPack("gastown") → true`, so
  `builtinpacks.IsSource(PublicGastownPackSource) == true`. With `IsSource` true,
  `packman.EnsureRepoInCache` takes `ensureBundledRepoInCacheLocked`, which calls
  `MaterializeSyntheticRepo` — "writes the running binary's bundled pack tree" —
  and `ValidateSyntheticRepo` checks `marker.ContentHash == SyntheticContentHash()`
  (the current binary's content), using the pinned commit only as a cache tag.
  So consuming `PublicGastownPackVersion` for the public Gastown source today
  resolves the **embedded** gastown bytes, validated against binary content, not
  the immutable commit's real tree. The plan's contract says "the subpath-aware
  lock/cache proof lands before the first Gas City slice that updates
  `internal/config/PublicGastownPackVersion`" and lists "stale-alias rejection"
  and "ordinary remote-pack install for the exact pin" as **Slice 2** gates — but
  it schedules "retire public synthetic aliases" at **Slice 6**. For slices 2–5
  the consumed pin is satisfied by `MaterializeSyntheticRepo`/`SyntheticContentHash`,
  i.e. embedded retired behavior, which is exactly lane question 2 ("can embedded
  bytes select retired Gastown content after the public pin lands?") and red flag
  3. The Slice-2 contract and the Slice-6 schedule contradict each other, and the
  code confirms the Slice-2 gates cannot pass while `publicSubpathForPack`/
  `NameForSource`/`IsSource` still treat the public source as bundled-synthetic.

- **[Major] `RepoCacheKey` identity is subpath-blind in shipped code, and the
  plan asserts the opposite without naming the change or its cache-invalidation
  blast radius.** `config.RepoCacheKey(source, commit)` builds its key from
  `NormalizeRemoteSource(source) + commit`, and `NormalizeRemoteSource` returns
  `remotesource.Parse(source).CloneURL`, whose doc comment says it is "stripping
  subpath and ref suffixes." So `…/gascity-packs.git//gastown` and
  `…/gascity-packs.git//maintenance` at the same commit hash to the **same**
  cache key today (true in both the ordinary and the `SyntheticCacheNamespace`
  branch). The plan says "`RepoCacheKey` includes normalized source, exact
  commit, and subpath," but `RepoCacheKey(source, commit)` has no subpath
  parameter and the normalizer discards it. Making the key subpath-aware (new
  parameter, or stop stripping subpath in `NormalizeRemoteSource`) changes the
  key for **every** remote pack and invalidates existing `~/.gc/cache/repos`
  entries; `packman.RepoCacheKey` delegates to `config.RepoCacheKey`, so both
  move together. `Data And State` names lock fields (source, commit, subpath,
  pack `sha256`, manifest `sha256`) but never names this cache-key migration or
  its one-time invalidation. RepoCacheKey identity is the center of my lane and
  the plan leaves the mechanism and blast radius unspecified.

- **[Minor] The one-command pin-coherence gate omits the registry source
  identity and the packcompat transcript.** The gate "compares
  `PublicGastownPackVersion`, `public-gastown-pins.yaml`, fresh-init output,
  lockfile provenance, cache proof, pack digest, and behavior-manifest digest in
  one command." It does not bind `internal/builtinpacks` source normalization
  (`PublicRepository` / `NormalizeRemoteSource`) or the packcompat transcript's
  resolved commit+subpath+digest into the same check, so packcompat and the
  consumed pin are proven separately rather than proven identical. Lane question
  1 wants all of {`PublicGastownPackVersion`, pins.yaml, registry source,
  packcompat, direct cache proof} bound to the same commit and subpath; fresh-init
  covers the registry normalization indirectly, but packcompat is left outside
  the coherence command.

- **[Minor] Legacy synthetic → ordinary cache promotion is tested but its helper
  and digest-gate are not specified.** `Testing` lists "ordinary remote cache
  promotion" and `Proposed Implementation` states "Promotion and read hits verify
  source, commit, subpath, pack digest, and manifest digest," but no section
  names the promotion helper, when it runs, or that promotion of a stale
  synthetic Gastown cache into the ordinary key is allowed only after a full
  source+commit+subpath+pack-digest+manifest-digest match. Without that, red flag
  2 (promotion laundering stale bytes) is asserted-away rather than mechanized.

**Missing evidence:**

- Which slice makes `IsSource`/`NameForSource`/`publicSubpathForPack` stop
  recognizing the public Gastown source as bundled-synthetic, and proof that no
  pin consumption (lock generation, lock refresh, install, offline upgrade)
  occurs before that change.
- The concrete code change that makes `RepoCacheKey` subpath-aware, and the
  one-time repo-cache invalidation/migration it forces for all remote packs.
- A proof that, during slices 2–5, the consumed compatibility pin resolves from
  the real immutable commit (network or validated ordinary cache) and never from
  `MaterializeSyntheticRepo`/`SyntheticContentHash`.
- How the Core `core.maintenance_worker` binding contract (implemented in Slice
  4a) is frozen before the activation public pack — which patches
  `target_binding = "core.maintenance_worker"` — is cut immutable in Slice 1c.

**Required changes:**

- Move synthetic-alias retirement for `gastown` (and `maintenance`) — the
  `publicSubpathForPack`/`NameForSource`/`IsSource` change plus the subpath-aware
  `RepoCacheKey` and ordinary-remote resolution — into Slice 2 (or a Slice 0
  before any pin consumption), not Slice 6. The Slice-2 "ordinary remote-pack
  install for the exact pin" and "stale synthetic-cache rejection" gates require
  this code to exist when the first pin is consumed.
- Specify the `RepoCacheKey` subpath change explicitly (signature or normalizer),
  state that `config` and `packman` move together, and add the repo-cache
  invalidation/migration to `Data And State`.
- Add a test asserting that for `PublicGastownPackSource` at
  `PublicGastownPackVersion`, resolution never takes the bundled-synthetic path
  (`IsSource` must be false) and is validated against the commit's real pack
  digest, not `SyntheticContentHash()`.
- Fold the registry source identity and the packcompat resolved-commit/subpath/
  digest into the single pin-coherence command so all five surfaces prove the
  same immutable commit and subpath.
- Name the synthetic→ordinary cache promotion helper and require a full
  source+commit+subpath+pack-digest+manifest-digest match before promotion writes.

**Questions:**

- Until `publicSubpathForPack` is changed, does any slice-2–5 path (lock refresh,
  `gc pack install/check`, offline upgrade) resolve the compatibility pin through
  `MaterializeSyntheticRepo`? If so, how is that not "the public pin resolving
  from embedded bytes"?
- Is `PublicGastownPackVersion` ever permitted to be a non-`sha:` (durable ref)
  value, or is `sha:` enforced at parse time? A test should pin this.
- How is the activation pin (immutable at Slice 1c) protected from a binding-
  contract change in Slice 4a that would force re-cutting the "immutable" commit
  and lengthening the gascity-packs → activation window?
- Should `status` be `blocked` rather than `draft`, given the Summary and Open
  Questions state external prerequisites block all dependent slices?
