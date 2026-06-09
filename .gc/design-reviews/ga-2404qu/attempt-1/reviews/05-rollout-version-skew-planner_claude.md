# Yuki Hayashi - Claude

**Verdict:** approve-with-risks

I reviewed only the rollout lane: two-repo sequencing, public-pin integrity,
intermediate-state safety, and rollback granularity. The rollout *shape* is
sound and approvable — it is staged, sha-pinned, non-destructive, and mandates
no flag-day on the happy path. The risks below are real and concentrate at the
single irreversible step (activation / Maintenance removal); they must be closed
before that slice merges, but they do not block starting slice 1.

**Top strengths:**
- **Pin integrity is genuinely immutable and content-verified, not ref-based.**
  Verified in-tree: `internal/config/public_packs.go` pins
  `PublicGastownPackVersion = "sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b"`
  against source `https://github.com/gastownhall/gascity-packs.git//gastown`
  (a real remote, not a synthetic alias). The design's `current_baseline`
  ledger row (design.md L2214) records that exact shipped sha, so the ledger is
  grounded in reality, and the pin validator must "observe file content digests
  from the actual checkout or cache path used for pack installation" with
  registry/direct-source equality proof (L2242–2244, L2520–2529). This closes
  "points at a mutable/untested commit."
- **The staged 7-slice rollout avoids a flag-day on the happy path and isolates
  the irreversible step.** Pin adoption (slice 2, compatibility) is separated
  from activation+Maintenance removal (slice 5, activation), with green
  intermediate commits A–F (L2263–2271) and per-slice gates. Registry/synthetic
  cleanup (slice 6) is explicitly *after* and must not be the first pin update
  (L2232–2233, L3442–3448).
- **Intermediate state and rollback are contract-bound, not best-effort.**
  Runtime-state migration is non-destructive (legacy Maintenance state retained
  for old binaries; migration-completed marker prevents false conflicts,
  L2564–2578), and the duplicate-active matrix is required across skew combos
  including "old v1.2.1 binary + activation pin," "new binary + old pack," and
  "rollback after pin adoption" (L1273–1279, L2229–2241).

**Critical risks:**

- **[Major] Staged safety depends on a gascity-packs compatibility commit that
  can omit colliding active assets — a cross-repo precondition asserted, not
  proven.** The compatibility pin must ship public Gastown that coexists with
  *still-bundled* Maintenance (Maintenance isn't removed until slice 5), so it
  "must omit colliding formulas, orders, prompt fragments, scripts, agents,
  patch targets, hooks, and runtime-state owners from active discovery"
  (L1850–1856). If such a commit cannot be authored, the design's own fallback
  is "skip the compatibility pin and land a paired cross-repo activation/removal
  boundary" (L1599–1600, L1870–1871) — i.e. a coordinated cross-repo cutover,
  which is the flag-day this rollout is built to avoid. The likelihood of the
  fallback is never assessed, and Gas City's design can only *gate* on the
  property, not guarantee it.

- **[Major] Activation reversibility is left undecided ("supported OR documented
  one-way"), which is exactly the rollback-granularity question.** The design
  repeatedly defers: "`v1.2.1` plus activation pin | either supported … or named
  as a one-way boundary before activation" (L2233); "Doctor-mutated manifests
  must remain readable by `v1.2.1` unless the release notes name the exact
  one-way boundary and manual recovery path" (L1498–1500, L3471). After
  activation, an old binary re-bundles Maintenance into `requiredBuiltinPackNames`
  while the migrated city now carries an explicit public Gastown import — a
  duplicate-active / no-clean-downgrade hazard. The ledger forces the old-binary
  result to be *recorded* before merge (L2526–2528), but the design itself does
  not commit to whether activation is reversible or document the manual recovery
  if it is not. For this lane that decision cannot stay open into the activation
  slice.

- **[Major] The availability model changes for existing cities, not just fresh
  init.** Gastown moves from embedded-in-binary to a fetched public pack with
  "no embedded fallback" and "offline no-cache = explicit failure"
  (L2239, L2954–2958). The design states the offline/air-gap dependency only for
  `gc init --template gastown`. But every *existing* new-binary Gastown city now
  also depends on the pinned commit being reachable (network or pre-populated
  ordinary remote cache) whenever the pack is (re)installed — fresh host, cache
  miss, lock refresh, or `doctor --fix` repair — since Gastown is no longer
  embedded (a steady-state load of an already-materialized pack reads from disk).
  There is no operator-facing
  statement that air-gapped or previously-embedded deployments must pre-populate
  the ordinary cache to keep running Gastown after upgrade, nor what an existing
  city does if its pinned commit becomes unreachable mid-life (the matrix only
  covers "new binary + old pack" as a version-skew doctor case, L3469).

- **[Minor] Pin durability over time is not guaranteed by a ref policy.** The
  baseline is a bare commit on gascity-packs `main` (per the in-tree comment,
  public_packs.go L9–10). The ledger has a "durable public ref" field
  (L2220, L2526), but the design never requires gascity-packs to publish a
  durable tag/release per pinned commit. A history rewrite, force-push, or GC on
  `main` could make a pinned `sha:` unfetchable for every cache-miss city. Immutable
  *content* ≠ durable *reachability*.

**Missing evidence:**
- The compatibility and activation commit SHAs (TBD until slice 1) and proof
  that a safe compatibility commit (colliding assets omitted) can actually be
  authored — the staged path's load-bearing assumption.
- The recorded "old `v1.2.1` binary + activation pin" duplicate-active result
  that decides whether activation is reversible.
- A stated behavior for `gc init --template gastown` executed by a slice-2
  binary in the window before/after pin update (it should produce the prior
  `current_baseline` pin and be deployable, but the doc never says so).

**Required changes:**
- Decide and document activation reversibility: either guarantee old-binary
  readability of activated cities, or name the exact one-way boundary and the
  operator downgrade/recovery procedure, in the design and the release notes,
  before the activation slice.
- Require gascity-packs to publish a durable tag/release ref for every consumed
  pin and make the pin ledger validator verify reachability through that durable
  ref (not just a commit on `main`).
- Add an Operational-risk statement covering the embedded→fetched availability
  shift for *existing* Gastown cities and air-gapped deployments (pre-populate
  the ordinary remote cache; no embedded fallback), plus the failure mode when a
  pinned commit becomes unreachable for a running city.
- State explicitly that the compatibility-pin precondition (omittable colliding
  assets) is verified in slice 1 and that its failure converts the rollout to
  the documented paired cross-repo boundary, with an estimate of how likely that
  fallback is.

**Questions:**
- Is activation (slice 5) the only mandatory paired cross-repo step, and is its
  rollback story the documented-manual path or an auto-safe path?
- What durable-ref policy will gascity-packs adopt so pinned commits remain
  fetchable for the supported lifetime of binaries that consume them?
- For an existing new-binary Gastown city whose pinned commit becomes
  unreachable (and not in cache), does config load fail closed, and is that the
  intended availability posture versus the old embedded behavior?
