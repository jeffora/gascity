# Camille Sato - Claude

**Verdict:** approve-with-risks

> Lane: required Core/provider pack loading, typed participation provenance,
> deny-by-default loaders, bypass containment, fail-closed behavior. Reviewed
> against the current `implementation-plan.md` (528 lines,
> `updated_at: 2026-06-09T01:20:00Z`) — §"Required System Pack Loader"
> (171–213), §"Pack Registry, Cache, And Retired Source Authority" (215–241),
> §"Bootstrap Fixture Isolation" (304–322), and the
> `RequiredSystemPackParticipation` record in §"Data And State" (353–357). I
> verified the load-bearing repo claims rather than trust the prose. This
> iteration retires the prior live-reload silent-continue blocker and now names
> participation-provenance fields; two trust-boundary gaps remain in this exact
> lane.
>
> Output written to this iteration's reviews dir (`attempt-3/`) alongside the
> iteration-3 Codex sibling. Note: my routed bead `ga-st2xjf` is a dynamic
> expansion instance carrying `gc.attempt=1` while its iteration is 3 (logical
> bead `ga-gpgq5y` is `attempt=3`); the literal `attempt-${gc.attempt}` path
> would have overwritten the unrelated iteration-1 Claude review, so I used the
> live iteration-3 dir and report it via `design_review.output_path`.

**Schema conformance:** Conforms to `gc.mayor.implementation-plan.v1`. Front
matter carries the required keys with `phase: implementation-plan` and no
`design_file`; the eight required top-level sections appear once each in the
required order (`Summary` → `Current System` → `Proposed Implementation` →
`Data And State` → `Testing` → `Rollout And Recovery` → `Open Questions`), and
`Open Questions` is `None`. No appended attempt/review prose in the artifact.

**Top strengths:**
- Fail-closed live reload is now concrete and answers my Q3: no-refresh does not
  repair; on missing/corrupt/stale/shadowed/participation-missing Core the
  controller keeps last-known-good config **only for read-only status/reporting**,
  publishes an event + diagnostic, and refuses behavior-changing
  dispatch/formula/order/hook/prompt/agent-start until a refreshed load succeeds,
  with API and CLI surfacing the same diagnostic (194–200). This retires the
  prior silent-continue blocker and red flag #1 for loader-path callers.
- The two-gate shape is the right deny-by-default structure and directly targets
  red flag #3 (path/digest conflation): a pre-resolution required file-set
  validation (188) separate from a post-resolution typed
  `RequiredSystemPackParticipation` check (189–192), with the record now carrying
  a `resolved config layer id` and an `import edge proving participation in final
  config resolution` (353–357) — not just a digest.
- Retired-source containment is non-destructive and explicit: stale
  `.gc/system/packs/{maintenance,gastown}` and `.gc/runtime/packs/maintenance`
  are ignored by active discovery, reported as legacy, and never auto-deleted or
  selected by new lock/runtime resolution (230–241); and `GC_BOOTSTRAP=skip` is
  explicitly barred from skipping systempacks materialization, file-set
  integrity, participation, collision, or provider checks (318–322).

**Critical risks:**

- **[Major] The only named bypass-containment mechanism cannot catch the bypass
  vectors the plan itself enumerates.** Line 204 requires the scanner to reject
  "aliases, wrappers, method/function values, and manual required-pack include
  assembly," but says it is "modeled on `cmd/gc/worker_boundary_import_test.go`."
  I read that test: it is a pure **substring** scanner (`os.ReadFile` +
  `strings.Contains(content, needle)`; no `go/ast`, `go/types`, or
  `packages.Load`). A substring scan keyed on `config.Load`/`config.LoadWithIncludes`
  provably cannot see an aliased import
  (`cfg "…/internal/config"; cfg.LoadWithIncludes(…)`), a function/method value
  (`load := config.LoadWithIncludes`), a wrapper in another package, or a
  generated/API loader — exactly red flag #2. Compounding it: the typed
  participation gate is validated **inside** `LoadRuntimeCity` (189–192), so a
  caller that bypasses the loader gets an off-loader `config.City` that simply
  carries *no* participation record, and nothing on the behavior-dispatch /
  `internal/api` state-read paths is required to re-check a participation stamp —
  so a bypass drives orders/prompts/formulas/API reads with no gate at all.
  → Treat as **[Blocker]** if the scanner is implemented literally as substring
  matching per the cited model.

- **[Major] The production point for trusted participation is still unnamed,
  leaving room for post-hoc digest/name reconstruction (my Q2).** The record
  lists the right fields (353–357), but the plan never says *who emits* the
  `import edge` / `resolved config layer id`: whether `internal/config`'s
  resolver produces participation edges as a first-class output of resolution, or
  `internal/systempacks` reconstructs them by matching pack id/digest against the
  already-resolved config. Only the former is unforgeable; the latter can be
  satisfied by a user/imported `core` of matching name+digest — the coincidence
  the gate must exclude. Concretely, `config.LoadWithIncludes` already returns
  `(*City, *Provenance, error)` (`internal/config/compose.go:108`); the plan
  should anchor `RequiredSystemPackParticipation` to that resolver `Provenance`
  and state that edges are captured *during* resolution, not derived after it.

- **[Major] Provider-pack selection before resolution is under-defined
  (bootstrapping gap, in-lane via "provider pack loading").** The pre-resolution
  gate validates "Core plus provider packs (`bd` and `dolt` as selected today)"
  (188), and `MaterializeRequiredPacks`/`RuntimeIncludes` take a `provider`
  argument, but `RequiredPackNames(cityPath string)` (174) takes only a path and
  the plan never states how the provider is determined *before* config is
  resolved — when provider can come from root config, includes, `GC_BEADS`/scoped
  env, an `exec:`/file provider, or change across a no-refresh reload. Without an
  authoritative rule (detect-first two-phase, or validate-all-built-in-provider-packs)
  plus a provider-change / missing-provider-pack matrix, the pre-resolution gate
  can validate the wrong provider set (false fail, or miss a required provider
  pack).

- **[Minor] Live-reload enforcement has no single named guard surface.** Lines
  194–200 enumerate the surfaces that must refuse in prose but name no
  controller-held validity object (e.g. a `RuntimeConfigStatus`) that CLI, API,
  dispatch, formula, order, hook, prompt, and agent-start all consult.
  Enumerated-in-prose enforcement risks one entrypoint forgetting the check — the
  same shape as red flag #1 leaking back in through a single un-guarded path. A
  named guard plus an entrypoint-coverage test would close it.

**Missing evidence:**
- The config integration point where trusted required-pack descriptors enter
  resolution and where typed participation records are produced, relative to the
  existing `config.Provenance` returned by `compose.go:108`.
- A statement that the bypass scanner is AST/type-aware (`go/types` /
  `packages.Load`), or a runtime participation-stamp that behavior-dispatch and
  `internal/api` state reads must check.
- The provider-selection algorithm + acceptance matrix (`bd`, `dolt`, file/`exec:`
  provider, scoped `GC_BEADS`, missing provider pack, provider change during
  no-refresh reload).
- The seeded partial-read allowlist: today there are ~40 non-test
  `config.(Load|LoadWithIncludes)` call sites across `cmd/gc` + `internal`; the
  plan defers the allowlist path/schema/owner/freshness test and does not seed
  that inventory (206–208).
- At least one named `internal/api` behavior-driving state-read entrypoint
  covered by a participation/invalidity test (191, 198).
- The concrete event type / diagnostic id emitted on each rejected live reload
  (missing/corrupt/stale/shadowed/participation-missing) and the test that
  asserts it (196).

**Required changes:**
- Specify the scanner as AST/type-aware (resolves import aliases, method/function
  values, wrapper/generated/API loaders, multiline calls) **or** add a runtime
  participation-stamp guard that behavior dispatch and `internal/api` state reads
  must consult; do not rely on a substring scanner modeled on
  `worker_boundary_import_test.go` for the alias/wrapper/function-value cases the
  plan lists.
- Name the participation production point: have `internal/config` resolution emit
  per-pack import-edge/layer records (extending the existing `*Provenance`) that
  `internal/systempacks` validates, and state explicitly that participation
  cannot be recreated from path, pack name, or digest alone. Add negative tests
  for user/imported `core`, repaired-but-unmerged Core, an empty pack with
  matching name+digest, and a required pack contributing only non-agent assets.
- Add the provider-pack selection rule and matrix; state whether provider
  membership is detect-first or validate-all, and cover a provider change during
  `LoadRuntimeCityNoRefresh`.
- Name the controller validity guard (e.g. `RuntimeConfigStatus`), list allowed
  read-only vs refused behavior-changing surfaces, and require an
  entrypoint-coverage test across CLI/API/dispatch/formula/order/hook/prompt/agent-start.
- Seed the partial-read allowlist with the current `config.Load*` inventory and
  define its path, row schema, owner, and freshness test.

**Questions:**
- Is participation produced by `internal/config` resolution (first-class edges)
  or reconstructed by `internal/systempacks` after resolution? The plan's
  unforgeability claim only holds for the former.
- Does any behavior-execution path (dispatch / `internal/api` state read)
  independently check a participation stamp, or is the scanner the sole bypass
  guard? If it is the sole guard, what is its resolution model?
- For the pre-resolution gate, is the provider set detected from `city.toml`/env
  first, or are all built-in provider packs validated unconditionally?
