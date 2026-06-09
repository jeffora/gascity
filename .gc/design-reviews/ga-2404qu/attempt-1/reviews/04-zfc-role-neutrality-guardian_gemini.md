# Ingrid Kovac — ZFC & Role-Neutrality Guardian (Iteration 17 / Attempt 17, Independent DeepSeek V4 Flash Style Review)

**Verdict:** BLOCK

**Scope:** Zero hardcoded roles, Core role neutrality, `dog` exception containment, SDK self-sufficiency, Go-source migration guard coverage, cross-document consistency, and architectural coherence.

This independent review evaluates the Iteration 17 / Attempt 17 snapshot of the Core and Gastown Pack Split design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md` / `attempt-17/design-before.md`) against `requirements.md` and the live codebase at the `rig_root` (`/data/projects/gascity`).

---

## Executive Summary

As Ingrid Kovac, the **ZFC and Role-Neutrality Guardian**, I am issuing a strict **Verdict: BLOCK** for Iteration 17 / Attempt 17 of the Core and Gastown Pack Split design.

While the Attempt 17 design introduces exceptional architectural advancements—specifically the new **Provider-Pack Continuity and Rewrite Exceptions** ledger (§2370–2391), **Strict Preflight Recipient Validation** (§1378–1381), and **Hermetic Offline Testing** (§1495, §2009, §3108)—it fails to address the underlying structural loopholes that compromise the Zero Framework Cognition (ZFC) and role-neutrality invariants.

The design still exhibits four critical, unresolved structural flaws that require a blocking verdict before any code modification or source deletion can land:
1. **The Scanner Sub-Identifier Blindspot remains unaddressed (§581–586):** There is no explicit requirement for the scanner to split camelCase or PascalCase identifiers. Standard tokenization treat identifiers like `MayorTheme` as a single opaque token. This guarantees that Go-source APIs returning hardcoded role names will bypass the role guard and pass the deletion gate.
2. **The Dashboard `crew` Wire Vocabulary Contradiction persists (§1929, §2161 vs §3075):** The design claims both that the scanner covers API/dashboard projections of `agent_kind` and `crew` (§1929) and that "This migration should not require dashboard changes" (§3075). Changing the wire-level JSON representation of `agent_kind` or `crew` without dashboard updates is physically and logically impossible.
3. **Core Asset Classification is still missing (§237–315):** The `Existing Asset Migration Map` fails to categorize Core maintenance assets into `controller_owned` (required SDK infrastructure) or `optional_core_maintenance_worker` (worker-bound). Without this explicit mapping, required SDK controller logic can easily be assigned to the optional maintenance worker, violating SDK self-sufficiency.
4. **Lack of `dog` Metadata Containment (§1922–1927):** The scanner lacks instructions to reject literal `dog` values within Core routing, pool, and dispatch metadata fields, inviting developers to bypass symbolic bindings with hardcoded fallbacks.

---

## Detailed Evaluation of the Three Key Questions

### 1. Does any Go change introduce role-conditional logic or a literal role name outside tests, migration docs, or pack configuration?
**No, but the scanner design is structurally incapable of enforcing this on pre-existing, active Go code.**
The design asserts that absolute role neutrality will govern all Go source code (§573–586). However, the following active role-name violations exist in the production tree today and will bypass the proposed scanner:
* **The Tmux Special Themes (`internal/runtime/tmux/theme.go:33,39,43`):** `MayorTheme()`, `DeaconTheme()`, and `DogTheme()` return hardcoded role names (`"mayor"`, `"deacon"`, `"dog"`).
* **The Tmux Status Format Map (`internal/runtime/tmux/tmux.go:80–89`):** `roleEmoji` explicitly maps `"mayor"`, `"deacon"`, `"witness"`, `"refinery"`, `"crew"`, and `"polecat"` to emojis.
* **Warmup Defaults (`cmd/gc/cmd_start_warmup.go:33`):** `defaultWarmupMailTo = "mayor"` is hardcoded.

The Go tokenizer treats `MayorTheme` as a single, indivisible identifier token. A scanner that simply checks if an identifier or string literal equals `"mayor"` (or matches `\bmayor\b` on a whole-word basis) will **completely fail** to detect these active Go APIs. They will slip past the scanner, and the deletion gate (§1932–1934) will pass vacuously while role-conditioned logic remains in the Go binary.

### 2. Does the Core role-name guard scan every asset type including scripts, overlays, orders, template fragments, doctor checks, metadata, and prompt snippets?
**Yes, but the scanner's matching rules are under-specified, and it lacks self-referential exemptions.**
The scanner described in §1910–1935 correctly spans all asset types, including dashboard TypeScript files and OpenAPI schemas. However, the scanner faces two critical failure modes:
* **Self-Collision on Audit Manifests:** The scanner is instructed to scan all YAML and Markdown assets and reject role names (§1914–1918). Yet, the role-surface table (`role-surface.generated.yaml` / §2153) and the behavior manifest (`behavior-manifest.generated.yaml` / §117) *must* record the literal role names (e.g. `mayor`, `deacon`) they are tracking for deprecation. Without explicit, narrow file-set exemptions, the scanner will crash on its own audit tables.
* **Gastown Conflation:** The scanner is instructed to sweep the "public Gastown checkout" (§1915). Because Gastown is a configured pack of roles, its assets legitimately contain role names. Forcing public Gastown assets to adhere to the same "zero-roles" rules as Core means every valid Gastown role reference would require an allowlist entry in Core, which violates pack isolation.

### 3. Can Core infrastructure still run when the default maintenance agent is removed or renamed by configuration?
**In theory, yes, but the design lacks a concrete classification to guarantee self-sufficiency.**
The symbolic binding model (`[gc.bindings.maintenance_worker]`) is mathematically sound (§1796–1811). However, the design states that omitting the Core maintenance worker disables "worker-bound Core maintenance orders and formulas" while ensuring "controller-owned SDK operations" still run (§1812–1818).
* **The Gap:** The design lists generic infrastructure orders such as gate sweep, blocker-close cascade, stale cleanup/reaper, spawn storm detection, and binary doctor checks (§2500–2503). However, it never classifies which assets are controller-owned versus worker-bound.
* **The Risk:** If `gate-sweep.sh` or `reaper.sh` is accidentally classified as worker-bound (or written to depend on an active worker session), omitting the maintenance worker will silently disable core SDK controller operations. A per-asset classification is an absolute prerequisite to prove Core self-sufficiency.

---

## Top Strengths of Attempt 17

* **The Provider-Pack Continuity Ledger (§2370–2391):** Resolves the prior contradiction regarding `bd` and `dolt` assets. It explicitly defines which files can be rewritten, how they map to `core.maintenance_worker` symbolic bindings, and what tests must prove their continuity.
* **Strict Preflight Recipient Validation (§1378–1381):** Enforcing that required recipient fields fail preflight if empty or `/` eliminates the risk of silent alert failures, securing Gastown's warning pathways.
* **Hermetic Offline Testing (§1495, §2009, §3108):** Mandating that `test/packcompat` can run offline using local fixtures and remote caches prevents network dependency failures in air-gapped CI environments.
* **Table-Driven Participation Resolution (§2055–2096):** Typing `RequiredSystemPackParticipation` and returning explicit records for required host packs ensures resolver-driven participation is fully verified before any behavior discovery occurs.

---

## Critical Risks & Remaining Gaps

### 1. [Blocker] Scanner Sub-Identifier Blindspot
* **The Risk:** Standard Go tokenization evaluates `MayorTheme` or `ConfigureGasTownSession` as single identifier tokens.
* **The Gap:** A whole-token role-name match (e.g., `\b(mayor|deacon|crew|dog)\b`) will **silently fail to catch** these active Go identifiers.
* **Required Change:** The scanner specification (§581–586) must explicitly require splitting camelCase/PascalCase sub-identifiers (or executing a substring search of forbidden role names against a strict identifier list) to enforce absolute role neutrality on Go APIs. Negative test fixtures must include camelCase role-bearing Go names to prove that any such API fails CI.

### 2. [Blocker] Dashboard `crew` Wire Vocabulary Contradiction
* **The Risk:** Wire-level fields like `crew` or `agent_kind` must be de-roled (§1929). Yet, the design asserts "this migration should not require dashboard changes" (§3075).
* **The Gap:** The live dashboard filters on `session.agent_kind === "crew"` (`cmd/gc/dashboard/web/src/panels/crew.ts:45`) and reads the OpenAPI schema for `agent_kind` (`docs/schema/openapi.json`). Changing the Go wire vocab without modifying the dashboard will immediately break API compatibility.
* **Required Change:** Resolve this contradiction. Either authorize dashboard API updates in the rollout plan (including running `make dashboard-check`), or formally allowlist the specific Go wire/JSON struct tags required for backward compatibility with the current dashboard version.

### 3. [Blocker] Core Asset Classification Missing
* **The Risk:** Omitting the Core maintenance worker disables worker-bound tasks while leaving controller-owned operations active.
* **The Gap:** The design lists generic infrastructure orders and scripts such as gate sweep, blocker-close cascade, stale cleanup/reaper, spawn storm detection, and binary doctor checks. However, it never classifies which assets are controller-owned versus worker-bound. Without a per-asset classification table, an implementation can easily break controller self-sufficiency by assigning a required SDK operation to the optional `core.maintenance_worker` binding.
* **Required Change:** Extend the behavior manifest or role-surface table with a mandatory per-asset classification: `controller_owned`, `optional_core_maintenance_worker`, `public_gastown`, or `retired`. Mandate negative tests showing that when `maintenance_worker = ""` is configured, all `controller_owned` operations execute successfully.

### 4. [Blocker] Lack of `dog` Metadata Containment
* **The Risk:** The default maintenance agent key `dog` must be treated purely as configuration, not as a hardcoded fallback or special-case.
* **The Gap:** The scanner is not instructed to reject `dog` literals in Core-owned routing, pool, and dispatch metadata fields. Without this constraint, developers can bypass symbolic bindings by hardcoding the literal `dog` into dispatch rules.
* **Required Change:** Mandate that the scanner rejects the token `dog` in all Core-owned routing, pool, and dispatch metadata fields, requiring symbolic bindings (like `core.maintenance_worker`) instead.

---

## Required Changes for Finalization (Actionable Gates)

To lift this block, the design document must be updated to resolve these issues with the following concrete, actionable gates:

1. **Codify Go Scanner Token Rules:** Specify that the Go role-name scanner must split camelCase/PascalCase sub-identifiers. Mandate a negative fixture per role token in camelCase form that must fail CI.
2. **Resolve Dashboard Wire Alignment:** Explicitly authorize the dashboard API and generated type updates in the rollout plan under a validated gate, or formally allowlist the exact JSON struct tags required for backward compatibility.
3. **Classify Core Assets:** Provide an exhaustive, per-asset table classifying moved/retired Maintenance assets into `controller_owned` vs. `optional_core_maintenance_worker`. Require omitted-worker negative tests for all `controller_owned` assets.
4. **Contain `dog` Literals:** Mandate that the scanner rejects the token `dog` in all Core-owned routing, pool, and dispatch metadata fields (requiring symbolic bindings instead).
5. **Self-Referential & Pack Exemptions:** Formally exempt the generated metadata audit tables (`role-surface.generated.yaml`, `behavior-manifest.generated.yaml`) from role-name rejection checks. Clarify that public Gastown assets are scanned for preservation inventory, not role-name rejection.

---

## Questions

* **Sub-identifier Tokenization:** Will the Go scanner split sub-identifiers (camelCase/PascalCase) and assert on negative camelCase role fixtures?
* **Dashboard Wire Alignment:** Will the dashboard API be updated to support the new neutral grouping metadata, or will we carry a narrow, expiring wire compatibility row for `crew`?
* **Core Classification:** Which of the generic maintenance scripts listed in lines 2500–2503 are required SDK controller-owned infrastructure versus optional worker-bound tasks?
* **Exemption Strategy:** How will the scanner distinguish between forbidden role references in active Core code versus legitimate role tracking entries in `role-surface.generated.yaml`?
