# Lena Hoffmann - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The rollout separates compatibility and activation pins, so Gas City can prove public Gastown coexistence before the no-Maintenance activation slice consumes the stricter pin (`design-before.md:124`-`design-before.md:128`).
- The behavior manifest binds public-pack rows to an immutable public Gastown commit and the consuming `internal/config/PublicGastownPackVersion` value (`design-before.md:150`-`design-before.md:151`).
- Cache identity is correctly subpath-aware and digest-verified: `RepoCacheKey` includes normalized source, exact commit, and subpath, while promotion/read hits verify source, commit, subpath, pack digest, and manifest digest (`design-before.md:230`-`design-before.md:236`).

**Critical risks:**
- [Major] Offline cache behavior is implied by the cache model but not made into an explicit acceptance proof in this plan. The requirements call for offline cache-hit/cache-miss behavior, and this design should name tests that prove an exact pinned public Gastown cache hit works while a missing or wrong source/commit/subpath/digest cache entry fails without falling back to in-tree examples or synthetic system packs.
- [Minor] The plan says stale synthetic Gastown/Maintenance cache entries cannot satisfy a public `sha:` pin, but it should require a negative test for cache promotion laundering: stale bytes under a valid-looking commit key must fail if pack or manifest digest differs.
- [Minor] The rollback narrative is present, but the "old binary + new compatibility pack" and "new binary + old locked pack" rows should name the exact diagnostic/proof artifact that operators can use to distinguish unsupported version skew from transient network/cache failure.

**Missing evidence:**
- Offline public-pack cache-hit and cache-miss test names tied to exact source, commit, subpath, pack digest, and manifest digest.
- A negative test for stale synthetic alias or stale ordinary-cache bytes that match path shape but not digest/provenance.
- A proof that lock refresh and install paths use the same `internal/packsource` classifier and `RepoCacheKey` identity as runtime load and packcompat.
- Operator-facing version-skew diagnostic examples for new binary + old locked public Gastown pack.

**Required changes:**
- Add explicit offline cache-hit/cache-miss tests to the packcompat or cache test plan, including wrong-subpath, wrong-digest, stale-synthetic, and missing-cache cases.
- Require cache promotion/read validation to fail closed on digest mismatch even when source, commit, and subpath strings match.
- Tie lock refresh, install, runtime load, doctor, and packcompat to the same public-pack identity proof so no path can refresh through a retired synthetic alias.
- Name the version-skew diagnostic artifact or golden output for rollback/offline ambiguity.

**Questions:**
- Which test package owns exact public-pack cache hit/miss coverage: `test/packcompat`, a config/cache package, or both?
- Does `public-gastown-pins.yaml` record both source URL and subpath in addition to commit, or does that live only in Gas City's `PublicGastownPackSource`?
- What will an offline operator see when the lockfile names the activation commit but the ordinary remote cache has only the compatibility commit?
