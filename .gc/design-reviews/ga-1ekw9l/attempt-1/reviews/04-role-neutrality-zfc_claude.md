# Anand Krishnaswamy - Claude

**Verdict:** block

I reviewed only my lane: zero hardcoded roles in Go and assets, the symbolic
`maintenance_worker` binding, SDK self-sufficiency, and ZFC judgment
containment. I also schema-checked the artifact against
`implementation-plan.schema.md`. The plan's *binding architecture* is genuinely
strong and ZFC-correct — symbolic bindings with city > pack > env precedence,
an explicit "no Go fallback may substitute a concrete role name," and tested
controller-only operation when the worker is renamed or omitted. The block is
narrow: AC8's role-neutrality scan is a *binding evidence contract*, and this
plan defers the three policy decisions that determine whether that scanner
actually prevents the headline red flag. As written, an implementer must infer
how the scanner separates an inert `dog` default from `dog` routing, whether
Core keeps role-token-bearing `mol-dog-*` asset names, and whether the API
`crew` classification is in scope. Those inferences are the substance of my
lane, so the plan is not yet ready to produce correct AC8 evidence.

**Top strengths:**
- **Binding indirection is concrete and Go-fallback-free.** The "Role Neutrality
  And Configurable Bindings" section (L432–446) adds `[gc.bindings.*]`,
  `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and
  `GC_CORE_MAINTENANCE_WORKER`; precedence is a pure data merge (city overrides
  pack; env supplies a default only when neither names a binding); and "No Go
  fallback may substitute `mayor`, `deacon`, `dog`, or another concrete role
  name." This is the right ZFC shape — transport, not judgment — and directly
  answers my "default binding encodes a Go judgment call" red flag.
- **SDK self-sufficiency is tested, not asserted.** L435–437 and the AC9 matrix
  row (L714) require binding tests for default, renamed, omitted, disabled, and
  no-executor cities proving Core-only cities load and controller-owned SDK
  operations still run with no dependency on a user agent entry. The loader's
  `Mode`/`RequireReady(op)` gate (L251–286) keys "may this operation run" to a
  mechanical participation check, not a role heuristic — ZFC-clean.
- **Scanner coverage explicitly includes the surfaces the red flag names.** The
  role-surface manifest spans "generated command text, API classifications,
  dashboard/OpenAPI generated references, tmux theme helpers, default
  scaffolding, warmup mail defaults, prompt fallbacks, formulas, overlays,
  metadata, tests, docs, and public Gastown companion files" (L412–416), and the
  active-root enumeration mandate (L418–424) pushes discovery through
  `internal/packsource.ActiveRootsFor(kind)` instead of raw globs — a real ZFC
  containment of "where behavior comes from."

**Critical risks:**

- **[Major] The role-neutrality scanner's inert-`dog`-vs-routing semantics are
  unspecified, leaving the headline red flag unenforced.** AC8 (requirements
  L114) draws a sharp line: literal `dog` is allowed *only* as "exact inert
  configured-default pack data keys" and is prohibited as routes, notification
  targets, formula bindings, prompt defaults, overlays, generated defaults,
  environment override defaults, or Go fallbacks. The plan restates the
  permission ("`dog` is allowed only as Core's default configurable maintenance
  worker in pack configuration", L429–430) but never says the scanner is
  *field/position-aware* — i.e., that `default_agent = "dog"` at the named
  binding key passes while `pool = "dog"`, `gc.routed_to = "dog"`, a mail/nudge
  recipient `"dog"`, a warmup default, or `{{dog}}` in a prompt must fail. The
  allowlist row fields it does name — "owner, reason, expiry, source path, token
  kind, and a negative fixture" (L593–596) — are close but "token kind" is
  ambiguous between lexical kind and semantic field position. Without explicit
  position-awareness, the scanner can be implemented as a blanket `dog`
  allowlist that lets routing `dog` through — exactly my red flag
  (`gc.routed_to`, mail, nudge, warmup, theme still hardcodes `dog`).

- **[Major] Core-retained `mol-dog-*` formula/order names have no stated
  disposition.** The plan moves `mol-polecat-*` and `mol-refinery-patrol` to
  Gastown and replaces formula-name heuristics with declared metadata (L448–456),
  but is silent on `mol-dog-jsonl`/`mol-dog-reaper`, which the design lineage
  keeps in Core. Those are Core-owned asset *identifiers* containing a role
  token, scanned under "TOML" and "moved asset names." A formula name is not a
  "pack data key," so under AC8's literal wording it is not obviously covered by
  the inert-default exception. The plan must decide: keep the names behind
  explicit, expiring allowlist rows with negative fixtures, or rename to
  binding-neutral names with order-skip aliases. As written this is an unstated
  contradiction between "`dog` is the Core maintenance worker" and "no literal
  role tokens in active Core assets," and decomposition cannot resolve it without
  inventing policy.

- **[Major] The API `crew`/`agent_kind` classification is not named as a concrete
  de-roling surface.** My red flags and AC8 call out role branching on "API
  route" / API classifications. The scanner *coverage* claims "API
  classifications" (L413), but the concrete "Move or rewrite active role-specific
  surfaces" list (L448–457) names branch pruning, `mol-polecat-*`,
  `mol-shutdown-dance`, `mol-review-quorum`, overlays, Dog prompt fragments,
  review checks, tmux theme APIs, and formula-name heuristics — and omits the
  Go-side `crew`/`agent_kind` API/session family classification, the one concrete
  Go API role-branch in the requirements lineage. If the manifest flags `crew`
  but no slice is tasked to neutralize it (with same-slice OpenAPI/dashboard
  regeneration), Slice 3's role scan cannot go green.

- **[Minor] Binding optionality/requiredness is described as if Go-known, not
  data-declared.** L443–445 says "Missing optional bindings skip… Missing
  required provider-pack escalation bindings fail," but does not restate AC9's
  requirement that required/optional keys are *declared in pack data* and
  enforced generically. If Go decides per-binding-name which bindings are
  optional vs required, that is a ZFC judgment call. The plan should state the
  optional/required flag is read from the binding declaration and resolved
  generically.

**Missing evidence:**

- No current-system inventory of the raw `fs.WalkDir`/`Glob`/string-prefix
  enumeration call sites that the `ActiveRootsFor` mandate (L418–424) must
  replace. The schema requires current-system claims to cite concrete files; the
  Current System section names no enumeration sites, yet Proposed Implementation
  bans them everywhere — so the scope (and false-positive risk for non-pack
  walks) is unsized.
- No statement of which exact TOML key path(s) constitute the permitted inert
  `dog` default, which is the anchor the position-aware scanner needs.
- The inline `<!-- REVIEW: added per … -->` markers (e.g., L108, L222, L409,
  L460, L690) are review-process provenance leaking into the artifact; the schema
  says attempt notes belong in the workflow artifact directory, not in
  `implementation-plan.md`. Minor schema-hygiene, not a structural failure (the
  required front matter, section set, and order all conform).

**Required changes:**

1. Specify that the AC8 role-neutrality scanner is field/position-aware: name the
   exact binding-default key path(s) where literal `dog` is permitted, and
   require negative fixtures proving `dog` fails in route/pool/recipient/
   warmup/prompt/overlay/env-default/Go-fallback positions.
2. State the disposition of Core `mol-dog-*` formula/order names (allowlisted with
   owner+expiry+negative fixture, or renamed with order-skip aliases), so a
   known role-token asset class is not left to implementer inference.
3. Name the API/dashboard `crew`/`agent_kind` classification as an explicit
   de-roling surface with a neutral-metadata replacement and same-slice
   OpenAPI/dashboard regeneration.
4. State that binding optionality/requiredness is declared in pack data and
   resolved generically, with no per-binding-name Go branch.
5. Add a current-system call-site inventory for raw pack-root enumeration so the
   `ActiveRootsFor` containment can be sized and won't become an allowlist sink.

**Questions:**

- What is the exact key path of the permitted inert `dog` default, and does the
  scanner treat any `dog` outside that path as a violation by default?
- Are `mol-dog-jsonl`/`mol-dog-reaper` staying in Core under their current names,
  or being renamed? Either answer needs an allowlist or alias plan.
- Is `crew`/`agent_kind` API classification in scope for Slice 3 role
  neutralization, and if so what is the neutral replacement contract?
