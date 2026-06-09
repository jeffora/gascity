# Petra Novak - Claude

**Verdict:** approve-with-risks

Lane: builtinpacks registry; embed path migration; Maintenance retirement;
downstream references. Reviewed `plans/core-gastown-pack-migration/requirements.md`
(`status: draft`, `updated_at 2026-06-09T17:23:58Z`, `Open Questions: None`)
against the live build surface in `internal/builtinpacks/registry.go`,
`internal/bootstrap/`, `cmd/gc/embed_builtin_packs.go`, and the dolt/maintenance
script trees. Judged schema conformance against
`assets/skills/mayor/requirements.schema.md` (`gc.mayor.requirements.v1`). Schema
shape is conformant in my lane: exact top-level section order, happy/negative/edge
examples present, no review-residue markers. Adjacent matters (resolution-matrix
internals, doctor exit codes, rollout policy) are left to their reviewers.

**Top strengths:**
- The embed/registry/materialization closure is required *as a unit*, not
  piecemeal. AC2 + the W6H "How much / scale" row + the "release build embeds
  and materializes Core" happy path collectively demand that "every Go embed
  package, builtin registry function, materialization command, direct
  hook/import path, generated or synthetic layout/hash, test fixture" close over
  the single `internal/packs/core` authority. That maps cleanly onto the real
  consumers (`registry.go:All()`, `cmd/gc/embed_builtin_packs.go`,
  `internal/packman/cache.go`, the five `MaterializeBuiltinPacks` call sites).
- The downstream-script red flag is handled with surgical precision. AC5 names
  the exact `port_resolve.sh` → `dolt-target.sh` shape and requires surviving
  Core/provider/public-pack assets to be rehomed/inlined/retired *before*
  Maintenance can be removed, backed by a `source-consumer-closure.yaml`
  validator row (Example Mapping edge case). This matches the live dependency in
  `examples/dolt/assets/scripts/port_resolve.sh`.
- Maintenance retirement is specified as fail-closed across all the bundling
  vectors that actually exist: "not bundled, public-source recognized,
  auto-included, materialized as an active system pack, selected by lock
  refresh, or presented as an implicit dependency" (AC5), with the synthetic
  content hash / generated layouts called out as "generated or synthetic
  layouts/hashes" in AC2.

**Critical risks:**
- **[Major] The orphaned `//go:embed packs/**` glob compile break is not named,
  and a name-based closure can skip it.** Core is embedded *twice*: once via the
  dedicated `internal/bootstrap/packs/core/embed.go` (`core.PackFS`) and again
  via the wildcard `//go:embed packs/**` in `internal/bootstrap/bootstrap.go:22`.
  `internal/bootstrap/packs/` contains *only* `core`. The doc requires
  `internal/bootstrap/packs/core` to "migrate into that authority, be deleted,
  or be isolated as a non-runtime fixture." If it is moved/deleted, the
  `packs/**` pattern matches zero files and **Go fails the build** — the literal
  red flag "stale bootstrap or embed paths break compilation." AC2's generic
  "every Go embed package … must close over that source authority" *should*
  cover this, but a `source-consumer-closure` that enumerates named `PackFS`
  embed packages can silently miss a wildcard `//go:embed` directive in an
  otherwise-unrelated package. The evidence contract needs to force discovery of
  glob embeds, not just named pack embeds.
- **[Major] The bundled offline public-alias serving path contradicts AC16 and
  is not enumerated as a build consumer to retire.** `registry.go:31-32`
  documents that the binary "can serve its bundled public-pack aliases from the
  embedded pack set when the network is unavailable during init or doctor
  repair," implemented via `publicSubpathForPack` (hardcodes `gastown` and
  `maintenance`, lines 126-133) feeding the `PublicRepository` entries in
  `syntheticPackLayouts()`. AC16 mandates offline = fail-closed and "never
  select embedded, in-tree, or retired synthetic content," and AC4 forbids
  reliance on "bundled synthetic public aliases." The end-state *behavior* is
  specified, but removing only the `maintenance` row from `All()` leaves the
  synthetic-public-alias machinery — and the embedded **gastown** offline alias
  — intact, which would let embedded synthetic content satisfy a public pin
  offline. The ledger (AC6) / closure (AC5) must explicitly list
  `publicSubpathForPack`, the `PublicRepository` synthetic layouts, and this
  offline-serving code path as consumers to retire/repoint.
- **[Minor] Hardcoded `required := []string{"core", "maintenance"}` is the
  concrete auto-include of Maintenance and must be an explicit closure row.**
  `cmd/gc/embed_builtin_packs.go:237` hardcodes Maintenance as an always-
  refreshed *required* pack, and `registry.go:All()` (line 56) lists it.
  Removing it also shifts `SyntheticContentHash()` and the synthetic file-set
  validation (`syntheticRepoAllowedPaths`). AC5 covers the behavior, but the
  closure validator must treat these in-Go hardcoded *lists* (not only file-path
  references) as consumers, or the materialization layer keeps a stale
  `"maintenance"` requirement.
- **[Minor] Schema: mandating exact `support/*.yaml` filenames is the doc's
  closest brush with the "no implementation file assignments" rule.** The schema
  (AC1) forbids choosing implementation files. The ACs name concrete artifacts
  (`asset-migration-ledger.yaml`, `source-consumer-closure.yaml`, etc.). The doc
  defends these as "acceptance evidence, not inline implementation design," and I
  read them as conforming evidence contracts rather than Go file assignments —
  but a strict schema lane may push back. Not a blocker in my lane.

**Missing evidence:**
- No statement that `internal/packs/core` does not yet exist. The doc treats it
  as "the accepted end-state source authority," but the live tree only has
  `internal/bootstrap/packs/core`. This is correct *requirements* framing
  (desired end state), so it is not a defect — but the AC6 ledger's "source
  snapshot frozen before any … root is deleted" must capture the *current*
  `internal/bootstrap/packs/core` content, and the doc does not explicitly tie
  the snapshot origin to that legacy path.
- The doc does not enumerate that Core is dual-embedded (dedicated PackFS +
  `bootstrap.go` glob). Whether the source-consumer closure is expected to find
  both embed sites is left implicit (see Major #1).

**Required changes:**
- In AC2 and AC5 verification (the `source-consumer-closure.yaml` contract),
  require the embed-closure to cover **wildcard `//go:embed` directives**
  (explicitly including `internal/bootstrap/bootstrap.go`'s `packs/**`), not only
  dedicated `PackFS` embed packages, and require a **compiling-build witness**
  (`go build ./...`) after Core relocation so an orphaned glob is caught as a
  gate failure rather than at release time.
- Require the AC5 closure / AC6 ledger to enumerate the synthetic-public-alias
  build sites — `builtinpacks.publicSubpathForPack`, the `PublicRepository`
  layouts in `syntheticPackLayouts()`, and the offline embedded-alias serving
  path described at `registry.go:31-32` — as rows with an explicit
  retire/repoint decision, so AC4/AC16 fail-closed behavior is provably wired,
  not just asserted.
- State in AC5 (or its verification) that hardcoded in-Go pack lists
  (`registry.go:All()` and `cmd/gc/embed_builtin_packs.go`'s
  `requiredBuiltinPackNames` → `{"core","maintenance"}`) are in-scope closure
  consumers, with a witness that `maintenance` is absent from both after
  retirement and that `SyntheticContentHash()` / synthetic file-set validation
  still pass for Core, bd, and dolt.

**Questions:**
- Is the `source-consumer-closure.yaml` validator expected to *statically
  discover* embed/registry/materialization consumers (so a new wildcard embed or
  a new `MaterializeBuiltinPacks` call site is caught automatically), or only to
  validate a hand-authored row set? If the latter, the dual-embed glob and the
  offline public-alias path must be pre-listed or they will be missed.
- Does "Maintenance is … not public-source recognized" (AC5) require deleting
  the dead `case "maintenance"` branch in `publicSubpathForPack`, or only that
  it never resolves at runtime? The role-neutrality / docs-audit lanes may treat
  a residual literal `"maintenance"`/`"gastown"` string in that switch as
  in-scope; my lane only needs the resolution path dead, but the answer affects
  whether this is a build change or a string-scan change.
