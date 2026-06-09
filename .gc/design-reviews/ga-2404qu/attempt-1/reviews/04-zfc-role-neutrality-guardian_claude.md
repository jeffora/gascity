# Ingrid Kovac - Claude

**Verdict:** block

Reviewed `.gc/design-review-inputs/core-gastown-pack-migration/design.md`
(`updated_at: 2026-06-09T08:40:42Z`, latest section Attempt 17) against
`requirements.md`. Lane only: zero hardcoded roles, Core role neutrality, `dog`
exception containment, SDK self-sufficiency. Every claim re-grounded against the
live tree at `/data/projects/gascity`.

The role-neutrality machinery is, after 17 attempts, genuinely strong — the
`dog` red flags are fully contained and the scanner is finally enforceable. The
block is narrow and surgical: two forbidden-token Go surfaces the scanner
*names* but the design never *dispositions*, one canonical implementer-facing
section that lags the binding contract it depends on, and one live internal
contradiction. None is a flaw in the role model; each is an unresolved gating
decision the design defers to the generator when its own rules say it must
decide first.

**Top strengths:**

- The scanner is sound where it counts. It "splits snake_case, kebab-case, dot
  paths, camelCase, PascalCase ... comments" and rejects "sub-identifier
  surfaces such as `DogTheme`, role icon maps, theme functions, warmup defaults,
  prompt fallbacks, `crew` API/dashboard vocabulary, `gastown` template wiring
  ... and dispatch formula-name heuristics" (2487-2494). I verified that net
  reaches the real live offenders a delimiter-only scan would miss:
  `MayorTheme`/`DeaconTheme` (`internal/runtime/tmux/theme.go:33-46`), the
  `roleEmoji` map (`internal/runtime/tmux/tmux.go:80`, aliased `roleIcons` at
  `:2823`), `defaultWarmupMailTo = "mayor"` (`cmd/gc/cmd_start_warmup.go:33`),
  and `SlingFormulaUsesBaseBranch` (`internal/sling/sling.go:888`).

- `dog` containment and SDK self-sufficiency are concrete and testable, not
  aspirational: `[gc.bindings.maintenance_worker]` with
  `optional_for_controller` (1799-1803); the empty-value contract (omission must
  not prevent Core load, config, controller-owned SDK ops, beads, events,
  formulas, sessions; worker-bound work degrades with
  `core.maintenance_worker.omitted`, 1812-1817); symbolic
  `target_binding`/`gc.run_target_binding`/`GC_CORE_MAINTENANCE_WORKER`
  (1819-1827); `[[patches.agent]] name = "dog"` reduced to compatibility-only
  (1838-1845). I verified the precondition holds today: **no production Go
  branches on a `dog` agent** — every non-test `dog` hit is a comment/example
  or the display-only `DogTheme`.

- Allowlist growth — my top red flag — is structurally bounded: rows carry
  owner + expiry + proof test + a release gate that fails after expiry
  (1925-1927, 2589); Core-owned rows *fail* on hardcoded Gastown roles (2498);
  compatibility rows are confined to fixtures, docs, old-pack diagnostics, and
  `maintenance_worker` data.

**Critical risks:**

- [Major] The `gastown` init/config wiring is named by the scanner but has no
  disposition, and the design never decides the underlying policy. Verified in
  production Go: `cmd/gc/cmd_init.go:134` (`case "2", "gastown"`) and `:975`
  (`wiz.configName == "gastown"`); `internal/config/config.go:3706-3738`
  (`GastownCity` hardcodes the `"gastown"` import and
  `DefaultRigImportOrder: []string{"gastown"}`);
  `internal/config/public_packs.go:7,11`
  (`PublicGastownPackSource`/`PublicGastownPackVersion`);
  `cmd/gc/import_state_doctor_check.go:141` (`case "gastown"`). The scanner
  rejects "`gastown` template wiring ... unless a row allows them" (2492-2493),
  yet the Go disposition table (2506-2514) supplies no such row — so by the
  design's own deferral rule (2516-2518) every slice touching these files is
  blocked. This is a gating decision the design must make, not defer to the
  generator: is the SDK permitted to hard-encode the canonical public Gastown
  identity and the `gastown` template token (the requirements *do* endorse
  `gc init --template gastown` importing the public pack explicitly), or must
  template→source resolution become data/registry-driven? Either answer is
  defensible; the absence of one blocks implementation.

- [Major] The canonical "Core Maintenance And Notification Contract" (2831-2858)
  — the section an implementer actually reads — never names the binding it
  relies on. It says only "resolve the configured maintenance agent name rather
  than a Go constant" (2851-2852), which a literal `pool = "dog"` in a Core
  formula/order satisfies, and the scanner *allows* `dog` in Core. The binding
  (`target_binding = "core.maintenance_worker"`,
  `gc.run_target_binding`) lives only in the Attempt-11/14/17 contracts, and the
  precedence rule "prose is advisory when it conflicts with a generated row"
  sits at 2603, far from Proposed Design. Result: a plausible implementation
  hardcodes `dog` pools/routes in Core assets that pass the scanner and fail
  only if the renamed-worker test is written at execution level.

- [Major] Live internal contradiction on the `agent_kind`/`crew` wire surface.
  The role-surface table de-roles it: "API/OpenAPI/dashboard `crew` vocabulary →
  replace with neutral session grouping or config-supplied labels; **regenerate
  schema/types in the same slice**" (2509). But Testing still states "Run
  `make dashboard-check` only if API, dashboard, or generated schema files are
  touched. **This migration should not require dashboard changes**" (3369-3370).
  Verified that `agent_kind` is a live wire field
  (`internal/api/handler_sessions.go:46,49`, set at `:118` from
  `classifyAgentKind`) baked into `internal/api/openapi.json:6166`,
  `docs/schema/openapi.json:6166`, and generated TS
  (`cmd/gc/dashboard/web/src/generated/{schema.d.ts:3861,types.gen.ts:2569}`) —
  exactly the CI-enforced paths AGENTS.md/CLAUDE.md say require
  `make dashboard-check` + OpenAPI/TS regeneration. Both claims cannot hold.
  `classifyAgentKind` (`internal/api/handler_agents.go:555-569`) is already
  structurally neutral ("The classifier never inspects role names"; it keys on
  the `crew` dir segment), so a defensible resolution is to record `crew` as an
  allowed neutral structural convention and leave the wire alone — but the
  design must pick one. (Carried from the prior iteration; Attempt 17 did not
  resolve it.)

- [Minor] SDK self-sufficiency is asserted without the classification that would
  prove it. Core absorbs orphan-sweep, wisp-compact, order-tracking, gate-sweep,
  spawn-storm, reaper, jsonl (2795-2798). The design states "controller-owned
  SDK operations still work when the worker is omitted" (211, 1813-1815, 2844)
  but never enumerates which Core orders survive worker omission versus go
  inert. The binding mechanism is sound, but the omitted-worker fixture (2476)
  cannot prove the AGENTS.md invariant ("removing a `[[agent]]` entry must not
  break an SDK feature") without that list — the test author silently draws the
  boundary.

- [Minor] The tmux remedy "config-driven display metadata" (2508) is right, but
  imprecise on disposition. Verified: `MayorTheme`/`DeaconTheme`/`DogTheme`
  (`theme.go:33-46`) have **no non-test callers** — they are dead code to
  *delete*, not "wrap." The behavior-bearing surface is `roleEmoji`
  (`tmux.go:80`, consumed as `roleIcons` at `:2823` for status display), keyed
  on `mayor/deacon/witness/refinery/crew/polecat`; that is what needs the
  config-supplied replacement.

**Missing evidence:**

- No behavior-manifest pilot row (the matrix at 2294-2306 omits it) preserving
  the **Core** formula `mol-scoped-work`'s base_branch behavior when the sling
  role-name heuristic is replaced by formula metadata.
  `SlingFormulaUsesBaseBranch` (`internal/sling/sling.go:888`) conflates
  `mol-scoped-work` with `mol-polecat-*`, and `:894`
  (`SlingFormulaUsesTargetBranch`) keys on `mol-refinery-patrol`; the neutral
  rewrite must keep `mol-scoped-work`'s behavior in Core, not drop it or push it
  to Gastown when the role-named formulas leave the binary.
- The per-order controller-owned-vs-worker-bound classification, plus proof the
  worker-omitted (inert) set contains zero load-bearing SDK infrastructure.
- The disposition (accepted constant vs data-driven) for the `gastown`
  init/config surfaces above.
- Whether `agent_kind`/`crew` ever drives a control decision or is strictly a
  display label (only the latter is safe to neutralize by renaming, or to keep
  as a neutral structural convention).

**Required changes:**

- Add a `role-surface.generated.yaml` disposition row and a prose policy for the
  `gastown` init/config wiring (`cmd_init.go`, `config.go`, `public_packs.go`,
  `import_state_doctor_check.go`): either accepted as the SDK's sanctioned
  knowledge of the canonical public pack (with a documented structural
  carve-out, since `dog` already has one) or data/registry-driven template
  resolution. This gates whether those files are rewritten; do not defer it to
  the generator.
- In "Core Maintenance And Notification Contract" (2831-2858), name
  `[gc.bindings.maintenance_worker]` and
  `target_binding`/`gc.run_target_binding`/`GC_CORE_MAINTENANCE_WORKER` as the
  REQUIRED resolution path for every worker-bound Core order/formula/script,
  state that a literal worker target (e.g. `pool = "dog"`) in any Core asset is a
  defect, and restate the "prose is advisory" precedence (2603) in Proposed
  Design so the canonical section stops under-specifying the binding.
- Resolve the `agent_kind`/`crew` contradiction explicitly: either keep
  `classifyAgentKind` and record `crew` as an allowed neutral structural
  convention (no wire change, "no dashboard changes" stays true), OR commit to
  the wire change, delete the "should not require dashboard changes" claim
  (3370), and add `make dashboard-check` + OpenAPI/TS regeneration to the owning
  slice's gates. Not both.
- Add the controller-owned-vs-worker-bound order classification; require the
  omitted-worker test to assert the full controller-owned set still functions
  AND each worker-bound order goes cleanly inert (no controller stall, no
  per-tick `core.maintenance_worker.omitted` diagnostic spam).
- State the per-surface tmux disposition: delete the dead
  `MayorTheme`/`DeaconTheme`/`DogTheme`; specify the config-supplied display
  metadata that replaces the `roleEmoji`/`roleIcons` role→icon map.
- Add a behavior-manifest row preserving `mol-scoped-work`'s base_branch
  behavior in Core when the sling role-name heuristic is replaced.

**Questions:**

- Is the SDK permitted to hard-encode the canonical public Gastown identity
  (`PublicGastownPackSource`/`Version`, `gc init --template gastown`,
  `GastownCity`), or must template→source resolution become data-driven to
  satisfy ZERO hardcoded roles? This is the crux of the block.
- Is `crew`/`agent_kind` a forbidden role name or an acceptable neutral
  structural grouping, and does the answer change the dashboard scope?
- When the maintenance worker is omitted in a Core-only city, is losing
  orphan-sweep / wisp-compaction / order-tracking an accepted operator tradeoff,
  or are any of those SDK infrastructure that must run controller-side
  regardless of the configured worker?
