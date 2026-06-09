# Lena Hoffmann - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan correctly separates public Gastown into compatibility and activation pins, and it blocks Gas City source deletion, Maintenance removal, and activation-pin consumption until public-pack evidence exists at immutable commits.
- The proposed pin artifacts carry the right identity fields: source URL, subpath, immutable commit, pack digest, behavior-manifest digest, generated timestamp, approving PR, and packcompat transcripts for both current-loader and host-Core/no-Maintenance modes.
- The cache design points in the right direction: public Gastown must resolve through ordinary remote-pack cache entries, `RepoCacheKey` becomes source plus commit plus subpath, stale synthetic aliases cannot satisfy a public `sha:` pin, and offline cache hit/miss/digest/subpath cases are named test gates.

**Critical risks:**
- [Major] The cache-key transition is underspecified. Current code and tests still treat `RepoCacheKey(source, commit)` as subpath-agnostic, while this plan requires subpath-aware identity. The plan says old records lacking subpath/digest fields become diagnostic-only, but it does not define the migration or air-gapped seeding path for an operator who already has the exact root repo cached under the old key.
- [Major] Synthetic public Gastown cutoff must happen before the first `PublicGastownPackVersion` adoption across every path, not just runtime resolution. The plan says Slice 2 rejects retired synthetic cache hits, but it should explicitly cover fresh init lock generation, `gc import install`, `gc import check`, lock refresh, `config.Resolve*`, packman cache promotion, doctor/import-state validation, and offline repair before the compatibility pin can be consumed.
- [Minor] There are two pin-ledger authorities in the prose: external `gascity-packs/gastown/public-gastown-pins.yaml` and local `support/public-gastown-pin-ledger.yaml`. That can work, but only if the pin-coherence gate treats one as external source evidence and the other as Gas City consumption evidence, then fails on drift or missing fields.

**Missing evidence:**
- The support directory currently contains only `maintenance-asset-classification.md`; the pin ledger, version-skew matrix, acceptance-proof matrix, pack-resolution matrix, and cache/offline proof artifacts are not present yet.
- No packcompat transcript for compatibility-pin mode or host-Core/no-Maintenance activation mode is present.
- Existing repository tests still reference public Gastown synthetic cache acceptance and a subpath-agnostic cache key. That is expected before implementation, but it means the plan's cache claims are not yet backed by current tests.

**Required changes:**
- Add an explicit cache-key migration and air-gapped seeding contract. Name whether the new key is computed from normalized root source plus separate subpath, or from the full `source//subpath` string plus parsed subpath, and define how old root-keyed cache entries are diagnosed, promoted, or rejected.
- Make the synthetic-cache cutoff a Slice 2 precondition for all install/check/lock/doctor/init/config paths that can touch `PublicGastownPackVersion`, not only for behavior-changing runtime loads.
- Define the pin-coherence command inputs exactly: `internal/config/PublicGastownPackVersion`, external `gascity-packs/gastown/public-gastown-pins.yaml`, local `support/public-gastown-pin-ledger.yaml`, fresh-init output, lock entry, cache entry, pack digest, behavior-manifest digest, and packcompat transcript paths.
- Extend the rollback matrix with cache and lock states: compatibility-pin rollback may restore the previous pin while leaving new ordinary cache entries inert; activation rollback must not re-enable retired synthetic aliases or embedded Gastown fallback.

**Questions:**
- Will the new `RepoCacheKey` be keyed by normalized repository root plus an explicit subpath field, or by the already-subpathed source string? The answer needs to be stable so old and new cache diagnostics agree.
- What is the supported non-network operator command for seeding the exact public Gastown activation pin into the new subpath-aware cache before an offline upgrade?
