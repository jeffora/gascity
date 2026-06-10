# Anand Krishnaswamy — Role Neutrality & ZFC Invariant Reviewer (Iteration 10 / Attempt 1, Independent DeepSeek V4 Flash Style)

**Verdict:** block

> **Lane:** Zero hardcoded roles in Go and assets, the symbolic maintenance-worker binding, SDK self-sufficiency, and ZFC (Zero Framework Cognition) judgment containment.
>
> Reviewed against the Iteration 10 design document (`plans/core-gastown-pack-migration/implementation-plan.md`, 835 lines, `updated_at: 2026-06-09T13:20:59Z`) — specifically §"Role Neutrality And Configurable Bindings" (lines 409–458), §"Required System Pack Loader" (lines 221–312), §"Data And State" (lines 532–603), and §"Testing" (specifically Go and Asset Role-Surface and Scanner tests at lines 624–638).
>
> This independent review is produced using the DeepSeek V4 Flash style, focusing rigorously on first-principles ZFC compliance, cross-document state consistency, and identifying assumptions other reviewers may accept too quickly.

---

## Schema Conformance

**Conforms.** The Iteration 10 implementation plan includes all required top-level sections in the correct order, carries the required front matter (`phase: implementation-plan`), and `Open Questions` is correctly marked as `None`. The role neutrality, symbolic binding resolver, and active-root scanner designs are properly integrated into the Proposed Implementation, Data And State, Testing, and Rollout sections, rather than being appended as unstructured review prose.

---

## Top Strengths of the Proposed Design

1. **Concrete Binding Indirection without Go-Level Fallbacks:**
   The plan's implementation of `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER` provides a robust, data-driven layer of indirection. Eliminating Go-level fallbacks that substitute concrete role names (lines 445–446) is a vital first-principles alignment with Zero Framework Cognition (ZFC).
2. **Active Root Containment and Enumeration Guarding:**
   Constraining loaders, installers, compilers, prompt scanners, and doctor checks to resolve directory roots via `internal/packsource.ActiveRootsFor(kind)` (lines 418–424) is an excellent architectural boundary. This structurally prevents ad-hoc globbing and prefix-based path scans, which historically has been the primary vector for stale or out-of-scope asset leaks.
3. **Decoupled SDK Self-Sufficiency and Participation Gates:**
   Decoupling controller-owned SDK infrastructure operations from configured user agents (lines 435–437) ensures that removing or renaming the maintenance worker does not dead-lock the system. Furthermore, verifying required-pack participation at config-resolution time via a typed `RequiredSystemPackParticipation` record ensures that Core's participation is a runtime invariant.

---

## Critical Risks & Architectural Inconsistencies

### 1. [Blocker] Un-de-roled Default Scaffolding in `gc init` and Config Defaults
* **The Risk:** In `internal/config/config.go`, the out-of-the-box non-Gastown SDK defaults `DefaultCity` and `WizardCity` still explicitly hardcode the `"mayor"` agent and `"prompts/mayor.md"` prompt template.
* **The Impact:** Running `gc init` or spawning a fresh default city creates an active city that depends on a literal `"mayor"` role and template. This violates the zero-hardcoded-roles SDK guarantee. Since AC8 scopes role neutrality to "init/template resolution," the static scanner will flag this unless it is allowlisted—and allowlisting a live SDK default defeats the entire de-roling migration.
* **The Gap:** While "default scaffolding" is listed under the manifest coverage in line 414, the plan never specifies the neutral end-state for these functions.
* **Resolution:** Explicitly declare the role-neutral scaffolding for `gc init`. The default city must scaffold zero agents and zero named sessions, or alternatively use purely symbolic-binding equivalents. Add an AC8 check proving the non-Gastown default path contains no concrete role literals.

### 2. [Blocker] Warmup Mail Fallback Hardcoded to `mayor` in Core CLI Logic
* **The Risk:** `cmd/gc/cmd_start_warmup.go` hardcodes `defaultWarmupMailTo = "mayor"`. Warmup runs under `gc start` (a core SDK/controller command).
* **The Impact:** When running `gc start` on a city with a renamed or omitted maintenance worker, warmup failure mail will attempt to send to a non-existent `"mayor"`. This directly contradicts the plan's own "No Go fallback may substitute `mayor`, `deacon`, `dog`, or another concrete role name" rule (lines 445-446).
* **The Gap:** The plan lists "warmup mail defaults" under manifest coverage (lines 414-415) but fails to define a warmup-recipient binding, leaving this hardcoded fallback active.
* **Resolution:** Define how the warmup mail recipient resolves with no Go fallback. It must resolve via a symbolic binding (e.g., `escalation_recipient`), or skip with a typed diagnostic when unbound. Name `cmd/gc/cmd_start_warmup.go` and its Core end-state in the design.

### 3. [Blocker] Compile-bound Tmux Theme and Emoji Maps in Go Runtime
* **The Risk:** `internal/runtime/tmux/theme.go` defines `MayorTheme()`, `DeaconTheme()`, and `DogTheme()`. Furthermore, `internal/runtime/tmux/tmux.go` defines a static emoji/icon map keyed on literal roles like `"mayor"`, `"deacon"`, `"witness"`, `"refinery"`, `"crew"`, `"polecat"`, `"coordinator"`, `"health-check"`.
* **The Impact:** These Go functions and map keys are compile-bound in the runtime packages; they cannot be moved to configuration packs.
* **The Gap:** The design states tmux theme APIs are "assigned by manifest rows before source moves" (line 452). This is a category error; Go functions cannot be relocated to config packs.
* **Resolution:** Fully de-role `internal/runtime/tmux`. Drive status themes and emoji maps dynamically from config/pack bindings or a consistent hash fallback (`AssignTheme(agentName)`). Rename `ConfigureGasTownSession` to a neutral name.

### 4. [Major] Suffix Binding Gap & `binding_prefix` Template Dependencies
* **The Risk:** Prompt templates currently resolve prefixes using `binding_prefix` but leave role name suffixes (e.g., `dog`, `deacon`, `witness`) hardcoded in the templates (e.g., `{{ .BindingPrefix }}dog` or prompt files like `prompts/mayor.md` without suffixes).
* **The Impact:** The new symbolic binding system is not reconciled with how suffix-level prompt templates are resolved, which will result in code that still hardcodes literal suffixes.
* **Resolution:** Require that all role suffixes in prompt templates are replaced by symbolic bindings, deprecating prefix-only `binding_prefix` in favor of fully symbolic config-driven bindings.

### 5. [Major] Scanner Excludes `dog` and Lacks Expiry Failures
* **The Risk:** The scanner's denied token set (lines 426-427) completely excludes `"dog"`.
* **The Impact:** Developers can silently hardcode `"dog"` in Go files or scripts without failing CI.
* **Resolution:** Add `"dog"`, `"coordinator"`, and `"health-check"` to the scanned denied token set, allowing `"dog"` *only* in the designated Core default pack configuration file and its associated tests. Enforce that any allowlist row with an expired `expiry` date fails the build in CI.

### 6. [Major] Declarative Required-versus-Optional Binding Semantics Gap (ZFC Violation)
* **The Risk:** The plan distinguishes between "missing optional bindings" and "missing required provider-pack escalation bindings" (lines 443-444), but never specifies where this optional-vs-required designation is declared.
* **The Impact:** If Go classifies a binding as required or optional by its name or purpose, that is a ZFC judgment call.
* **Resolution:** Ensure that whether a binding is optional or required is declared as metadata within the formula/order/pack config itself rather than checking hardcoded names in Go.

### 7. [Major] Missing Structural Binding Evidence and Target-Field Resolution Trace
* **The Risk:** The plan specifies how the binding resolver behaves, but does not provide an auditable trace.
* **The Impact:** Given the current tree has many literal role strings, token scanning alone cannot prove that active behavior came through the generic resolver rather than through string interpolation or a helper-specific fallback.
* **Resolution:** Require a machine-readable `BindingResolution` trace schema that records consumer id, consumer kind, source pack, binding key, required/optional mode, source of the selected binding, resolved target, absence behavior, and diagnostic id.

---

## Detailed Responses to Lane-Specific Questions

### Q1: After binding indirection, does any Go, prompt asset, script, formula, order, generated help, or API route still branch on dog, Mayor, Maintenance, or another concrete role name?

**Answer: Yes, several compile-bound role leakage vectors remain unneutralized in Core:**
1. **Default Scaffolding (`DefaultCity` / `WizardCity`):** Still contains hardcoded inline references to `"mayor"` and `"prompts/mayor.md"`.
2. **Warmup Mail Fallback:** `cmd/gc/cmd_start_warmup.go` hardcodes `defaultWarmupMailTo = "mayor"`.
3. **Compile-bound Tmux Helpers:** `internal/runtime/tmux/theme.go` defines `MayorTheme/DeaconTheme/DogTheme`, and `tmux.go` maps emoji icons to literal role keys like `"mayor"`, `"deacon"`, and `"witness"`.
4. **Required Provider Packs (`bd`/`dolt`):** Provider-pack scripts contain hardcoded routes like `gc mail send mayor/` or `gc session nudge deacon/` which will fail on non-Gastown cities.
5. **Prompt Template Suffixes:** Suffix role names remain hardcoded in templates behind `binding_prefix`.

### Q2: Can controller-owned SDK operations still run when the configured maintenance worker is renamed or omitted, with no dependency on a user agent entry?

**Answer: Yes, but with unresolved ZFC and declaration gaps:**
* If the `maintenance_worker` is renamed (e.g., from `dog` to `reconciler`), the framework resolves the target at runtime via `gc.run_target_binding` / `target_binding`, which is robust.
* However, if the `maintenance_worker` is omitted entirely from the config, the plan's behavior is unstated. Line 443 states: `"Missing optional bindings skip user-agent work with a typed diagnostic."` Under **ZFC**, the Go code must not make a judgment call about omitting required system-level transport workers; the config parser must fail-closed during pre-flight configuration validation or raise a descriptive pre-flight error rather than letting the dispatcher make an ad-hoc runtime judgment.
* Furthermore, the plan distinguishes between "missing optional bindings" and "missing required provider-pack escalation bindings" (lines 443-444), but never specifies where this optional-vs-required designation is declared. If Go classifies a binding as required or optional by its name or purpose, that is the judgment call ZFC forbids.

### Q3: Are role-name allowlists narrow, time-bounded, and failing when compatibility fixtures leak into live behavior?

**Answer: No, the allowlist model lacks strict temporal enforcement and excludes `dog` entirely:**
1. **The Scanner ignores `dog`:** The plan's proposed list of denied tokens (`mayor, deacon, witness, refinery, polecat, boot, crew, gastown`) completely omits `dog`. While `dog` is allowed in the Core default pack config, omitting it from the denied set means developers can silently hardcode `dog` in Go source code or script bodies without triggering a build failure.
2. **Missing Expiry Enforcement:** While the plan mentions that allowlist rows require an `expiry` date (line 594), it specifies no CI enforcement gate that fails the build when a row is past its expiry date. Without this, allowlists will grow indefinitely.

---

## Required Changes

Before the design can transition to implementation, the following changes must be incorporated into the proposed implementation plan:

1. **De-role Default Scaffolding (`gc init`):** Declare the role-neutral default for `DefaultCity` and `WizardCity`. They must emit a city with no agents/named sessions, or use symbolic-binding equivalents. Add an AC8 gate proving the non-Gastown default path contains no role literals.
2. **De-role Warmup Mail:** Define how the warmup mail recipient resolves with no Go role fallback. Name `cmd/gc/cmd_start_warmup.go` and its Core end-state.
3. **De-role Tmux Themes & Icons:** Deprecate `MayorTheme()`, `DeaconTheme()`, and `DogTheme()` in `internal/runtime/tmux/theme.go`. Drive status themes and emoji maps dynamically from config/pack bindings. Rename `ConfigureGasTownSession`.
4. **Suffix-Level Symbolic Bindings:** Require that all role suffixes in prompt templates are replaced by symbolic bindings, deprecating prefix-only `binding_prefix` in favor of config-driven bindings.
5. **De-role Required Provider Packs (`dolt`):** Map all hardcoded `mayor`/`deacon` mail/nudge escalation routes inside `examples/dolt` to configurable symbolic recipients, and fail CI on any hardcoded literal role route in a required provider pack.
6. **Specify Required-vs-Optional Semantics as Declarative Data:** Ensure that whether a binding is optional or required is declared as metadata within the formula/order/pack config itself rather than checking hardcoded names in Go.
7. **Add `dog` to Denied Set and Enforce Expiry:** Add `dog`, `coordinator`, and `health-check` to the active behavior denied token list (with narrow allowlists for Core defaults, Dolt/store-maintenance terms, and tests). Enforce that any allowlist row with an expired `expiry` date fails the build in CI.
8. **Add Binding-Resolution Trace Schema:** Expose a machine-readable resolution record/trace schema for target fields (`gc.routed_to`, mail/nudge targets, warmup recipient, prompt fallbacks) to verify all active routing flows through the generic resolver.

---

## Questions

1. Is `gc start` warmup Core SDK infrastructure or Gastown-owned? If Core, the recipient must be neutral/bound; if Gastown, the warmup-mail default belongs in the public pack entirely.
2. For Core-retained surfaces the manifest assigns "to Core," is the only de-roling primitive the maintenance-worker binding, or will the plan provide a general symbolic binding table (default-city agents, warmup recipient, tmux theme/icon)?
3. Does the CI scanner enforce a strict build failure if any allowlist row has expired?
