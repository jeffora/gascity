# Alistair Sterling - DeepSeek V4 Flash (Gemini 3.5 Flash (High))

**Verdict:** approve-with-risks

### Top Strengths:
- **Comprehensive Symbolic Binding Design**: The proposed design for `[gc.bindings.*]`, `[system_packs.*.bindings]`, and `target_binding` (lines 343–348) represents a massive leap forward in role neutrality. By resolving all Core formulas and orders through a dynamic binding table rather than hardcoded Go constants or literal strings, the design successfully prevents literal role-name coupling at the schema layer.
- **Robust Role-Surface Absence-Scanning**: The implementation of an automated absence scanner (AC8, lines 330–335) that scans Go code, Core assets, formulas, orders, and generated metadata using a strict denied-token manifest is a powerful defense. Having positive and negative controls ensures that the scanner itself cannot silently degrade or become a sieve over time.
- **ZFC Controller Self-Sufficiency**: The design successfully codifies that the SDK infrastructure is completely self-sufficient. Under the "no-executor city" edge case, the controller and Core config are proven capable of running all system operations without assuming that any Gastown agents or user-configured executors exist.

---

### Critical Risks:

- [Major] **"Dog Prompt Fragments" Generalization Deficit (Leakage of Role Assumptions in Core Prompts)**:
  - *Risk*: Under "External Public Gastown Prerequisite" (lines 120–121) and "Role Neutrality" (lines 363–364), the design mentions assigning and moving "Dog prompt fragments" to Core. This is a severe, unaddressed role-neutrality leak. If Core-owned prompt fragments contain literal references to `dog` or embody role-specific behavior (e.g., assuming a canine identity, using specific phrasing, or expressing Gastown-specific heuristics), then role coupling has simply been moved from the Go codebase into Core prompt templates under the guise of "default configurability." If an operator renames the default maintenance executor to `reaper-bot` using bindings, the prompt templates in Core will still generate prompt text referring to `dog` or canine personas, breaking role neutrality.
  - *Recommendation*: Mandate that all prompt fragments and templates moved to Core must be completely generalized and written in a role-neutral, functional manner (e.g., "you are the system maintenance executor"). Any reference to the executor's identity or name must be dynamically interpolated using Go template variables (e.g., `{{.Bindings.maintenance_worker}}` or `{{.Config.maintenance_worker_name}}`) rather than containing hardcoded literal names.

- [Major] **Implicit Hardcoded Go-Side Fallback on `GC_CORE_MAINTENANCE_WORKER`**:
  - *Risk*: The design introduces the `GC_CORE_MAINTENANCE_WORKER` environment variable (line 344) as an injection fallback. While environment variable default overrides are helpful, there is an implicit risk of Go-side "leakage" if the loading sequence falls back to the hardcoded `"dog"` string when both the config-bindings and environment variables are empty. If any Go-side function contains:
    ```go
    if worker == "" {
        worker = "dog"
    }
    ```
    it directly violates AC9 ("No Go fallback may substitute mayor, deacon, dog, or another concrete role name").
  - *Recommendation*: Explicitly require that the Go-side loader never hardcodes `"dog"` as a fallback value when configuration and environment variable bindings are absent. The default `"dog"` value must be declared *solely* within the Core pack's configuration file (`pack.toml` under `[system_packs.core.bindings]`). If all configuration layers and environment variables are empty, the binding must resolve to empty, and the controller must handle it as an unresolved optional binding (emitting a diagnostic and skipping agent-executed work, rather than substituting `"dog"`).

- [Major] **Silent Degradation of "No-Executor Cities" due to Deterministic Task Delegation**:
  - *Risk*: The design specifies that "missing optional bindings skip user-agent work with a typed diagnostic" (line 354). However, the implementation plan never defines a taxonomy of what "Core maintenance work" actually consists of. If safety-critical deterministic tasks (such as task store vacuuming, stale wisp reaping, process-table reconciliation, or database compaction) are delegated to the maintenance executor (the agent) rather than executed natively by the controller, then disabling or omitting the executor will cause the city's state and database to silently bloat and degrade.
  - *Recommendation*: The design must explicitly partition Core maintenance tasks. All structural, safety-critical, or deterministic operations must be implemented as native Go-side controller functions or deterministic scripts, completely independent of any agent. The maintenance executor (`dog`) must *only* be invoked for qualitative, non-deterministic tasks that genuinely require an LLM's cognitive judgment (e.g., analyzing execution transcripts of stalled sessions to summarize why they crashed).

- [Minor] **Context-Sensitive False Positive Mitigation in Wording/Docs Scanner**:
  - *Risk*: AC12 and AC13 mandate a generated wording/docs scanner. However, role names like `boot` and `crew` are also extremely common generic English words. A naive substring or regex search will trigger massive false-positive noise (e.g., matching "bootstrap", "reboot", or "team and crew"), causing developer fatigue and leading to the exclusion list becoming too broad.
  - *Recommendation*: Explicitly specify that the wording/docs scanner must be AST-aware or token-aware when scanning Go code, and must employ context-sensitive parsing (such as word-boundary matching and negative-lookahead filters) when analyzing Markdown/MDX prose, ensuring that legitimate, neutral occurrences of "boot" or "crew" do not trigger false positives.

---

### Missing Evidence:
- **Taxonomy of Maintenance Tasks**: No file or design block defines what operations constitute "Core maintenance work." There is no explicit contract proving that safety-critical database or process pruning is kept within the controller's deterministic bounds.
- **AST-Based TOML Parser Verification for Binding Updates**: While the plan introduces scoped TOML edits for config repair, it lacks a concrete test contract asserting that custom formatting, comments, and unrelated array order are fully preserved when repairing bindings in `city.toml`.

---

### Required Changes:
1. **Generalize/Template "Dog Prompt Fragments" in Core**: Refine AC8 and the "Role Neutrality" section (lines 330–335) to require that any prompt templates/fragments moved to Core are entirely role-neutral and refer to the executor dynamically using template variables (e.g., `{{.Bindings.maintenance_worker}}`), completely eliminating hardcoded "Dog" persona instructions from Core assets.
2. **Strictly Forbid Hardcoded Go-Side "dog" Fallbacks**: Add an explicit constraint to the required-pack loader and bindings parser stating that if the configuration layers and environment variable fallbacks are empty, no Go-side code may substitute "dog" or any other concrete name. The default value must live exclusively in the Core pack's TOML config.
3. **Confine LLM Maintenance Executor to Qualitative Tasks**: Add a product-level definition stating that all safety-critical, structural, or deterministic maintenance tasks are executed natively by the controller, and that the maintenance executor is strictly reserved for qualitative cognitive tasks.
4. **Token/AST-Awareness for Wording Scanner**: Specify that the wording/docs scanner must use token-based/AST analysis for code and context-sensitive filters for prose, rather than a naive substring grep.

---

### Questions:
- If an operator overrides the default `dog` executor name in `city.toml`, does the `gc doctor` command validate that the new executor name is bound to a valid active provider, or does it only check if the binding itself is declared?
- Will the AST-based absence scanner also scan Core-owned TOML files to ensure no literal `"dog"` strings appear in the `assignee` or routing fields?
