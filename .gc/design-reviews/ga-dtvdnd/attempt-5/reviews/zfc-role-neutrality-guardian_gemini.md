# Alistair Sterling - DeepSeek V4 Flash (Gemini 3.5 Flash (High))

**Lane:** zfc-role-neutrality-guardian (wave 1) — zero hardcoded roles, Core role neutrality, dog configurability, SDK self-sufficiency.

**Verdict:** approve-with-risks

The requirements and the corresponding implementation plan represent a major milestone in securing Gas City's role neutrality and SDK self-sufficiency. By moving legacy in-tree role behavior to an external public pack authority (`gascity-packs/gastown`) and establishing Core as the sole required host-side pack, the SDK successfully sheds Gastown-specific framework cognition. 

While the architectural direction is robust, several critical, low-level technical inconsistencies and chicken-and-egg bootstrap loops exist between the Requirements and the Implementation Plan. These risks must be resolved to ensure the neutrality and self-sufficiency guarantees are bulletproof and verifiably correct.

---

### Top Strengths:
1. **Dynamic Configurable Bindings (ZFC-Compliant)**: Resolving Core formulas and orders through a symbolic binding table (`[gc.bindings.*]` and `[system_packs.*.bindings]`) rather than Go-side constants or hardcoded literal targets is an outstanding ZFC-compliant design. It keeps Go code strictly focused on transport/infrastructure, leaving role assignments entirely to configuration.
2. **Strict Absence-Scanning (AC8)**: The inclusion of an automated role-surface scanner that spans Go production code, templates, and generated/materialized metadata—complete with positive and negative controls—ensures that role-name leakage cannot silently degrade the neutrality guarantee over time.
3. **Controller Self-Sufficiency**: Codifying that safety-critical, structural, and deterministic maintenance tasks (e.g., store vacuuming, lock/cache pruning) must run natively in Go within the controller, leaving only qualitative cognitive tasks for optional agent execution, perfectly preserves the Bitter Lesson and SDK self-sufficiency invariants.

---

### Critical Risks:

- **[Major] Precedence Conflict: Environment Injection vs. Core Pack Default Bindings**:
  - *Risk*: The Implementation Plan states that "environment injection may supply a default only when neither city nor pack config names a binding" (lines 352–354). However, AC9 specifies that the required Core pack ships a default configured executor named `dog` in its configuration data (which is declared in `[system_packs.core.bindings]`). Because the Core pack is always loaded and always names this binding, the pack-level default is *always* present. Therefore, under the proposed precedence rule, the environment variable `GC_CORE_MAINTENANCE_WORKER` will *never* be able to inject or override anything. This makes the environment variable completely unreachable and dead-code.
  - *Recommendation*: Revise the binding precedence rule. Environment variable injection must have a higher priority than pack-level defaults but a lower priority than city-level overrides (i.e., the resolution order must be: `city.toml` -> `environment variable` -> `pack.toml` defaults).

- **[Major] Chicken-and-Egg Bootstrap of Provider Pack Materialization**:
  - *Risk*: The Required System Pack Loader is tasked with materializing and validating the required file sets for Core plus the selected provider pack (`bd` or `dolt` as selected today) *before* config resolution (lines 220–221). However, the selected beads provider is specified inside the city's configuration (`city.toml`). This creates a chicken-and-egg bootstrap loop: the loader cannot know which provider pack to materialize and validate without resolving the config, but it cannot resolve the config without first materializing the provider pack! Pushing the burden of pre-parsing or guessing the provider onto CLI/API callers violates encapsulation and DRY.
  - *Recommendation*: Define a clear, two-phase bootstrap loading sequence within `internal/systempacks`. Phase 1 must perform a lightweight, raw, or partial parse of the city's config to extract the beads provider name and any explicit import URLs. Phase 2 must then materialize the matching provider pack, validate its file set, and perform final, verified config resolution.

- **[Major] City-Level Advisory Lock is Insufficient for Shared Cache Concurrency**:
  - *Risk*: Under "Doctor And Runtime-State Mutation Safety" (lines 297–298), the coordinator acquires a "city advisory lock before digest preflight." However, the public pack cache (used to resolve `gascity-packs/gastown`) is a shared, user-wide or system-wide directory, not isolated to a single city. If multiple cities or parallel test processes (such as sharded integration runs) execute concurrently, they will attempt to promote or write to the *same* shared cache. Because the advisory lock is city-scoped (bound to a single city's `.gc/` directory), it cannot prevent concurrent write collisions, file-handle races, or cache corruption in the shared system-wide cache.
  - *Recommendation*: The doctor/install coordinator must acquire a separate, global file-based lock (e.g., a write lock on the shared cache directory or cache promotion temp path) to protect promotions and cache writes from multi-city concurrent write collisions.

- **[Major] Schema/Metadata Deficit for Required vs. Optional Bindings (ZFC Violation)**:
  - *Risk*: The design states that "missing optional bindings skip user-agent work with a typed diagnostic" while "missing required provider-pack escalation bindings fail the formula/order before dispatch" (lines 354–356). However, there is no schema or configuration mechanism defined to declare whether a given binding is optional or required. If the engine has to hardcode which bindings are "optional" (like `maintenance_worker`) vs. "required" directly in the Go codebase, this introduces Go-side special-casing of roles/bindings, which directly violates the Zero Framework Cognition (ZFC) rule.
  - *Recommendation*: Prohibit hardcoded required/optional binding classifications in Go. Require all formula, order, or pack manifests to explicitly declare their bindings and whether they are `optional: true` or `required: true` in their metadata/TOML, allowing the engine to evaluate compliance data-drivenly.

- **[Minor] Go Source Omission from Wording/Docs Scanner**:
  - *Risk*: The wording scanner in lines 400–403 covers Markdown, MDX, JSON, TXT, TS, and OpenAPI, but explicitly omits `.go` (Go source) files. While the role-surface neutrality scanner (AC8) checks Go code for specific role names, it does not scan for retired path patterns or legacy pack terminology (such as `examples/gastown/packs/maintenance`). Omitting Go code from the wording scanner allows stale comments, outdated docstrings, or test assertions containing retired paths to leak into production Go files.
  - *Recommendation*: Add `.go` source files to the wording scanner's target list. Use a specialized token/AST-based scanner to lint only comments and string literals, preventing legacy path leakage in Go files without triggering false positives on actual Go identifiers.

---

### Missing Evidence:
- **No Verification of Public Pack File Completeness**: While `gascity-packs/gastown/ownership.yaml` is specified to assign every behavior-bearing asset to its successor, there is no tool or rule specified to verify that this list is *complete* (i.e., proving that no files present in the legacy source-tree were accidentally abandoned or forgotten during the migration).
- **No Concrete Schema for RequiredSystemPackParticipation**: The plan specifies the fields of this proof record (lines 435–439) but lacks a concrete serialization schema or structure, leaving its persistence format under `.gc/` unverified.

---

### Required Changes:
1. **Fix Precedence Resolution for Environment Variables**: Explicitly state in the binding precedence rules that environment variables override pack-level defaults but are overridden by city-scoped config.
2. **Establish Two-Phase Loading Sequence**: Document the two-phase bootstrap loading sequence in `internal/systempacks` to resolve the provider-conditioning chicken-and-egg problem without violating encapsulation.
3. **Add Global Cache Lock for Public Pack Promotion**: Require the coordinator to acquire a global file lock on the shared cache directory or cache promotion temp path before executing atomic writes or promotions.
4. **Data-Drive Binding Optionality**: Specify that the optional/required status of all bindings must be declared as metadata in the formula, order, or pack configurations, rather than being determined via Go-side special-casing.
5. **Include Go Comments in Wording Scanner**: Extend the wording scanner to analyze comments and string literals in `.go` source files to prevent legacy path and terminology leakage.

---

### Questions:
- If a city uses a custom local overlay that references a retired pack path, does the doctor mutation coordinate a safe rewrite, or does it refuse to fix it automatically and instead output manual remediation instructions?
- Will the global cache lock support a timeout and fail-closed behavior to prevent concurrent CLI commands from hanging indefinitely in a deadlock state?
