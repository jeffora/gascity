# Priya Menon - Claude

**Verdict:** approve-with-risks

> Lane note (verify-don't-copy + path handling).
> 1. *Re-grounding.* A prior `pack-resolution-architect_claude.md` exists at
>    `attempt-1/reviews/` carrying its own "re-grounded against the current doc"
>    note. It landed in `attempt-1/` because the persona prompt computes
>    `attempt-${gc.attempt}` and this bead's `gc.attempt=1`. I did not inherit
>    its conclusions: I re-read the current `requirements.md` (md5 `4f1ef50…`,
>    `updated_at 2026-06-09T01:20Z`, identical to `attempt-3/design-before.md`)
>    and re-verified every load-bearing claim against source. Where my finding
>    matches prior Claude it is because the code confirmed it; where it diverges
>    (AC3 scope, the `internal/packs/core` "canonical path") I say so below.
> 2. *Output path.* This review is written to `attempt-3/` (the live
>    `iteration.3` directory whose `reviews/` was empty), **not** the literal
>    `attempt-${gc.attempt}=attempt-1` the prompt prescribes. `attempt-2/synthesis.md`
>    documents that the `gc.attempt`-based path is a known workflow defect that
>    blocked attempt-2 synthesis; `attempt-1/` also holds the original 28KB-doc
>    review set that the literal path would overwrite. Writing attempt-local is
>    the only placement consistent with how synthesis reads. Flagging for the
>    operator; it is outside this document's scope.
> 3. *Schema.* Conformance is the schema-officer's lane. I touch it only where
>    it affects pack resolution: the doc is in v1 shape and `status: questions`,
>    so my contracts live in the ACs and Open Questions, which is the right
>    place for them.

**Top strengths:**
- **The required-Core / retired-Maintenance / explicit-Gastown target is coherent and consistently stated across my surfaces (lane Q1, Q2).** Core is "required in resolved config for real cities" (W6H *How*, AC2); its *absence* is a diagnosable fault with exact source attribution and idempotent repair (negative-path example, AC10, AC11), with a scoped dev/test escape hatch (AC2). The legacy fallback routes are prohibited at every surface — happy path ("No in-tree `examples/gastown/packs/gastown`, `.gc/system/packs/gastown`, or implicit Maintenance fallback"), negative path ("do not silently redirect to a hidden fallback"), and the offline edge case ("never falls back to in-tree examples or system packs"). My red flags "multiple Core source-of-truth paths remain active" and "resolution falls back to in-tree/implicit Maintenance" do not fire on the document.
- **The public Gastown import contract is grounded in current source, not invented.** AC4 / happy-path pin `https://github.com/gastownhall/gascity-packs.git//gastown` at `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`. These match the live constants `PublicGastownPackSource` and `PublicGastownPackVersion` in `internal/config/public_packs.go` verbatim, and the doc correctly demands lock/cache provenance and absence-of-fallback proof rather than asserting them.
- **Deterministic resolution and required-identity protection are first-class acceptance criteria (lane Q3).** AC3 enumerates the actual participants — required Core, provider-conditioned `bd`/`dolt`, root/city/default-rig imports, locked remotes, cached remotes, system packs, local overlays — and bars user imports from shadowing required pack identity "without an explicit collision diagnostic." That is the correct product-level contract for a requirements artifact (the resolver precedence already lives in `internal/config/pack.go`; the doc rightly constrains it rather than re-specifying its internals).

**Critical risks:**
- **[Major] The same-named asset precedence between required Core and an explicitly-imported external Gastown is unstated — and it is a product decision, not a design detail.** The migration *deliberately creates* the split (Core keeps role-neutral assets; Gastown supplies role behavior), so a Gastown city imports both. AC3 governs *pack-identity* shadowing and AC6/AC7 guard split-behavior preservation, but nothing states the winner when Core and external Gastown both define the *same-named* fragment/overlay/script/command/formula/order. Whether explicit Gastown overrides Core, or whether a collision diagnostic is required, decides whether an operator can customize Core behavior via Gastown at all — a user-visible product outcome. Until it is stated, AC7's "Gastown does not lose supported behavior" is not directly verifiable for any asset that exists on both sides.
- **[Minor] "Required Core" inverts current init+doctor+runtime behavior, and the doc lacks one consolidated invariant making the inversion explicit (lane red flag: consistency across init/doctor/runtime).** Today Core is *auto-included* as a system pack (`internal/bootstrap/packs/core/pack.toml`: "Gas City auto-includes this system pack when loading a city" → `.gc/system/packs/core`), while doctor tells operators the opposite — `import_state_doctor_check.go:194` emits `"should be removed; maintenance/core is supplied implicitly"`. The target model requires the reverse on both axes (absence *diagnosed*, not silently injected; explicit Core *kept*, not removed), and it must un-conflate "maintenance/core." AC2 + AC10 + AC11 + AC12 collectively imply this, but no single clause states the shared invariant "Core participates identically across init, doctor, import-state, CLI load, and runtime; absence is diagnosed and repaired, never auto-injected." Without it, design could keep auto-inject and the missing-Core diagnostics (AC10/AC11) become unreachable dead behavior.
- **[Minor] The runtime disposition for a real city missing Core *after* diagnostics, and the AC2 escape-hatch boundary, are undefined.** The negative-path example requires diagnostics to report "instead of failing before diagnostics," but the doc never says whether the city then refuses to start, runs degraded, or blocks until repaired. Today the auto-include path makes "missing Core" effectively impossible at runtime; removing it makes the terminal disposition a new product decision. Relatedly, AC2 grants a dev/test escape hatch without bounding it or saying whether `gc doctor` surfaces its use, so a real city must not be able to masquerade as a partial test config.
- **[Minor] Fresh offline `gc init --template gastown` with no pre-seeded cache is unaddressed.** The offline edge case covers an *already-initialized* city resolving from its lockfile/cache (hit uses it, miss fails actionably). First-time Gastown init with no network and no seeded cache — fail-closed, or require a pre-seed — is the one offline gap left, and the role-neutral default (fail-closed with an actionable diagnostic) should be stated rather than left to design.

**Missing evidence:**
- A stated winner (or required collision diagnostic) for same-named Core-vs-external-Gastown assets in a city importing both — the central unverifiable in AC7.
- The single Core-participation invariant shared by init, doctor, import-state, CLI load, and runtime, plus the explicit no-auto-injection rule that inverts today's auto-include (`pack.toml`) and today's "remove your core import" doctor guidance (`import_state_doctor_check.go:194`).
- The runtime disposition for a real city missing Core after diagnostics, and the exact boundary of the AC2 dev/test escape hatch (what disables the requirement, and whether doctor reports it).
- Fresh offline `gc init --template gastown` behavior with no network and no seeded cache.

**Required changes:**
- Add a product-level decision (new Open Question, or an AC) for same-named asset precedence between required Core and explicitly-imported external Gastown: explicit-Gastown-wins, Core-wins, or collision-diagnostic-required. This is the one genuine pack-resolution gap blocking a clean move from `questions` to `approved`.
- Add one consolidated "required Core" invariant to AC2/AC10/AC11 (or W6H *How*): Core participates identically across init, doctor, import-state, CLI load, and runtime; absence is diagnosed and repaired, never silently auto-injected; and legacy guidance that explicit core imports "should be removed because core is supplied implicitly" is retired and inverted. This pins the mechanism end-to-end so missing-Core diagnostics cannot become dead behavior.
- State the runtime disposition for a real city missing Core after diagnostics (refuse / degraded / block-until-repair) and bound the AC2 escape hatch so real cities cannot bypass the Core requirement undetected.
- State fresh offline `gc init --template gastown` behavior (fail-closed with an actionable diagnostic unless a configured/pre-seeded cache or mirror exists).

**Questions:**
- In a city importing both required Core and external Gastown, does an explicit Gastown import override a same-named Core asset, or must collisions raise a diagnostic? (Drives AC7 verifiability and whether Gastown can customize Core behavior.)
- Post-migration, does `core` remain a *required* pack identity inside the bootstrap collision gate — so AC3's "may not shadow required pack identity" protects it — or does it become an ordinary explicit import under normal precedence?
- Is offline support intended to rely solely on the pinned repo cache (per the edge case), with no binary-embedded Gastown bundle — i.e., must fresh `gc init --template gastown` require network or a pre-seeded cache?
