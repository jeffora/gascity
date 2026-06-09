# Camille Sato - Claude

**Verdict:** approve-with-risks

Lane: required Core and provider pack loading, typed participation provenance,
deny-by-default loaders, bypass containment, fail-closed behavior. Reviewed
`plans/core-gastown-pack-migration/implementation-plan.md` (`updated_at`
2026-06-09T13:20:59Z) against the current `requirements.md` (`updated_at`
2026-06-09T17:23:58Z) and `implementation-plan.schema.md`, with every
load-bearing claim re-verified directly against the live tree.

**Prior-iteration blocker is reconciled — recorded so it is not re-raised.** An
earlier iteration blocked on an AC2 source-authority contradiction (plan moves
Core to `internal/packs/core`; requirements then named
`internal/bootstrap/packs/core` as the authority). The requirements were updated
at 17:23:58Z and now agree with the plan: Problem Statement (`requirements.md:30`)
calls `internal/packs/core` "the accepted end-state source authority" and the
legacy `internal/bootstrap/packs/core` tree "a legacy source root that must
migrate... be deleted, or be isolated"; AC2 (`:108`) names `internal/packs/core`
as "the sole end-state canonical source authority"; How (`:79`) makes the
bootstrap root "migration input only." That conflict is resolved in the plan's
favor and is no longer a blocker.

Re-verified this pass: `internal/systempacks` and `internal/packs/core` do not
exist yet (genuinely new, as proposed); current Core is
`internal/bootstrap/packs/core`. Real loader surface is `config.Load`
(`internal/config/config.go:3777`), `config.LoadWithIncludes` /
`config.LoadWithIncludesOptions` (`internal/config/compose.go:108,113`).
`config.Provenance` (`internal/config/compose.go:72-91`) carries only `Root`,
`Sources`, `Imports` (binding-name → source-file, implicit = sentinel
`"(implicit)"`), and per-field `Agents`/`Rigs`/`Workspace` maps — no descriptor
id, digest, or resolved-layer id.

**Top strengths:**
- Fail-closed live-reload contract answers lane question 3 at the design level:
  `LoadRuntimeCity` returns a typed result with
  `Mode ∈ {ready, read_only_degraded, blocked}`; only `ready` drives dispatch/
  formula/order/hook/prompt/session-start; "No-refresh reloads do not repair"
  keeps last-known-good for read-only status while refusing behavior-changing ops
  (lines 251-286). Right shape to close red flag #1, and "callers... may not
  reinterpret diagnostics locally" closes a diagnostic-reinterpretation bypass.
- Two ordered, descriptor-id-bound fatal gates (lines 259-271) correctly separate
  "files materialized and digest-valid" (Gate 1, before any `pack.toml` is
  trusted) from "this descriptor participated in the resolved layer graph"
  (Gate 2). "A path match without descriptor participation, or participation
  without a matching validated file set, is `blocked`" is the right conceptual
  defense against red flag #3, and Gate 2 covers provider packs (`bd`/`dolt`),
  not just Core.
- The bypass-scanner *intent* is correct (lines 296-302): aliases, selector
  values, function variables, wrapper functions, and package-local helpers "count
  as bypasses" — the right answer to red flag #2.

**Critical risks:**
- **[Major] Fail-closed has a live `--no-strict` escape the plan never closes.**
  Verified: the reload path the new loader must absorb — `tryReloadConfig`
  (`cmd/gc/controller.go:895`) — hand-builds includes via
  `cityConfigIncludesWithBuiltinPacks(...)` (`:900`, the exact "manual
  required-pack include assembly" the plan forbids) and downgrades reload warnings
  via `splitStrictConfigWarnings(reloadWarnings)` (`:926`) with
  `reloadStrictWarningHint = "use --no-strict to disable strict checking"`
  (`:878`); the doc comment confirms tables "stay strict-fatal unless --no-strict"
  (`:893`). The plan calls the two Core gates "fatal" (lines 259-271) but never
  states they are NOT disableable by `--no-strict` or the warning-downgrade path.
  If `--no-strict`/`splitStrictConfigWarnings` can suppress required-Core integrity
  or participation, a corrupted or partially materialized Core can load — a
  deny-by-default violation and a direct hit on red flag #3.
- **[Major] Typed participation presupposes provenance the config layer does not
  emit (lane Q2 / red flag #3).** `RequiredSystemPackParticipation` is specified
  to carry "embedded source id," "resolved config layer id," and an "import edge
  proving participation in final config resolution" (lines 234-243, 541-545).
  Verified: `config.Provenance` (`internal/config/compose.go:72-91`) is name-keyed
  (`Imports[name]=sourceFile`, `"(implicit)"` sentinel) with no per-layer
  embedded-source identity or digest. Because it keys by binding name it cannot
  distinguish a required `core` descriptor from a user import also named `core` —
  the precise same-name shadow Gate 2 claims to reject (lines 267-269). The plan
  names the new `config.WithRequiredSystemPacks` option but never states that
  `config.Provenance` must be extended to a trusted-identity edge, and Current
  System omits `Provenance` entirely. As written, Gate 2 risks collapsing into the
  path coincidence it forbids.
- **[Major] Scanner scope "behavior-driving `internal/` packages" is undefined and
  demonstrably undecidable.** The scanner is scoped to "`cmd/gc` and
  behavior-driving `internal/` packages" (lines 288-294) with no enumeration.
  Verified why that is not deny-by-default: `internal/dispatch/control.go:836`
  calls `config.LoadWithIncludes` and is behavior-driving (must be in scope), while
  `internal/doctor/checks.go:{80,1005,1917,2042}` call `config.Load` as
  bootstrap-diagnostic reads that must be *allowlisted* (forcing them through the
  fail-closed loader breaks the AC11 bootstrap-only diagnostics). There are 8
  non-test `internal/` direct loader sites outside `internal/config`. A hand-picked
  "behavior-driving" pre-filter is itself a bypass vector: misclassify or newly add
  a package and its direct `config.Load*` escapes the gate.
- **[Major] The bypass-scanner cites a model that cannot deliver its own stated
  power.** The plan says "Add scanner tests modeled on
  `cmd/gc/worker_boundary_import_test.go`" (line 288) *and* that the validator "is
  AST/type-aware" (line 297). Verified: that test is pure
  `strings.Contains(content, needle)` (`worker_boundary_import_test.go:45`) — a
  literal-substring scan that cannot catch an aliased, function-valued, wrapped, or
  selector-reached `config.Load`. Following the citation literally yields red flag
  #2. The scope is also silent on generated code (`internal/api/genclient`,
  generated server handlers), which red flag #2 names explicitly; `internal/api/`
  has zero direct `config.Load*` today, so the requirement is to keep it provably
  at zero through regeneration.
- **[Major] The readiness guard (`RequireReady`) is convention, not an enforced
  boundary — unlike the loader scanner.** `LoadRuntimeCity` produces a live
  `Config` regardless of `Mode`; behavior-changing entry points are only described
  as something that "receive ... and call `RequireReady(op)`" (lines 281-286).
  Nothing proves every such entry point calls it; any path that reads
  `snapshot.Config` without calling `RequireReady` silently resurrects red flag #1.
  The example `RuntimeGuard` unit test (lines 628-631) is not a totality proof.
  Fail-closed is voluntary while the loader scanner is enforced — an asymmetry that
  must be closed.
- **[Minor] The scanner names a non-existent symbol.** Line 290 lists "`config.Load`,
  `config.LoadCity`, `config.LoadWithIncludes`," but `config.LoadCity` does not
  exist; the real loaders are `config.Load` (`config.go:3777`),
  `config.LoadWithIncludes`, and `config.LoadWithIncludesOptions`
  (`compose.go:108,113`). Correct the target list.

**Missing evidence:**
- The `internal/config` provenance additions (per-resolved-layer embedded source
  id + pack/manifest digest) that turn a name+path edge into a trusted-identity
  edge, plus a Current-System citation of the existing `config.Provenance` as the
  thing being extended.
- Whether the Core integrity/participation gates are exempt from `--no-strict` and
  from the `splitStrictConfigWarnings`/`reloadWarnings` downgrade path in
  `tryReloadConfig` (`controller.go:926`).
- Which live `config.Load*` sites migrate to `LoadRuntimeCity[NoRefresh]` vs land
  on the partial-read allowlist (at minimum `controller.go:900` reload,
  `internal/dispatch/control.go:836` in-scope, and the four
  `internal/doctor/checks.go` reads allowlisted).
- Whether `MaterializeRequiredPacks` is atomic (temp-tree → rename) and
  concurrency-safe (two `gc` processes / controller + CLI), and how a partially
  materialized pack is *classified* (partial vs corrupt vs stale) when Gate 1's
  digest check rejects it.
- Whether the loader scanner covers generated API/client code.

**Required changes:**
- State explicitly that required-Core/provider file-set integrity and participation
  gates are unconditionally fatal and NOT disableable by `--no-strict` (only
  composition-warning strictness is), and that `tryReloadConfig` replaces
  `cityConfigIncludesWithBuiltinPacks` include assembly with
  `LoadRuntimeCityNoRefresh`.
- Name the `internal/config` provenance additions that make the participation edge
  a typed trusted-identity edge, cite `config.Provenance` in Current System, and
  require the `GateBinding` test to prove that both a same-named user import and a
  byte-identical copied tree at a different path fail Gate 2.
- Define the scanner scope as deny-by-default over all production Go (`cmd/gc` +
  every non-test `internal/` file), with `config.Load*` failing CI unless an
  allowlist row proves a partial read — explicitly allowlist the `internal/doctor`
  bootstrap reads and put `internal/dispatch` in scope. Drop "behavior-driving" as
  the selector or make it a closed, test-asserted list guarded against new
  unclassified production packages. Replace the `worker_boundary_import_test.go`
  citation with a concrete `go/packages`+`go/types` approach, add negative fixtures
  for aliased/wrapped/function-value/generated loader calls, and state whether
  generated code is in scope.
- Specify concurrency safety + atomicity for runtime `MaterializeRequiredPacks`
  (process-unique temp-dir → atomic rename, or shared advisory lock) so Gate 1 can
  never observe a partially materialized required pack written by a concurrent
  process.
- Make the readiness gate non-optional in the API shape: either have the snapshot
  withhold a behavior-usable `Config` unless `ready` (behavior accessors return
  `(Config, error)` that errors in `read_only_degraded`/`blocked`), or add a
  scanner/test proving every enumerated behavior-changing entry point calls
  `RequireReady(op)`.
- Give the config-loader bypass allowlist the same row rigor as the role-neutrality
  allowlist (`owner`, `expiry`, negative fixture), not just `file/function/call
  kind/fields/reason` (line 293).
- Correct the scanner target list (`config.LoadCity` is not a real symbol).

**Questions:**
- Does the live controller reload use `LoadRuntimeCityNoRefresh` (fail-closed, no
  silent Core regeneration) rather than the repair-capable `LoadRuntimeCity`? A
  reload that silently regenerates Core from the embedded manifest would mask an
  operator-edited or partially materialized Core instead of degrading.
- In `read_only_degraded`/`blocked` mode, exactly which reads are served and from
  what source? The plan permits reads of last-known-good without bounding them to
  status-only, so behavior-content reads (formula bodies, prompt/order text,
  dispatch state) risk being served from unvalidated config as authoritative.
- Does Gate 2's "import edge proving participation" require a single authoritative
  edge, or can a required pack participate via multiple edges (direct + transitive)
  and still pass — is multi-edge participation a collision or a provenance-preserving
  dedupe?

**Schema conformance (implementation-plan.schema.md, non-empty):** Conforms. Front
matter carries all required keys with `phase: implementation-plan`,
`requirements_file` pointing at the approved requirements, `status: draft`, and no
`design_file`. The seven required body sections are present and in order. Summary is
within the 2-5 sentence bound; Current System is source-grounded; Proposed
Implementation names concrete files/APIs/boundaries; Testing names exact `go test`
proof commands; Open Questions is "None" (with a trailing clarifying paragraph that
slightly stretches, but does not violate, the bare-`None` contract). Inline
`<!-- REVIEW: ... -->` comments (e.g. lines 108, 222, 314, 345) are provenance
notes, not "Attempt N Review Response" sections, so no design-iteration rule is
violated, though strictly that apply-residue belongs in the workflow artifact
directory. The risks above are on content in my lane, not schema shape.
