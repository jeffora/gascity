# Ingrid Kovac — DeepSeek V4 Flash (Independent)

**Verdict:** block

Lane: zero hardcoded roles, Core role neutrality, `dog` exception containment,
SDK self-sufficiency, Go-source migration guard coverage. Reviewed against the
revised `design.md` (updated 2026-06-05T20:30, including the
`<!-- REVIEW: added per blocker-core-role-neutrality -->` contract table and
the `<!-- REVIEW: added per blocker-rollout-and-test-slicing -->` invariants)
and the live tree at the `rig_root`.

The revised design is meaningfully stronger than the original: the Core
maintenance/notification contract table (lines 204-211) and the role-token
scanner specification (lines 220-224) address most of the prior round's surface
gaps on the pack-asset side. The rollout invariants and the explicit
staging-with-gates structure are also improvements. I block on three problems
the revision does not resolve and one new problem it introduces.

**Top strengths:**

- The contract table at lines 197-218 is the right mechanism. "Mail uses
  configured recipients," "nudge/route resolves through configured agent names,"
  "scripts accept targets from env/formula vars," "Go code may not contain
  mayor/deacon/witness/refinery/polecat/boot/crew/gastown as a control
  decision" — these are enforceable, testable, and structural. The design now has
  a genuine specification instead of a wish.

- SDK self-sufficiency gets its proof obligation: "Tests must prove a Core-only
  city can still load, run normal SDK infrastructure, and evaluate non-agent
  controller operations when the maintenance agent is renamed or omitted"
  (lines 213-218). This is the right test shape, even though the proof is not
  yet delivered.

- `mol-review-quorum.toml` is genuinely role-neutral: every dispatch target is
  a required formula variable. This proves the pattern works and should be
  cited as the reference implementation for all other Core formulas.

- The review-gated migration invariants (lines 39-47) and the seven-slice
  rollout with explicit gates per slice are a significant structural improvement.
  They prevent the previous risk of batching incompatible changes.

**Critical risks:**

- **[Blocker] The role-token scanner is defined only over Core pack assets, but
  the most behaviorally dangerous role-specific logic lives in `internal/` Go
  code, and the design's own Cross-Pack Ownership row for Go theme/tmux APIs
  (line 570) gates source-tree deletion on resolving those references — without
  specifying the resolution.** Verified in the live tree:

  - `internal/runtime/tmux/tmux.go:80`: `roleEmoji` maps `mayor`, `deacon`,
    `witness`, `refinery`, `crew`, `polecat` to display emoji, aliased at
    line 2854 as `roleIcons`. This is role-keyed control logic in production
    Go — the exact class the ZFC principle prohibits.
  - `internal/runtime/tmux/theme.go:33-46`: `MayorTheme()`, `DeaconTheme()`,
    `DogTheme()` are Go functions named after specific Gastown roles. The design
    says `dog` must not be "a Go special case" (line 167), yet `DogTheme()` is
    precisely that.
  - `internal/config/config.go:4458`: `DefaultCity` and `WizardCity` both
    hard-code `{Name: "mayor", PromptTemplate: "prompts/mayor.md"}` and
    `{Template: "mayor", Mode: "always"}`.
  - `cmd/gc/cmd_start_warmup.go:33`: `defaultWarmupMailTo = "mayor"` is a
    hardcoded role name as a Go constant.
  - `cmd/gc/cmd_prompt.go:631-632`: falls back to `prompts/mayor.md` when no
    role-specific prompt exists — a Gastown role baked into the CLI fallback
    path.
  - `internal/sling/sling.go:917`: `strings.HasPrefix(formulaName,
    "mol-polecat-")` and line 923 `formulaName == "mol-refinery-patrol"` are
    Gastown-formula-name heuristics in the Go dispatch path. These are
    framework cognition: Go code making decisions based on Gastown workflow
    names.

  The design acknowledges this category at line 570: "Hardcoded role-theme/tmux
  APIs ... Source scanner and config-driven theme test." But it does not
  specify: (a) what replaces the role-named theme functions, (b) what replaces
  `roleEmoji`/`roleIcons`, (c) what replaces the `mayor` default in
  `cmd_start_warmup.go` and `cmd_prompt.go`, (d) what replaces the
  `mol-polecat-*`/`mol-refinery-patrol` heuristics in `sling.go`, or (e) what
  the Go "Source scanner" covers (root, prohibited tokens, allowlist policy).

  Lines 555-557 state these decisions "must be resolved before deleting source
  trees or moving the public Gastown pin," which makes them blocking — but no
  slice in the rollout addresses them. Slice 7 (source deletion) cannot proceed
  until they are resolved, and the design provides no resolution. This is a
  self-imposed deadlock.

- **[Blocker] The existing role-name guard is deleted by the migration with no
  replacement covering Go.** The only role-name test in the tree is
  `examples/gastown/tmux_theme_script_test.go:148`
  `TestTmuxThemeScriptHasNoHardcodedRoleNames`, which checks
  `examples/gastown/packs/gastown/assets/scripts/tmux-theme.sh`. The design
  removes `examples/gastown/packs/gastown` (lines 137, 509), so this test and
  its target leave the repo. The new Core-asset scanner does not cover Go. After
  the migration, the repo has *less* role-name enforcement than today, while
  the Go violations above remain. This is a net regression in the very
  invariant the design claims to strengthen.

- **[Blocker] `dog` is already a Go special case, directly contradicting the
  design's stated invariant.** Line 167 says `dog` is "allowed only as pack
  configuration, not as a Go special case or SDK infrastructure dependency."
  Verified contradictions in the live tree:

  - `theme.go:43-47`: `DogTheme()` is a Go function returning a theme for the
    `dog` role. The design requires either removing it or documenting an
    exemption; it does neither.
  - `embed_builtin_packs.go:237`: `requiredBuiltinPackNames` hard-codes
    `"maintenance"` (which provides the `dog` agent definition). After the
    migration folds Maintenance into Core, this line must change to only
    `"core"`, but the design's line 236 still reads
    `required := []string{"core", "maintenance"}` in the current tree and the
    design does not call out this specific line change as a slice-5 gate.
  - The `dog` agent definition lives in
    `examples/gastown/packs/maintenance/agents/dog/agent.toml` — not in Core.
    If Core orders (`mol-dog-jsonl`, `mol-dog-reaper`) route to `pool = "dog"`,
    they reference an agent that does not exist in Core configuration. The
    design must state explicitly: does `internal/packs/core/` contain an
    `agents/dog/agent.toml` with a minimal pool definition, or is `dog` always
    user-supplied? If Core ships `dog` as a default pack configuration, the
    design needs a test proving Core orders dispatch correctly when that agent
    is renamed. If `dog` is user-supplied, every Core order with `pool = "dog"`
    violates SDK self-sufficiency.

- **[Major] The Core asset scanner specification is still an outcome, not a
  scanner contract, despite the revision's improvements.** Lines 220-224 say
  "Add a role-token scanner over Core TOML, shell, Markdown, templates,
  overlays, skills, embedded command text, generated manifests, and Core tests."
  This lists file types but does not specify: the scanner root directory, the
  exact TOML fields to check (step descriptions, `[vars].*.default`, metadata
  values, order execution fields), the prohibited token set (just role names,
  or also `gastown` as a pack name reference?), the `dog` allowlist policy
  (token-by-field-and-path vs. whole-file), or the expected negative fixtures.
  The prior reviews converge on this: the existing contamination patterns prove
  that a file-level string scan misses `default = "deacon"` inside `[vars]`,
  `<rig>/witness` inside step description strings, and `gc mail send mayor`
  inside prompt prose. The design must specify a token-, path-, and
  field-scoped scanner contract to close this gap.

**Cross-review convergence analysis:**

All four reviewers (Claude, Codex, Gemini/DeepSeek, and this review) block on
the same root cause: the Go source surface is unguarded while the design claims
role neutrality. The specific Go hits vary in enumeration but converge on
three categories: (1) role-keyed data structures (`roleEmoji`, `roleIcons`,
named theme functions), (2) role-named string constants (`defaultWarmupMailTo`,
`DefaultCity`/`WizardCity`/`GastownCity` agent definitions, `cmd_prompt.go`
fallback), and (3) role-derived dispatch heuristics
(`mol-polecat-*`/`mol-refinery-patrol` prefix matching in `sling.go`).

The disagreement is on scope: Claude and Codex want a production Go role-token
guard for the migration surface; Gemini argues the Go surface should be
explicitly out of scope but documented. My assessment: both are needed. A
production guard for Go files touched by this migration prevents regression,
and a documented exclusion list for untouched legacy Go prevents overreading
the guard's meaning. The design currently does neither.

The `dog` ownership ambiguity is also convergent: all reviewers note that
Core orders reference a `dog` pool whose agent definition lives in the
Maintenance pack, and the design does not resolve whether Core ships that
definition or expects the user to supply it. Gemini's Option (b) — make `dog`
entirely user-supplied and remove `pool = "dog"` from Core orders — is the
ZFC-clean solution, but it requires Core maintenance orders to use the
configurable maintenance-agent name instead of a literal `dog`.

Gemini's observation that `mol-review-quorum` is the role-neutral reference
pattern is correct and should be elevated to an explicit design decision. All
Core dispatch should follow its variable-based pattern.

**Missing evidence (consolidated):**

- A Go migration table: for each Go role literal or heuristic listed above,
  the slice that resolves it, the replacement mechanism, and the test that
  proves the replacement works. Without this, line 570's "Source scanner and
  config-driven theme test" is an unscoped promise, not a plan.

- A concrete `dog` ownership decision: does Core ship
  `internal/packs/core/agents/dog/agent.toml`, or does every Core order
  replace `pool = "dog"` with a formula variable? The design must choose one
  and prove it against the SDK self-sufficiency test.

- A scanner contract specification: root (`internal/packs/core`), closed set
  of file extensions, TOML field enumeration (step descriptions, `[vars]`
  defaults, metadata, order exec fields, template fragments), prohibited token
  set, `dog` allowlist format (token + path + field, not whole-file), and
  negative fixtures per category.

- A replacement for `TestTmuxThemeScriptHasNoHardcodedRoleNames` that covers
  Go source in addition to pack assets, or an explicit scope statement that the
  Go guard is a separate migration and the asset guard alone does not claim
  end-to-end role neutrality.

- A `requiredBuiltinPackNames` transition test: changing from `["core",
  "maintenance"]` to `["core"]` plus verifying that an existing
  `.gc/system/packs/maintenance/` directory is ignored without fatal errors.

**Required changes:**

1. **Resolve the Go role surface explicitly.** For each Go violation, state
   whether it is in-scope for this migration or deferred. If in-scope (as
   line 555-557 implies by gating source deletion on it): add a rollout slice
   for Go de-roling, specify the replacement mechanism (config-driven theming
  for `MayorTheme`/`DeaconTheme`/`DogTheme`, lookup table for `roleEmoji`,
  configurable default for `defaultWarmupMailTo`, formula-variable dispatch
  for `sling.go` prefix heuristics, config-driven agent template for
  `DefaultCity`/`WizardCity`), and specify the Go scanner root, token set, and
  allowlist policy with the same rigor as the Core-asset scanner. If deferred:
  remove line 570 from the blocking decision table and add an explicit exclusion
  note that a passing asset guard does not imply end-to-end Go role neutrality.

2. **Add `DogTheme()` (and any other `dog` Go reference) to the dog-containment
   analysis.** Either remove `DogTheme()` in favor of config-driven theming or
   document it as a deliberate exemption so line 167's "not a Go special case"
   invariant is actually true.

3. **Replace the existing role-name guard.** Either relocate
   `TestTmuxThemeScriptHasNoHardcodedRoleNames` to a Gas City package and
   extend it to cover Go source, or add a separate Go role-literal guard and
   explicitly scope the asset-only guard so that "guard passes" is not read as
   "Core is role-neutral end to end."

4. **Specify the `dog` ownership contract.** Choose: (a) Core ships a minimal
   `agents/dog/agent.toml` that users can override/rename/disable, with a test
   proving renamed/omitted operation; or (b) Core orders use a formula variable
   for the maintenance agent pool name, and `dog` is purely a Gastown convention.
   Do not leave Core orders routing to a literal `dog` pool while `dog` is
   defined in a pack that the design is removing from the embedded set.

5. **Convert the role-token scanner from an outcome to a contract.** Specify the
   root, file types, TOML field coverage, prohibited tokens, `dog` allowlist
  format (token + path + field), and required negative fixtures. Cite
  `mol-review-quorum.toml` as the reference pattern for role-neutral Core
  dispatch.

6. **Add a `requiredBuiltinPackNames` transition gate in slice 5.** The current
   code hard-codes `[]string{"core", "maintenance"}`; the migration must change
   it to `[]string{"core"}` and verify that cities with stale
   `.gc/system/packs/maintenance/` on disk still load correctly.

7. **Clarify rollout ordering between public Gastown publication, embedded
   registry changes, and source-tree deletion.** State explicitly that the
   public Gastown replacement must be published and pinned before the embedded
   registry stops referencing Maintenance/Gastown, and that source directories
   are removed only after both are complete.

**Questions:**

- Is Go de-roling (`theme.go`, `tmux.go`, `config.go`, `sling.go`,
  `cmd_start_warmup.go`, `cmd_prompt.go`) in-scope for this migration or a
  separate effort? The design both gates source deletion on it (lines 555-557)
  and provides no mechanism, which is a self-imposed deadlock.

- What replaces `roleEmoji`/`roleIcons`? The obvious answer is config-driven
  theming (agent definitions include a theme key), but the design does not
  state this. If agent configs carry `theme: {name, bg, fg}` entries, the Go
  map becomes unnecessary. If not, what is the replacement?

- Should `DefaultCity`/`WizardCity`/`GastownCity` continue to hard-code a
  `mayor` agent, or should the default agent come from Core pack configuration?
  This is the same question Codex raised, and the design does not answer it.

- Is `mol-review-quorum.toml` the intended reference pattern for all Core
  dispatch? If so, the design should state this explicitly and require all
  other Core formulas to follow its variable-based dispatch pattern.
