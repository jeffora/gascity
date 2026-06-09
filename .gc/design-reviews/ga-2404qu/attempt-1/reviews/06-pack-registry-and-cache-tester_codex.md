# Marcus Driscoll - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design requires public Gastown pins to resolve through the ordinary remote path at an exact immutable commit/subpath, never through embedded bytes, bundled aliases, old cache namespaces, or stale `.gc/system/packs/gastown` materialization (`design.md:750`, `design.md:1040`, `design.md:2935`).
- Registry cleanup is decomposed by surface: `PublicRepository`, `publicSubpathForPack`, `RepoCacheKey`, `SyntheticContentHash`, `requiredBuiltinPackNames`, required-pack materialization, and retired-source recognizers each have an intended disposition and gate (`design.md:887`).
- Provider continuity is treated as content/provenance, not path-count preservation: `bd` and `dolt` participate through typed required-pack records, byte continuity is required for untouched files, and role-cleaning exceptions need old/new witnesses (`design.md:293`, `design.md:2064`, `design.md:2674`).

**Critical risks:**
- [Major] The public Gastown cache-key contract needs a canonical source-normalization matrix. The design says `RepoCacheKey` includes the normalized source plus exact `PublicGastownPackVersion` (`design.md:2947`) and that public source normalization remains for real public repo sources only (`design.md:892`), but it does not enumerate accepted source spellings or rejection cases. Without explicit tests for `github.com/...`, HTTPS, SSH, `.git` suffixes, subpath forms, case/path normalization, and lookalike sources, a stale synthetic alias or wrong remote could still collide with the ordinary cache key.
- [Major] Provider-pack continuity gates are correct but scattered. The final activation slice names "provider pack continuity and role-cleaning provenance tests" (`design.md:3436`), while the concrete required witnesses live much earlier (`design.md:293`, `design.md:2674`). If those are not copied into `slice-gates.generated.yaml`, an implementer could satisfy the slice with formula/order smoke tests but miss byte-identical digest proof for untouched `bd`/`dolt` assets after Core repair.
- [Minor] The optional legacy synthetic-cache promotion path is underspecified. The design allows legacy synthetic public Gastown cache bytes to be copied only through an explicit promotion helper into the ordinary cache key (`design.md:227`), while later sections correctly forbid selecting old cache namespaces directly (`design.md:1048`). That is safe only if the helper is explicit, non-automatic, digest-validated against the exact source/commit/subpath, and covered by negative tests; otherwise it can become an accidental fallback path.

**Missing evidence:**
- A source-normalization table for public Gastown inputs, with expected cache key, accepted/rejected status, and lock/install behavior.
- Generated `slice-gates.generated.yaml` rows that bind provider byte-continuity and Core repair isolation to exact commands, not only prose.
- A clear decision on whether synthetic-cache promotion is in scope; if yes, the design needs helper ownership, invocation path, validation inputs, and tests.

**Required changes:**
- Add a canonical public-source normalization/test matrix for `RepoCacheKey`, `IsSource`, `NameForSource`, install-lock generation, and materialization.
- Copy the `bd`/`dolt` untouched-file digest proof, provider matrix, and Core repair isolation witnesses into the final slice gate artifact contract.
- Either remove the synthetic-cache promotion mention or define it as an explicit operator/helper path that never runs during normal init/load/doctor and can only populate an ordinary remote cache entry after exact digest validation.

**Questions:**
- Which public Gastown source spellings are valid inputs, and which one canonical string becomes part of the cache key?
- Is legacy synthetic-cache promotion intended for this migration, or should operators pre-populate the ordinary remote cache through the normal remote-pack installer only?
