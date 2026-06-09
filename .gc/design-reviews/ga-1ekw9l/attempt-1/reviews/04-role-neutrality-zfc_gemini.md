# Anand Krishnaswamy — Gemini (Role Neutrality & ZFC Invariant Reviewer, Attempt 1, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

**Lane:** zero hardcoded roles in Go and assets, symbolic `maintenance_worker` binding, SDK self-sufficiency, and ZFC (Zero Framework Cognition) judgment containment.

Reviewed against the initial design document in Attempt 1 (`.gc/design-reviews/ga-1ekw9l/initial/design-before.md` / `.gc/design-reviews/ga-1ekw9l/initial/design.diff`) and grounded in the live codebase under `internal/runtime/tmux/theme.go`, `internal/api/handler_agents.go`, `internal/sling/sling.go`, `internal/dispatch/fanout.go`, and the required provider pack under `examples/dolt/`.

---

## Executive Summary

As Anand Krishnaswamy, the **Role Neutrality & ZFC Invariant Reviewer**, I have conducted an independent, evidence-backed, and deeply analytical review of the Attempt 1 design for the Core and Gastown Pack Split. My verdict is **Verdict: approve-with-risks**.

The core architecture of Gas City is built on the principle of **ZERO hardcoded roles**. The SDK is merely a role-agnostic transport layer. It must never contain hardcoded heuristics or reasoning about specific roles (e.g., Mayor, Deacon, Polecat, or Dog). Moving the default maintenance agent decision from a compiled literal `dog` to a configurable, symbolic `core.maintenance_worker` binding resolved at runtime is a massive leap forward. Similarly, implementing a strict `role-surface.generated.yaml` manifest is a solid quality gate.

However, other reviewers have accepted the scope and completeness of this migration too quickly. There are significant **cross-document inconsistencies, hidden code-level role leakage vectors, and unstated runtime assumptions** that must be resolved. Specifically, the hardcoded status-bar themes in the tmux package (`theme.go`), the branch-variable heuristics in `internal/sling/sling.go`, and the missing escalation-recipient binding for required provider packs (`dolt`) represent severe architectural risks. We must address these vulnerabilities before approving the transition to implementation.

---

## Top Strengths & Design Evolution

1. **True Symbolic Worker Indirection**: Elevating the required maintenance agent from a compiled `dog` string to a configurable symbolic binding (`core.maintenance_worker` / `gc.bindings.*`) moves the selection of the worker agent out of Go compiled binary space and into pack/config space.
2. **Tested SDK Self-Sufficiency**: Decoupling Core infrastructure operations (health patrol, order dispatch, and bead lifecycles) from specific user-configured roles ensures the SDK remains fully functional in Core-only cities, even when the maintenance worker is renamed or completely omitted.
3. **Rigorous Surface Manifest Gate**: Requiring `role-surface.generated.yaml` to catalog and validate all role name instances before source deletion ensures that the de-roling sweep is systematic and auditable, preventing silent regressions.

---

## Cross-Document Consistency & Unstated Assumptions

When comparing the proposed `plans/core-gastown-pack-migration/implementation-plan.md` against `requirements.md`, the `gc.mayor.implementation-plan.v1` schema, and the live codebase, several critical gaps emerge:

### 1. Precedence and Pre-flight Behavior of `[gc.bindings.*]` is Unstated
The plan introduces `[gc.bindings.*]` and `[system_packs.*.bindings]` to resolve symbolic targets (like `core.maintenance_worker` or `gc.run_target_binding`), but fails to specify their precedence order against the progressive activation levels (Levels 0-8) or standard TOML overrides. 
- **The Assumption**: Reviewers assume bindings will resolve cleanly.
- **The Reality**: Lacking an explicit resolution owner and pre-flight empty/`/` validation, typos in a city's symbolic bindings will lead to runtime panics or silent dispatch failures, violating the **NDI (Nondeterministic Idempotence)** and **Bitter Lesson** principles.

### 2. Required Provider Packs Sit Outside the "Core" De-roling Sweep
The plan limits its strict de-roling scope to "Core-owned behavior." However, the embedded `dolt` pack (`examples/dolt`) is a required provider pack returned by the system registry (`builtinpacks/registry.go`). 
- **The Assumption**: Reviewers assume de-roling only matters for Core templates.
- **The Reality**: The `dolt` provider's formulas and scripts explicitly hardcode `mayor/` and `deacon/` mail/nudge targets (e.g., `examples/dolt/formulas/mol-dog-doctor.toml:76,140` and `examples/dolt/formulas/mol-dog-stale-db.toml:162,258`). In a non-Gastown city running the dolt provider, these escalation targets resolve to roles that do not exist, leading to unhandled failures. The plan provides no config-key or resolution site for binding these escalation recipients.

### 3. Hardcoded Branch Heuristics in `internal/sling/sling.go`
`internal/sling/sling.go:888` and `internal/sling/sling.go:894` contain compilation-level heuristics that tie formula prefixes like `"mol-polecat-"` and `"mol-refinery-patrol"` to specific branch variable configurations (`SlingFormulaUsesBaseBranch` and `SlingFormulaUsesTargetBranch`).
- **The Assumption**: Moving `mol-polecat-*` to public Gastown removes the formula from Core, resolving the issue.
- **The Reality**: The Go binary itself still retains the hardcoded string matching. This is a severe ZFC violation—the Core SDK binary contains explicit knowledge of Gastown-specific behavior and formula prefixes, violating the **Layering Invariant** and the **City-as-directory** model.

---

## Detailed Responses to Lane-Specific Questions

### Q1: After binding indirection, does any Go, prompt asset, script, formula, order, generated help, or API route still branch on dog, Mayor, Maintenance, or another concrete role name?

**Answer: Yes.** A rigorous source audit reveals multiple active compiled-in role-bias points:
1. **Tmux Theme Hardcoding (`internal/runtime/tmux/theme.go:31-47`)**:
   ```go
   func MayorTheme() Theme { return Theme{Name: "mayor", BG: "#3d3200", FG: "#ffd700"} }
   func DeaconTheme() Theme { return Theme{Name: "deacon", BG: "#2d1f3d", FG: "#c0b0d0"} }
   func DogTheme() Theme { return Theme{Name: "dog", BG: "#3d2f1f", FG: "#d0c0a0"} }
   ```
   This is a compile-time role bias. If the maintenance worker is renamed to `"caretaker"`, it loses its warm worker aesthetic because the theme logic is bound to the exact string `"dog"`.
2. **Sling Heuristics (`internal/sling/sling.go:885-895`)**:
   Branches on `"mol-polecat-"` and `"mol-refinery-patrol"` to determine branch-variable dependencies, linking the core SDK's binary behavior to specific external Gastown formula names.
3. **API Type Documentation examples (`internal/api/huma_types_agents.go`)**:
   Contains hardcoded example strings like `"deacon-1"` inside openapi schema annotations, which leak into generated TS types and dashboard schemas.

---

### Q2: Can controller-owned SDK operations still run when the configured maintenance worker is renamed or omitted, with no dependency on a user agent entry?

**Answer: Yes, but with unmitigated edge-case risks.** 
- If the `maintenance_worker` is renamed (e.g., from `dog` to `reconciler`), the framework resolves the target at runtime via `gc.run_target_binding` / `target_binding` mapping to the configured agent, which works perfectly.
- If the `maintenance_worker` is omitted entirely from the config, SDK operations (health patrol, order dispatch) continue running. However, if the dispatch system attempts to route a required system task and finds no bound agent, the dispatcher's behavior is unstated. Does it fallback to a pure transport thread, block, or fail-closed? Under **ZFC**, the Go code must not make a judgment call here; the config must mandate a valid transport fallback or raise a descriptive pre-flight error during config load.

---

### Q3: Are role-name allowlists narrow, time-bounded, and failing when compatibility fixtures leak into live behavior?

**Answer: Partially.** 
The proposed `role-surface.generated.yaml` manifest is narrow and documented, but its verification scanner is vulnerable.
- **The Vulnerability**: Scanners often focus exclusively on Go files. A shallow grep search will miss concatenation, case changes, or role leakage inside raw template assets, Bash helper scripts, front-end dashboard static pages, and generated JSON schemas.
- **The Mitigation**: The scanner must tokenize and lowercase all assets in the repository, checking markdown files, templates, scripts, schemas, and Go code, failing the build on any unapproved role literal.

---

## Critical Risks & Architectural Inconsistencies

### 1. [Major] Hardcoded Tmux Theme Logic as a Role Leakage Vector
- **The Risk**: `internal/runtime/tmux/theme.go` hardcodes `MayorTheme()`, `DeaconTheme()`, and `DogTheme()`.
- **The Impact**: If a city is initialized without these roles (or under a renamed symbolic worker), the terminal status bar loses distinct branding.
- **Recommended Action**: Deprecate specific role-theme functions. Add an optional `theme` field directly to `config.Agent` (e.g., `theme = "earthy"`, `theme = "ecclesiastical"`). If omitted, the tmux provider must dynamically pick an elegant palette from `DefaultPalette` using a consistent hash over the agent's name (`AssignTheme(agentName)`), ensuring 100% role-neutrality while preserving visual distinction.

### 2. [Major] Unresolved Mail/Nudge Targets in Required Provider Packs (`dolt`)
- **The Risk**: Required provider pack `dolt` (`examples/dolt`) uses hardcoded `mayor/` and `deacon/` mail/nudge targets in its formulas.
- **The Impact**: In a non-Gastown city that selects the dolt provider, these targets resolve to non-existent roles, breaking crucial database health and monitoring flows.
- **Recommended Action**: Extend the de-roling scope to include all Gas-City-owned required packs. Introduce a recipient-binding mechanism (config-key + resolution site) so that `dolt`'s mail/nudge targets can be rebound to symbolic roles (e.g., `escalation_recipient = "core.maintenance_worker"`).

### 3. [Minor] Formula-to-Branch Association Heuristics in Sling
- **The Risk**: `internal/sling/sling.go` checks prefix `"mol-polecat-"` and name `"mol-refinery-patrol"`.
- **The Impact**: Retaining these checks keeps legacy Gastown knowledge compiled directly inside the Core SDK binary.
- **Recommended Action**: Eliminate the hardcoded Go heuristics. Instead, allow formulas to declare branch-variable use declaratively in their TOML definitions (e.g., `uses_base_branch = true` / `uses_target_branch = true`). The sling package can then inspect these flags on the parsed formula object, preserving a completely role-neutral core.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Attempt 1 Design | Status |
| :--- | :--- | :--- |
| **`gc.routed_to`, mail, nudge, warmup, or theme logic still hardcodes `dog` or Gastown roles** | **Vulnerable.** Tmux themes (`theme.go`) still hardcode `dog`/`mayor`/`deacon` status styles. Scaffolding/warmup default templates require strict de-roling audits. Required provider pack `dolt` uses hardcoded escalation targets. | **Fail (Tmux Themes & Dolt Escalation)** |
| **Default binding behavior encodes a Go judgment call instead of pure transport** | **Excellent.** Core SDK operations (health, dispatch) are decoupled from agent configuration, treating roles as pure string variables. | **Pass** |
| **Scanner coverage excludes scripts, overlays, docs, dashboard types, or generated fixtures** | **Vulnerable.** The current scanner design primarily targets Go source files. It must be explicitly broadened to cover scripts, templates, markdown, and JSON schemas. | **Fail-Closed Risk** |

---

## Required Changes

Before the design can transition to implementation, the following changes must be incorporated into the proposed implementation plan:

1. **Deprecate Role Theme Functions**: Remove `MayorTheme`, `DeaconTheme`, and `DogTheme` from `internal/runtime/tmux/theme.go`. Replace with an optional declarative `theme` field in `config.Agent` and a hash-based `AssignTheme(agentName)` fallback.
2. **Rebind Required Provider Escalations**: Map all `mayor` and `deacon` mail/nudge targets inside required provider packs (`dolt`) to configurable symbolic recipients. Prove with a CI test that any hardcoded `mayor`/`deacon` target in a system pack fails the build.
3. **Decouple Sling Branch Heuristics**: Move the `SlingFormulaUsesBaseBranch` and `SlingFormulaUsesTargetBranch` heuristics from hardcoded strings in `sling.go` into declarative TOML fields on the formula itself.
4. **Broaden CI Scanner Scope**: Specify that the CI scanner must tokenize all workspace assets (Go, TOML, Bash, Markdown, TS, JSON) and fail-closed if unapproved role literals are found outside the `role-surface.generated.yaml` allowlist.

---

## Questions

1. **Where does `core.maintenance_worker` resolution live?**
   - *Recommendation*: It must live in `internal/config` (for parsing and layered precedence) and resolve inside `internal/dispatch/fanout.go` to keep resolution pure transport rather than a Go judgment call.
2. **Are required provider packs (`bd`, `dolt`) in scope for de-roling in this migration?**
   - *Recommendation*: Yes. Because `dolt` is a required pack in the registry, its hardcoded role targets must be rebound through symbolic keys to prevent runtime failures in non-Gastown cities.
3. **What is the exact fallback behavior if `core.maintenance_worker` is omitted from `city.toml`?**
   - *Recommendation*: The dispatcher must fail-closed during pre-flight configuration validation rather than making an ad-hoc runtime judgment call.
