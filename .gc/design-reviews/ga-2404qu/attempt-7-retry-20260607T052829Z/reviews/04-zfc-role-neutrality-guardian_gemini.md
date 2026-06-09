# Ingrid Kovac — DeepSeek V4 Flash (Independent)

**Verdict:** block

Lane: zero hardcoded roles, Core role neutrality, `dog` exception containment, SDK self-sufficiency, Go-source migration guard coverage, cross-document consistency, and pattern drift. Reviewed against the revised `design.md` (updated 2026-06-07, including the Core Maintenance & Notification Contract table at lines 727–742 and the review-gated migration invariants at lines 42–58), the `requirements.md`, the live tree at the `gascity` root, and the prior reviewer outputs.

The revised design is materially stronger than the original: the contract table (727–742) and the role-token scanner specification (750–755) are genuine structural controls rather than aspirations. The rollout invariants (42–58) and seven-slice rollout with explicit gates are a significant structural improvement over the original monolithic approach. The behavior-preservation inventory requirement and the `requiredBuiltinPackNames` transition table are also strong additions.

However, I block on three self-consistency problems and one missing mechanism that the revision does not resolve.

---

### Top Strengths

- **Enforceable Notification Contract:** The contract table at lines 727–742 is the right abstraction. It turns "Core assets must not contain role names" from an outcome into a per-operation contract with concrete replacement patterns: mail uses configured recipients, nudge/route resolves through configured agent names, scripts take targets from env/formula vars, and Go code may not contain Gastown role names as control decisions. This is enforceable and testable.
- **Role-Neutral Reference Design:** `mol-review-quorum.toml` is genuinely role-neutral: every dispatch target is a required formula variable. It proves the pattern works and should be cited as the reference implementation for all other Core dispatch.
- **Review-Gated Rollout:** The review-gated invariants (42–58) prevent the prior risk of batching incompatible changes. Each slice must be independently deployable and test-green before downstream work can proceed.
- **Precise Builtin Transitions:** The `requiredBuiltinPackNames` transition table is precise and addresses real production-safety concerns about stale directories and cache entries.

---

### Critical Risks

#### 1. [Blocker] Go Role Surface Deadlock
The design gates source deletion on resolving the Go role surface (1102–1103), but no rollout slice addresses Go de-roling, and the specified scanner does not cover `internal/`. This is a self-imposed deadlock. The Cross-Pack Ownership table at lines 1105–1114 lists "Hardcoded role-theme/tmux APIs" with evidence "Source scanner and config-driven theme test," and lines 1102–1103 state these decisions "must be resolved before deleting source trees or moving the public Gastown pin." This makes Go de-roling a blocking dependency for slice 7 (source deletion). Yet no slice in the rollout (1–7) addresses the Go surface, and the scanner specification at 750–755 covers only Core pack assets, not Go source. The Go violations are concrete, behavior-bearing, and verified in the live tree:

- `internal/runtime/tmux/tmux.go:80`: `roleEmoji` maps six Gastown role names to emoji, consumed at line 2829 via `roleIcons`. This is role-keyed control logic in production Go.
- `internal/runtime/tmux/theme.go:33–46`: `MayorTheme()`, `DeaconTheme()`, `DogTheme()` are role-named Go functions. The design says `dog` must not be "a Go special case" (696), yet `DogTheme()` is exactly that.
- `internal/config/config.go:3670–3691` (or related configuration lines): `DefaultCity` and `WizardCity` hardcode `{Name: "mayor", PromptTemplate: "prompts/mayor.md"}` and `{Template: "mayor", Mode: "always"}`.
- `cmd/gc/cmd_start_warmup.go:33`: `defaultWarmupMailTo = "mayor"` hardcodes a Gastown role name as a Go constant.
- `cmd/gc/cmd_prompt.go:632–633`: falls back to `prompts/mayor.md` when no role-specific prompt exists — a Gastown role baked into the CLI fallback path.
- `internal/sling/sling.go:888–894`: `SlingFormulaUsesBaseBranch` matches `mol-polecat-*` by name prefix and `SlingFormulaUsesTargetBranch` matches `mol-refinery-patrol` exactly. These are Gastown-formula-name heuristics in the Go dispatch path — framework cognition that violates ZFC.
- `internal/api/handler_agents.go:555–569`: `classifyAgentKind` returns `"crew"` via `isCrewDir(a.Dir)`, which checks `dir == "crew" || HasSuffix(dir, "/crew")`. This is an `if role == "crew"` branch for a Gastown role in the API layer.

The design must either:
- **(a)** Add a Go de-roling rollout slice with a replacement mechanism for each violation, a Go scanner contract (root, prohibited tokens, allowlist), and a gate for source deletion, or
- **(b)** Remove line 1114 from the blocking decision table and explicitly state that a passing asset-only guard does not imply end-to-end Gas City role neutrality, with a separate tracking issue for Go de-roling. Today it does neither.

#### 2. [Blocker] Enforcement Regression on Example Deletion
The migration removes the only existing role-name guard (`TestTmuxThemeScriptHasNoHardcodedRoleNames` in `examples/gastown/tmux_theme_script_test.go`) with no replacement covering Go source, creating a net regression in role-neutrality enforcement. This test lives under `examples/gastown/` and checks `examples/gastown/packs/gastown/assets/scripts/tmux-theme.sh` for forbidden role names. When slice 7 deletes the `examples/gastown/` tree, this test leaves with it. The proposed Core asset scanner (750–755) covers pack assets only. The design must specify either:
- **(a)** A relocated pack-asset test plus a Go-source role-literal guard, or
- **(b)** An explicit scope statement that the asset-only guard does not cover Go. Without this, post-migration Gas City has *less* automated role-neutrality enforcement than it has today.

#### 3. [Blocker] Ambiguity in `dog` Ownership
The `dog` ownership question is unresolved, and every answer creates a different obligation the design does not address. The design says `dog` is "a default pack configuration value for the Core maintenance agent" (743) and that Core may define a default `dog` maintenance agent (requirements.md:34). But the current `agents/dog/agent.toml` lives in `examples/gastown/packs/maintenance/agents/dog/agent.toml` — a pack the design retires. Three outcomes are possible:
- **(a) Core ships `internal/packs/core/agents/dog/agent.toml`:** Then Core owns an agent definition, and the design must specify the minimal fields, the rename/disable test, and what happens when a user overrides or removes it. Core orders like `mol-dog-reaper` (currently controller-executed) would dispatch to a Core-configured pool, which is self-consistent.
- **(b) Core orders reference `dog` as a formula variable with no default:** Then every Core formula must require the user to configure a maintenance agent pool, violating SDK self-sufficiency for cities that don't configure one.
- **(c) Core orders take a config-resolved "maintenance agent" indirection, and `dog` is purely a Gastown convention:** Then Core formulas like `mol-dog-reaper` must resolve the maintenance pool from config, and `pool = "dog"` must not appear in any Core order TOML.

The design says "Core maintenance-agent behavior is Core pack configuration" (1073) and "Core may define a default configurable maintenance agent named `dog`" (requirements.md:34), but does not commit to (a), (b), or (c). It also does not address `DogTheme()` in `theme.go`, `defaultWarmupMailTo` in `cmd_start_warmup.go`, or default configurations in `config.go` — all of which are Go references to `dog` or `mayor` that become Go special cases if (a) is chosen without a config-driven replacement.

#### 4. [Major] Incomplete Notification Target Application
The Core notification target contract (727–742) is the right mechanism, but the design does not apply it to every Core asset that currently names a role. Verified in the current tree:
- `internal/bootstrap/packs/core/formulas/mol-polecat-base.toml` routes failure escalation to `<rig>/witness` in three places.
- `mol-polecat-commit.toml` routes push-conflict escalation to `<rig>/witness`.
- `mol-prompt-synth.toml` routes all failure modes to "Mail Witness".
- `pool-worker.md` and `graph-worker.md` both say `gc mail send mayor` for escalation.
- `gc-dispatch/SKILL.md` references mayor/refinery/polecat examples and patrol formula references.

The contract table says these should be "removed or renamed by configuration," but does not specify which replacement pattern applies to each asset. The three patterns have different implications: formula variable with no default (requires config), config-resolved maintenance pool (Core self-sufficient only if `dog` is Core-shipped), or remove escalation from Core and overlay in Gastown (Core fails silently without Gastown). The design must choose one per asset and prove each choice against SDK self-sufficiency.

---

### Other Significant Findings

- **[Major] Scanner Specification is an Outcome, Not a Contract:** Lines 750–755 say "Add a role-token scanner over Core TOML, shell, Markdown, templates, overlays, skills, embedded command text, generated manifests, and Core tests." This lists file types but does not specify: the scanner root directory, the exact TOML fields to check (step descriptions, `[vars].*.default`, metadata values, order execution fields), the prohibited token set (just role names, or also `gastown` as a pack name reference?), the `dog` allowlist format (token-by-field-and-path vs. whole-file), or the expected negative fixtures.
- **[Major] Comment Scanning Leak:** The existing test `TestTmuxThemeScriptHasNoHardcodedRoleNames` deliberately strips comment lines before scanning. Moved Core orders contain role names in comments (`orphan-sweep.toml`: "Replaces deacon patrol step"; `spawn-storm-detect.toml`: "a polecat is..."; `cross-rig-deps.toml`: "Replaces deacon patrol step"). A scanner that follows the existing pattern and strips comments will pass these orders. The design must state whether the scanner covers comments, `description` fields, and metadata strings — and if so, require those order comments scrubbed with behavior preserved in Gastown per the inventory.
- **[Major] Escape-Hatch Allowlist Policy:** The design permits "historical docs and explicit Gastown fixture names" via "a reviewed allowlist" (518) and forbids role names "outside allowed docs/tests", but defines no allowlist format, per-entry justification, owner, or cap. Specify path/token/field/category/reason/owner per entry, with no broad wildcards. Without this, the allowlist becomes the escape hatch that normalizes violations.
- **[Major] Untestable Core Self-Sufficiency:** The design requires a Core-only city to "load, run normal SDK infrastructure, and evaluate non-agent controller operations when the maintenance agent is renamed or omitted" (744-748), but does not list the controller-owned operations that must still work, which agent-executed operations may be disabled or retargeted, or how orders resolve without a literal `dog` pool. Without that list, the proof obligation is untestable.
- **[Major] Crew Hardcoding in API Layer:** `classifyAgentKind` returning `"crew"` in `internal/api/handler_agents.go:555–569` is a Go control path that hardcodes a Gastown role name. This is in the same class as `sling.go`'s formula-name heuristics — Go code making decisions based on Gastown-specific knowledge. None of the prior reviews identified this as a distinct violation. It should be tracked alongside the other Go violations in whatever resolution the design chooses.
- **[Major] Rollout Ordering & Gates:** The design's rollout (1–7) does not state that `All()` must stop embedding Maintenance/Gastown before their source trees are deleted, or that the public Gastown commit must be available and pinned before `PublicGastownPackVersion` is updated. Slice 6 (registry/cache) and slice 7 (source deletion) must have an explicit ordering dependency: pin first, registry second, deletion third.

---

### Cross-Review Convergence Analysis

All four reviewers (Claude, Codex, Gemini/DeepSeek, and this review) block on the same root cause: the Go source surface is unguarded while the design claims role neutrality. The specific Go hits vary in enumeration but converge on three categories:
1. **Role-keyed data structures:** `roleEmoji`, `roleIcons`, named theme functions (`MayorTheme`, `DeaconTheme`, `DogTheme`).
2. **Role-named string constants:** `defaultWarmupMailTo`, `DefaultCity`/`WizardCity` agent definitions, `cmd_prompt.go` fallback.
3. **Role-derived dispatch heuristics:** `mol-polecat-*`/`mol-refinery-patrol` prefix matching in `sling.go`, and `classifyAgentKind` check for `crew` in the API layer.

The disagreement is on scope: Claude and Codex want a production Go role-token guard for the migration surface; Gemini argues the Go surface should be explicitly out of scope but documented. My assessment: both are needed. A production guard for Go files touched by this migration prevents regression, and a documented exclusion list for untouched legacy Go prevents overreading the guard's meaning. The design currently does neither.

The `dog` ownership ambiguity is also convergent: all reviewers note that Core orders reference a `dog` pool whose agent definition lives in the Maintenance pack, and the design does not resolve whether Core ships that definition or expects the user to supply it. Gemini's Option (b) — make `dog` entirely user-supplied and remove `pool = "dog"` from Core orders — is the ZFC-clean solution, but it requires Core maintenance orders to use the configurable maintenance-agent name instead of a literal `dog`.

Gemini's observation that `mol-review-quorum` is the role-neutral reference pattern is correct and should be elevated to an explicit design decision. All Core dispatch should follow its variable-based pattern.

---

### Missing Evidence (Consolidated)

- **Go Migration Table:** For each Go role literal or heuristic listed above, the slice that resolves it, the replacement mechanism, and the test that proves the replacement works. Without this, lines 1105–1114's "Source scanner and config-driven theme test" is an unscoped promise, not a plan.
- **Concrete `dog` Ownership Decision:** Does Core ship `internal/packs/core/agents/dog/agent.toml`, or does every Core order replace `pool = "dog"` with a formula variable? The design must choose one and prove it against the SDK self-sufficiency test.
- **Scanner Contract Specification:** Root (`internal/packs/core`), closed set of file extensions, TOML field enumeration (step descriptions, `[vars]` defaults, metadata, order exec fields, template fragments), prohibited token set, `dog` allowlist format (token + path + field, not whole-file), comment/description coverage, and negative fixtures per category.
- **Replacement for script test:** A replacement for `TestTmuxThemeScriptHasNoHardcodedRoleNames` that covers both relocated pack assets and Go source, or an explicit scope statement.
- **`requiredBuiltinPackNames` Transition Test:** Changing from `["core", "maintenance"]` to `["core"]` plus verifying that an existing `.gc/system/packs/maintenance/` directory is ignored without fatal errors.
- **`PublicGastownPackVersion` Pinning Proof:** An immutable pin check demonstrating remote reachability and secure checksum verification.

---

### Required Changes

1. **Resolve the Go Role Surface Explicitly:** For each Go violation, state whether it is in-scope for this migration or deferred. If in-scope (as lines 1102–1103 imply by gating source deletion on it): add a rollout slice for Go de-roling, specify the replacement mechanism (config-driven theming for `MayorTheme`/`DeaconTheme`/`DogTheme`, lookup table for `roleEmoji`, configurable default for `defaultWarmupMailTo`, formula-variable dispatch for `sling.go` prefix heuristics, config-driven agent template for `DefaultCity`/`WizardCity`), and specify the Go scanner root, token set, and allowlist policy with the same rigor as the Core-asset scanner. If deferred: remove line 1114 from the blocking decision table and add an explicit exclusion note that a passing asset guard does not imply end-to-end Go role neutrality.
2. **Add `DogTheme()` Theme to Dog-Containment Analysis:** Either remove `DogTheme()` in favor of config-driven theming or document it as a deliberate exemption so line 696's "not a Go special case" invariant is actually true.
3. **Replace the Existing Role-Name Guard:** Either relocate `TestTmuxThemeScriptHasNoHardcodedRoleNames` to a Gas City package and extend it to cover Go source, or add a separate Go role-literal guard and explicitly scope the asset-only guard so that "guard passes" is not read as "Core is role-neutral end to end."
4. **Specify the `dog` Ownership Contract:** Choose:
   - **(a)** Core ships a minimal `agents/dog/agent.toml` that users can override/rename/disable, with a test proving renamed/omitted operation, or
   - **(b)** Core orders use a formula variable for the maintenance agent pool name, and `dog` is purely a Gastown convention.
   Do not leave Core orders routing to a literal `dog` pool while `dog` is defined in a pack that the design is removing from the embedded set.
5. **Convert the Role-Token Scanner from an Outcome to a Contract:** Specify the root, file types, TOML field coverage, prohibited tokens, `dog` allowlist format (token + path + field), and required negative fixtures. Cite `mol-review-quorum.toml` as the reference pattern for role-neutral Core dispatch.
6. **Add a `requiredBuiltinPackNames` Transition Gate in Slice 5:** The current code hard-codes `[]string{"core", "maintenance"}`; the migration must change it to `[]string{"core"}` and verify that cities with stale `.gc/system/packs/maintenance/` on disk still load correctly.
7. **Clarify Rollout Ordering:** State explicitly that the public Gastown replacement must be published and pinned before the embedded registry stops referencing Maintenance/Gastown, and that source directories are removed only after both are complete.

---

### Questions

- Is Go de-roling (`theme.go`, `tmux.go`, `config.go`, `sling.go`, `cmd_start_warmup.go`, `cmd_prompt.go`) in-scope for this migration? The design both gates source deletion on it (lines 1102–1103) and provides no mechanism, which is a self-imposed deadlock.
- What replaces `roleEmoji`/`roleIcons`? The obvious answer is config-driven theming (agent definitions include a theme key), but the design does not state this. If agent configs carry `theme: {name, bg, fg}` entries, the Go map becomes unnecessary. If not, what is the replacement?
- Should `DefaultCity`/`WizardCity`/`GastownCity` continue to hard-code a `mayor` agent, or should the default agent come from Core pack configuration?
- Is `mol-review-quorum.toml` the intended reference pattern for all Core dispatch? If so, the design should state this explicitly and require all other Core formulas to follow its variable-based dispatch pattern.
- Does the Core scanner cover comments, `description`, and metadata strings, given the existing guard strips comments and the moved orders carry role names in comments?
- How should `classifyAgentKind` be replaced? A generic classification based on agent config (e.g., a `kind` field in `agent.toml`) would avoid hardcoding Gastown role names in the API layer.
