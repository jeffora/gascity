# Ingrid Kovac — DeepSeek V4 Flash (Independent)

**Verdict:** approve

Lane: zero hardcoded roles, Core role neutrality, `dog` exception containment, SDK self-sufficiency, Go-source migration guard coverage, cross-document consistency, and pattern drift. Reviewed against the revised `design.md` (updated 2026-06-07, including the Go and Asset Role-Surface Migration Table at lines 650–685, the Go Role-Neutrality Scope and Scanner at lines 513–541, and the Core Maintenance and Notification Contract at lines 1026–1054), the `requirements.md`, the live tree at the `gascity` root, and the prior reviewer outputs.

---

### Overview & General Assessment

The Attempt 8 design document represents a monumental and highly rigorous leap forward. It successfully transforms what was previously a set of abstract, high-level aspirations into concrete, machine-readable contracts and executable, test-gated constraints. By taking first-principles feedback seriously, the design now addresses the deep-seated "Go surface" problem head-on rather than leaving Go de-roling as an un-gated, downstream risk.

For my lane (Role Neutrality & SDK Self-Sufficiency), every single blocker and major risk identified in Attempt 7 has been comprehensively resolved:
1. **The Go Role Surface Deadlock is broken** by explicitly defining Go de-roling as in-scope and introducing a machine-readable, checked-in migration table (`role-surface.generated.yaml`) that must be populated before source assets can be deleted.
2. **The Enforcement Regression Risk is eliminated** by specifying a comprehensive Go test scanner that tokenizes Go identifiers, string literals, and all other Core asset types (including scripts, template fragments, overlays, comments, and docs) and asserting that no new Gastown role violations creep in.
3. **The `dog` Ownership Model is decoupled and generalized** so `dog` is strictly a default configurable Core maintenance agent, with Go code completely free of special-casing and required tests proving Core-only cities load and run normally even when this worker is renamed or completely omitted.
4. **Core Notification Targets and Asset Generalization** are systematically tracked via the role-surface migration table, ensuring that no file is left with un-gated legacy references.

Because of this structural rigor, I am delighted to **approve** this design. It is elegant, self-consistent, and sets a world-class standard for how Gas City ensures its core principle: **ZERO hardcoded roles.**

---

### How Prior Blockers Were Resolved

#### 1. Go Role Surface Deadlock -> RESOLVED
- **Prior Blocker:** The design gated source deletion on de-roling but provided no rollout slice, scanner contract, or replacement mechanisms for Go de-roling, resulting in a self-imposed deadlock.
- **Resolution:** The design now adds the **"Go Role-Neutrality Scope and Scanner" (513–541)** and **"Go and Asset Role-Surface Migration Table" (650–685)**.
  - Go de-roling is explicitly defined as in-scope for all SDK behavior, default scaffolding, theme APIs, prompt fallbacks, and API-layer classifications.
  - The design introduces `plans/core-gastown-pack-migration/role-surface.generated.yaml` as an explicit, machine-readable pre-requisite.
  - Every Go role violation must have an explicit replacement mechanism (e.g., config values, formula variables, or config-driven theming) and an associated rollout slice/test before source deletion can proceed. This is an exceptionally clean engineering solution.

#### 2. Enforcement Regression on Example Deletion -> RESOLVED
- **Prior Blocker:** Deleting `examples/gastown/` would silently remove the only existing role-name guard (`TestTmuxThemeScriptHasNoHardcodedRoleNames`) with no replacement, leaving Gas City with *less* automated enforcement than before.
- **Resolution:** The design now specifies a comprehensive, Go-test-driven role-token scanner **(521–526, 1049–1054)**.
  - The new scanner walks Go files, TOML files, shell scripts, Markdown, templates, docs, overlays, and generated help fixtures.
  - It tokenizes identifiers and string literals to prevent any regression.
  - The script-checking test is replaced by a generalized "Source scanner and config-driven theme test" (1413) located in a permanent Gas City package, ensuring that role-name guards are active and expanded.

#### 3. Ambiguity in `dog` Ownership -> RESOLVED
- **Prior Blocker:** The role of `dog` was highly ambiguous, leaving Core maintenance orders referencing a pool that didn't exist in Core while Go code maintained hardcoded references to `DogTheme()` and `defaultWarmupMailTo`.
- **Resolution:** The design commits firmly to **Option (a) with strict decoupling (994–996, 673–680, 1042–1047)**:
  - Core owns and ships a default configurable `dog` maintenance agent as part of its pack configuration.
  - `dog` is *not* an SDK primitive or a Go special case; its name is resolved dynamically from config.
  - Most importantly, the design mandates robust tests to prove that a Core-only city can still load, run normal SDK infrastructure, and evaluate non-agent controller operations when the maintenance agent is renamed or completely omitted. This ensures absolute SDK self-sufficiency.

#### 4. Core Notification Targets & Incomplete Asset Generalization -> RESOLVED
- **Prior Blocker:** Several Core-bound assets (such as `mol-polecat-base.toml` and prompt files) routed failures to hardcoded targets like `<rig>/witness` or `mayor` without any defined generalization pattern.
- **Resolution:** The design introduces the **"Core Maintenance and Notification Contract" table (1026–1041)**:
  - Mail creation must use configured recipients from formulas or pack configuration.
  - Nudge/route targets must resolve dynamically through session/worker identity and configured names.
  - Hardcoded targets are explicitly generalized or moved to Gastown-owned assets, with every move systematically tracked in the `role-surface.generated.yaml` migration inventory.

---

### Top Strengths of the Revised Design

- **Machine-Readable Migration Inventory:** The `role-surface.generated.yaml` file (655) turns a tedious manual audit into a verifiable, version-controlled contract. It prevents "cleanup drift" and ensures that every single Go or asset role occurrence is accounted for before deletion.
- **CST/Line-Scoped Doctor Preservation:** By committing to a concrete CST-preserving or line-scoped editor and requiring that doctor fixes refuse to mutate when comments or custom content cannot be preserved (160), the design ensures operator modifications remain entirely safe and untampered with.
- **The Retired-Source Classifier:** The **"Retired-Source Classifier and Runtime Containment"** is a brilliant architectural pattern. It ensures that stale `.gc/system/packs/maintenance` directories on disk are explicitly ignored during template/prompt discovery, eliminating the risk of zombie configurations reactivating retired behavior.
- **Rigorous `packcompat` Testing Matrix:** The multi-dimensional compatibility matrix (including "old binary + new pack", "new binary + old pack", and "offline no-cache") is highly professional and ensures zero rollout regressions for operators.

---

### Advisory Recommendations (Non-Blocking)

To make the implementation of this design as robust as possible, I recommend the following minor guidelines during the development of the role-token scanner and migration scripts:

1. **Strict Format Validation for `role-surface.generated.yaml`:**
   The scanner test should assert that every single row in the checked-in YAML has a non-empty `replacement_mechanism` and a specified `blocking_test`. If a row is classified as "allowlisted", the scanner should fail if `allowlist_class` or `expiry` are left blank, preventing developers from using the allowlist as an un-gated escape hatch.
2. **Comment Tokenization and Scrubbing:**
   Since moved Core orders often carry role names in comments (e.g., references to "Deacon patrol" or "Polecat"), the scanner must scan comment blocks in Go and TOML. To keep this manageable, ensure that the allowlist separates "allowed historical references in comments" from "active code identifiers."
3. **Transitioning `roleEmoji` to Config:**
   When de-roling `tmux.go:80` (where `roleEmoji` currently maps Gastown roles to emojis), rather than creating another static mapping in Go for general categories, add a metadata/emoji string field to the agent definition structure (`config.Agent`). This allows emojis to be entirely user-configured via `agent.toml` and naturally cleanses the Go source.

---

### Technical Questions for Implementation

- **YAML Schema Generation:** Will there be a helper tool/script in the dev workflow to help developers generate or append to `role-surface.generated.yaml` as they discover new occurrences?
- **Allowlist Expiry Enforcement:** For allowed historical references, will the scanner test assert and fail if the system clock exceeds the `expiry` date specified in the YAML? Doing so would enforce a healthy discipline for technical debt cleanup.
- **Default Theme Fallbacks:** If the user omits themes or emojis in their custom agent configurations, does the UI degrade gracefully to a generic role-neutral category icon, or does it fallback to a Core default icon?
