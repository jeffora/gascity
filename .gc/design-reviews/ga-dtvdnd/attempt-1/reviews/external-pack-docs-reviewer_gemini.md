# Simone Kaye — DeepSeek V4 Flash Perspective Independent Review (Iteration 9 / Attempt 9)

**Lane:** external-pack-docs-reviewer (wave 1) — external Gastown pack authority, registry behavior, source-tree cleanliness, and documentation consistency.

**Verdict:** approve-with-risks

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active Attempt 9 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, updated 2026-06-09) and the `gc.mayor.requirements.v1` schema. I evaluated the criteria and verified the external-pack dependency and caching hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/external-pack-docs-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/external-pack-docs-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 9 synthesis.
3. **Verdict Rationale.** The Attempt 9 requirements show exceptional progress and structural maturity. By explicitly requiring public `gascity-packs/gastown` docs and comments (AC12) to be audited, the plan successfully addresses our previous feedback regarding cross-repository terminology drift. Therefore, I award an **APPROVE-WITH-RISKS** verdict. However, from the perspective of **External Gastown Pack Authority, Registry Behavior, and Source-tree Cleanliness**, several critical bootstrapping paradoxes, remote process tracking edge cases, and rollout version-skew races remain unaddressed in the proposed requirements.

---

## Evaluation of the Three Key Questions

### 1. Does the requirement make gascity-packs/gastown the authoritative source for Gastown behavior, templates, overlays, and workflow checks?
**Reviewer Finding: Yes.**
The requirements strictly establish `gascity-packs/gastown` as the single source of authority (AC4 and AC14). The rollout plan ensures that no Gas City source deletion or Core role-generalization can land until the public Gastown repository contains all matching behavior, prompts, scripts, and tests, validated by an immutable `sha:` pin, pack digests, and behavior-preservation proofs.

### 2. Will no maintained pack source remain under examples/gastown/packs, and do docs and tutorials stop presenting retired locations as authoritative?
**Reviewer Finding: Yes.**
Under AC4 and AC5, legacy in-tree directories are removed or isolated as non-resolvable fixtures. The wording and docs scanner (AC12) actively audits all operator guides, tutorials, and CLI help text, including the public Gastown pack's own documentation, ensuring that no active documentation references retired folders or treats `examples/gastown/packs/*` as authoritative.

### 3. Are public registry behavior, template imports, and operator-facing terminology consistent across docs, CLI messages, and examples?
**Reviewer Finding: Yes.**
AC12 enforces consistent Core-required / Gastown-external / Maintenance-retired terminology across all docs, examples, CLI help, doctor, import-state, and registry/catalog/discovery surfaces. The inclusion of the public pack's own documentation files in AC12 ensures complete consistency and prevents retired terminology from slipping into the public-facing SDK surface.

---

## Critical Risks & Architectural Gaps

### 1. [Major] The Bootstrapping Paradox for Missing-Core Bootstrap Diagnostics (AC11)
* **The Risk:** AC11 requires that `gc doctor`, `gc import-state`, and `gc doctor --fix --non-interactive` have a bootstrap-only diagnostic mode that can run with no packs resolved (such as when Core is completely missing). However, JSON output is required to strictly conform to `migration-diagnostics.schema.json`. If the schemas, condition codes, or JSON rendering logic are defined in Core assets, the tool cannot load them to render schema-compliant diagnostics when Core is absent.
* **The Gap:** The requirements assume diagnostic schemas and metadata can be loaded dynamically, which is impossible if the required pack containing them is missing.
* **The Fix:** Amend AC11 to explicitly require that the bootstrap-only diagnostic schemas, condition codes, and exit-code mapping are compiled directly into the Go binary (or statically embedded), so that the CLI can render schema-compliant JSON diagnostics without loading any external pack assets.

### 2. [Major] Remote / Cloud Session Live-State Refusal (AC10)
* **The Risk:** AC10 states that "Live-state refusal evidence comes from direct runtime/process/session observation rather than stale status files" to refuse mutating repair while live sessions are running. While this works beautifully for local runtimes (by checking the process table, `ps`, or `lsof` for local tmux/subprocess wrappers), it fails for cloud or remote runtimes (such as the k8s provider). A local `gc doctor` running on an operator's workstation cannot query local process tables to discover active sessions running in remote Kubernetes pods.
* **The Gap:** The "no status files" rule is assumed to resolve all state, but local process table queries cannot observe remote session providers.
* **The Fix:** Require that for remote runtime providers (like Kubernetes), the live-state evidence must be queried directly from the provider's active API (e.g., querying active Kubernetes pod or job status) rather than assuming a local process table scan is sufficient.

### 3. [Major] Version-Skew on the Reverse Release Boundary (AC14 / AC15)
* **The Risk:** AC14 defines the "two-repository release order so Gas City never ships a public pin that lacks the validated Gastown behavior manifest." This implies releasing the public Gastown pack first, then updating Gas City's pin. However, if the public Gastown pack is updated to the new Core-neutral structure (no longer importing Maintenance) before the new Gas City binary is released, any existing legacy Gas City installation that tries to import the new public Gastown pack might fail or encounter unexpected behavior because it still expects Maintenance to be auto-included or imported.
* **The Gap:** The release order protects the new binary but can break compatibility for old binaries importing the new public pack.
* **The Fix:** Amend AC15 to require that the public Gastown pack maintains a backward-compatibility window or tag structure, or explicitly specify the diagnostic behavior when a legacy Gas City binary encounters the new decoupled public Gastown pack layout.

### 4. [Major] Non-Deterministic Resolution of Third-Party Imports (AC15 / AC16)
* **The Risk:** AC15 states that "Branch or tag refs are fetchability metadata only." However, third-party packs not owned by `gascity-packs` may still import Gastown using branch or tag names in their own `pack.toml` files. If an operator initializes a city that transitive-imports Gastown via a branch name, the resolution becomes non-deterministic because the branch ref can point to different commits over time.
* **The Gap:** The ledger pins the canonical Gas City templates, but does not prevent transitive third-party imports from introducing mutable branch/tag references that bypass the offline cache or lockfile.
* **The Fix:** Require the pack-resolution loader to reject mutable branch or tag imports for public system-recognized packs in production environments unless they are resolved to an immutable commit sha recorded in a lockfile, forcing all transitive imports to resolve deterministically.

---

## Required Changes for Finalization

1. **Statically Embed Diagnostic Schemas:** Amend AC11 to require that bootstrap-only diagnostic schemas, condition codes, and JSON templates are statically embedded or compiled into the `gc` Go binary, eliminating the dependency on external pack loading during missing-Core bootstrapping.
2. **Provider-Aware Live State Queries:** Amend AC10 to require provider-aware live state queries (such as calling the Kubernetes API for the k8s provider) to discover active remote sessions.
3. **Legacy Binary Compatibility Safeguards:** Require in AC14 or AC15 that the public Gastown pack release does not break backward compatibility with legacy Gas City binaries, or that a clear boundary (such as major versioning or explicit compatibility tags) prevents old binaries from fetching incompatible decoupled layouts.
4. **Enforce Immutable Transitive Resolution:** Amend AC15 to require that transitive imports of public packs resolve strictly to immutable commit shas, failing closed if a mutable branch or tag ref cannot be resolved deterministically.

---

## Questions

* **Registry Federation:** If an operator hosts a private registry mirror, does the public pack validation (AC14) run against the mirror, and does the mirror need to replicate the exact immutable commit pin?
* **Air-Gapped Bootstrap:** For completely air-gapped environments, will the release provide a combined offline-bundle artifact containing both the Core pack and the pinned public Gastown pack cache?
* **Local Override Policy:** If an operator needs to make local emergency modifications to the cached public Gastown pack, does the loader permit a local path-override without breaking the behavior-preservation manifest validation?
