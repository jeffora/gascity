# Lena Hoffmann

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The first public-pin trust boundary is still unsafe. All sources agree the plan must prove public Gastown pins through ordinary remote/cache provenance before trusting them. Codex notes the repo already consumes `PublicGastownPackSource`/`PublicGastownPackVersion` through bundled synthetic fallback, so Slice 2 is not a future first-use point unless the plan explicitly freezes or removes that authority before public-pin proof passes.
- [Blocker] Subpath-scoped proof is sequenced inconsistently. The plan says source, commit, subpath, pack digest, and behavior-manifest digest are verified before the first pin update, but rollout text leaves subpath-aware ordinary cache enforcement or proof to Slice 6. That permits Slices 2-5 to rely on source+commit cache/lock identity for a `//gastown` pin without enough subpath and digest provenance.
- [Blocker] The pin-coherence gate is not yet a non-circular authority. The plan compares useful outputs, but it must name one authoritative ledger/reconciliation command and independently compare `PublicGastownPackVersion`, `PublicGastownPackSource`, external public pins, the Gas City pin ledger, packcompat inputs/transcripts, fresh init output, lock/cache provenance, source URL, subpath, immutable commit, pack digest, and behavior-manifest digest.
- [Major] Cache promotion and read-hit integrity are asserted, not designed. Current promotion clones directly into the final cache path, and the plan does not yet specify process-unique staging, validation before publish, atomic rename, interrupted-promotion cleanup, or the behavior when another writer has already published a valid target.
- [Major] Digest and schema semantics remain underspecified. The plan needs the canonical lock/cache fields, schema versioning and old-schema behavior, the canonical pack-tree and behavior-manifest digest algorithms, volatile-field exclusions, mismatch diagnostics, offline behavior, and repair/revalidation paths.
- [Major] Synthetic alias retirement must cover both synthetic cache directories and retired bundled source URLs. Claude specifically warns that an old lock pinned to the former in-tree bundled source at an old commit could still be fetchable as an ordinary remote unless `internal/packsource` rejects the retired source URLs regardless of commit.
- [Major] Rollback and version-skew coverage is incomplete. In particular, Slice 5b can fold Maintenance assets into Core and then roll back by re-enabling Maintenance, creating a duplicate-active path unless the plan defines an un-fold step or declares the fold one-way with manual recovery.

**Disagreements:**
- There is no verdict disagreement: all three reviewers block.
- Claude and Codex/DeepSeek disagree on what "subpath-aware cache identity" should mean physically. Claude says the existing repo clone cache key should remain normalized clone URL plus commit, preserving clone sharing and existing invariant tests, while subpath plus digests are enforced in the lock/proof layer. Codex and DeepSeek phrase the fix as subpath-aware ordinary cache keys before Slice 2. Assessment: the plan must distinguish physical clone storage from logical proof identity. It can keep shared clone storage only if every lock, promotion, read hit, and pin-coherence check binds source URL, subpath, commit, pack digest, and behavior-manifest digest before public pin consumption.
- DeepSeek treats full digest verification on every read hit as a performance blocker and proposes a validated marker strategy. Claude and Codex focus on fail-closed digest correctness. Assessment: marker-based fast paths are acceptable only if marker contents, tamper/drift detection, invalidation, and explicit full revalidation are part of the fail-closed design.
- Claude views the existing two-pin compatibility/activation model as directionally strong; Codex emphasizes that current synthetic public fallback means the migration window is already active. Assessment: the two-pin rollout is useful, but it must begin with a compatibility freeze or authority removal for public synthetic fallback.
- Claude frames the retired-source risk as old bundled source URLs remaining fetchable from git history; Codex frames it as current synthetic fallback authority. Assessment: both paths need closure: public `sha:` pins must not resolve through synthetic aliases, and retired in-tree source URLs must be rejected even when an old commit contains the subpath.

**Missing evidence:**
- Which artifact is authoritative for `PublicGastownPackVersion`: external `public-gastown-pins.yaml`, Gas City `support/public-gastown-pin-ledger.yaml`, or a defined reconciliation between them.
- The owner package, command/test name, fixture inputs, and diagnostics for the pin-coherence gate.
- Whether packcompat derives pins from an independent ledger or from the Go constant it is meant to validate.
- The exact lock/cache provenance schema, including source URL, subpath, commit, pack digest, behavior-manifest digest, cache entry id or proof id, validation timestamp, cache kind, and schema version.
- The old-binary/new-lock, new-binary/old-lock, old-cache, offline-cache-miss, stale-alias, compatibility-pin, activation-pin, activation-rollback, and post-activation-downgrade behavior.
- The cache publication protocol for concurrent writers, interrupted promotions, and partially published targets.
- The digest canonicalization rules: stable path ordering, slash normalization, file mode and symlink policy, ignored paths, YAML/JSON canonicalization, and exclusions for volatile fields such as generated timestamps.
- The read-hit validation strategy: full recursive digest, tamper-evident marker, marker invalidation, lightweight drift checks, explicit full revalidation command, and fail-closed mismatch behavior.
- Network-disabled tests for exact cache hit, cache miss, digest mismatch, missing subpath, stale synthetic alias rejection, install, lock refresh, promotion, and runtime load.
- The concrete Slice 5b rollback mechanics after Maintenance assets have been folded into Core.

**Required changes:**
- Rewrite the early rollout so public-pin proof becomes true before `PublicGastownPackVersion` is updated or trusted: remove or freeze public Gastown/Maintenance synthetic fallback authority, enforce subpath-scoped proof identity, add the lock/cache provenance schema, reject stale aliases and retired bundled source URLs, and run offline/cache proof tests in Slice 2 or an earlier prerequisite.
- Correct the `RepoCacheKey` design language. If clone storage remains keyed by normalized clone URL plus commit, say so and preserve the existing invariant tests; bind subpath and digests in the lock/proof/read-hit layer. If the physical cache key changes to include subpath, name that as an explicit breaking change and account for the current tests.
- Define one non-circular pin-coherence command and authoritative ledger/reconciliation model that compares the Go constants, source URL, external and internal ledgers, packcompat evidence, fresh init output, lock/cache provenance, commit, subpath, pack digest, and behavior-manifest digest.
- Specify atomic cache publication in `internal/packman/cache.go`: process-unique staging checkout, validation before publish, atomic rename to the canonical target, success when a concurrently published target is already valid, and cleanup/retry behavior for failures.
- Specify promotion and read-hit validation semantics, including digest algorithms, canonicalization rules, marker or full-hash policy, mismatch diagnostics, invalidation, repair, and explicit full revalidation.
- Define lock/cache schema migration and version-skew behavior, including old-schema refusal or migration, old binary behavior after schema-v2 writes, offline diagnostics, and rollback/downgrade policy.
- Extend the rollout/recovery matrix to cover compatibility pin, activation pin, old/new binaries, old/new locks, old/new caches, stale synthetic aliases, offline states, activation rollback, post-activation downgrade, and compatibility-pinned cities under post-activation binaries.
- Fix Slice 5b rollback by un-folding moved assets before re-enabling Maintenance, or declare the Maintenance fold one-way and document the manual recovery procedure.
