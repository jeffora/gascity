# Petra Novak

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] The requirements set the correct end-state Core authority. All sources agree that `internal/packs/core` is the target release-bundled Core source and that Go embed declarations, builtin registry entries, materialization commands, generated hashes/layouts, direct hook/import paths, fixtures, and tests must close over that authority.
- [Info] Maintenance retirement is explicit and fail-closed. The requirements say Maintenance must not be bundled, public-source recognized, auto-included, materialized as an active system pack, selected by lock refresh, or treated as an implicit dependency.
- [Info] `source-consumer-closure.yaml` is the right acceptance artifact class, but it is load-bearing. Claude and Codex both warn that it must parse real embed declarations, registry rows, materialization calls, generated/synthetic layouts, scripts, docs, tests, and shell source calls rather than validating a prose inventory.
- [Major] Wildcard `//go:embed` discovery is a build-critical closure requirement. Claude identifies that Core is embedded both through the dedicated Core `PackFS` path and through `internal/bootstrap/bootstrap.go`'s wildcard `packs/**` glob. If `internal/bootstrap/packs/core` is deleted or emptied without changing that glob, Go compilation can fail because the pattern matches no files.
- [Major] The synthetic public-alias serving path must be enumerated and retired/repointed. Claude flags `publicSubpathForPack`, `syntheticPackLayouts()`, `PublicRepository` synthetic layouts, and offline bundled public-pack alias serving as concrete build/runtime consumers that could continue satisfying public Gastown or Maintenance behavior despite AC4/AC16's fail-closed requirement.
- [Major] Hardcoded Maintenance pack lists are closure consumers, not incidental strings. Claude and DeepSeek both point at the builtin registry and `cmd/gc/embed_builtin_packs.go` required-pack list as places where `"maintenance"` can remain auto-included or materialized unless explicitly removed and tested.
- [Major] The `dolt-target.sh`/`port_resolve.sh` dependency is mischaracterized in at least part of the source evidence. DeepSeek reports, and a live spot check confirms, that `examples/gastown/packs/maintenance/assets/scripts/dolt-target.sh` sources `examples/dolt/assets/scripts/port_resolve.sh`; `examples/dolt/port_resolve_test.go` still asserts against the legacy Maintenance path. Any requirements or implementation-plan text that treats `port_resolve.sh` as sourcing `dolt-target.sh` must be corrected before this can serve as a downstream-reference witness.
- [Major] Directory-level materialization needs an atomicity and concurrency contract. DeepSeek flags that writing directly into `.gc/system/packs/core` can leave partial or corrupted pack directories visible after a crash or concurrent command unless materialization uses process-unique staging and atomic promotion.
- [Minor] The legacy `internal/bootstrap/packs/core` disposition remains flexible in requirements, but the later design must choose one unambiguous path: delete it, isolate it as a non-runtime fixture, or keep a compatibility shim that cannot satisfy runtime Core.
- [Minor] Active code, scripts, and tests need a path-string scanner for retired Maintenance and in-tree Gastown paths so references such as `.gc/system/packs/maintenance` or `examples/gastown/packs/maintenance` cannot survive accidentally.
- [Minor] The AC filenames for support artifacts are acceptable as acceptance evidence in this lane, but they are close to the schema boundary that forbids implementation-file assignment; the schema lane may need to confirm that interpretation.

**Disagreements:**
- Claude and Codex describe the `port_resolve.sh` to `dolt-target.sh` helper shape as covered by the requirements. DeepSeek says the live dependency direction is the opposite. My assessment: the live files confirm the important preservation risk is `dolt-target.sh` sourcing `port_resolve.sh`, plus a test that still reads the retired Maintenance copy. The synthesis should treat this as a required correction.
- Claude rates the wildcard embed glob as a Major compile-break risk; DeepSeek also calls it out but marks it Minor; Codex frames it as an implementation-plan concern. My assessment: it is Major for this lane because it can break `go build` during source relocation even if runtime resolution is correct.
- DeepSeek raises atomic directory materialization while Claude and Codex focus more on embed/registry closure. My assessment: it belongs in this persona lane because materialization safety is explicitly part of the reviewer scope.
- Codex says no changes are required before requirements approval, while Claude and DeepSeek require concrete hardening. My assessment: approve with risks, but the required closure and materialization evidence must be added before implementation slices can safely delete or relocate sources.

**Missing evidence:**
- A source-consumer closure that statically discovers wildcard `//go:embed` directives, dedicated embed packages, builtin registry entries, hardcoded required-pack lists, materialization call sites, generated synthetic layouts/hashes, shell source calls, docs, fixtures, and tests.
- A compiling-build witness after Core relocation or legacy bootstrap deletion, especially covering the `internal/bootstrap/bootstrap.go` wildcard glob.
- Rows for `publicSubpathForPack`, `syntheticPackLayouts()`, `PublicRepository` synthetic layouts, and offline embedded public-alias serving with explicit retire/repoint decisions.
- A verified destination for `dolt-target.sh` after Maintenance retirement, plus updated `examples/dolt/port_resolve_test.go` coverage that no longer depends on the retired Maintenance path.
- Atomic materialization invariants and tests: process-unique temporary sibling directories, complete writes before promotion, atomic rename or equivalent replacement, crash cleanup, and concurrent command behavior.
- A path-string lint or packlint rule for active code/scripts/tests that rejects retired Maintenance and in-tree Gastown paths outside approved migration fixtures.
- Representative checks proving Maintenance is absent from active bundled packs, materialized system packs, lock refresh, synthetic content hash/file-set validation, and public-source recognition after retirement.

**Required changes:**
- Extend the AC2/AC5 source-consumer closure contract to require static discovery of wildcard `//go:embed` directives, explicitly including `internal/bootstrap/bootstrap.go`'s `packs/**`, and require a `go build ./...` witness after Core relocation.
- Require the AC5 closure or AC6 ledger to enumerate the synthetic-public-alias build sites: `publicSubpathForPack`, `syntheticPackLayouts()`, `PublicRepository` layouts, and the offline embedded-alias serving path, each with an explicit retire or repoint decision.
- Treat hardcoded in-Go pack lists as closure consumers, including the builtin registry rows and `cmd/gc/embed_builtin_packs.go` required-pack list, with checks proving `maintenance` is removed from active materialization paths.
- Correct the documented `dolt-target.sh`/`port_resolve.sh` dependency direction, choose a surviving owner for `dolt-target.sh` or inline it, and update `examples/dolt/port_resolve_test.go` to assert against the new location.
- Require directory-level atomic materialization for required packs: write to a process-unique temporary sibling, fully validate, atomically promote to the final `.gc/system/packs/<name>` path, and clean orphaned temp directories.
- Coordinate registry cleanup, filesystem deletion, and wildcard embed replacement in the same implementation slice so there is no intermediate compile-broken state.
- Add a packlint/path scanner that fails active non-fixture code, scripts, and tests containing retired Maintenance or in-tree Gastown paths.
