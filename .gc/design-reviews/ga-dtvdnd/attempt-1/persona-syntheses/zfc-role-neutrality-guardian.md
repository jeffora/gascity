# Alistair Sterling

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Strength] All reviewers agree the requirements are directionally role-neutral: Core behavior is required SDK infrastructure, Gastown roles remain external pack configuration, and normal SDK operation must not depend on any Gastown agent existing.
- [Strength] AC2, AC8, and AC9 encode the correct ZFC guardrails: controller-only infrastructure operation, denied-token scans across source/generated/rendered outputs, and configurable executor bindings rather than Go-side role exceptions.
- [Major] AC8 and AC9 need tighter reconciliation for the one permitted `dog` token. Claude identified ambiguity between inert pack-data keys and prohibited formula bindings; this is the highest requirements-level risk because it can either over-allow real role leaks or false-positive on the legitimate configured default.
- [Major] The controller-only proof surface is under-enumerated. The requirements assert SDK self-sufficiency, but they should name the operations that must pass without a configured executor or agent, including gate evaluation, health patrol, bead lifecycle, dispatch, and order dispatch.
- [Major] The denied-token inventory must be complete by construction, not hand-authored. It should derive from or be cross-checked against the authoritative Gastown role/identity inventory in the behavior ledger/manifest.
- [Major] DeepSeek identified implementation-plan risks outside the requirements text: a hardcoded `GC_CORE_MAINTENANCE_WORKER` environment override, precedence that can make environment overrides dead code below mandatory Core defaults, split-authority fallback risk during the Core folder move, and under-specified data-driven binding optionality.
- [Minor] Scan scope needs a positive/negative control pair for materialized external Gastown content: Gastown role names should be allowed in external pack output while denied in Core-owned or Gas-City-generated output.

**Disagreements:**
- Codex saw no required changes before requirements approval and treated AC8 allowlist risks as implementation-plan concerns. Claude and DeepSeek treated the same areas as concrete risks that should be tightened or carried forward. My assessment: the requirements can proceed, but the risk is material enough to keep `approve-with-risks`.
- Claude focused on requirements precision; DeepSeek additionally reviewed the implementation plan and found ZFC leaks there. My assessment: separate them during apply. Requirements edits should clarify AC8/AC9 and proof scope; implementation-plan edits must remove hardcoded role-specific override names and fix binding precedence.

**Missing evidence:**
- The exact permitted representation for a default `dog` binding and matching positive/negative scan controls.
- The enumerated SDK infrastructure operations covered by controller-only/no-executor tests.
- Proof that denied tokens are generated from or checked against the full Gastown role/identity inventory.
- A scan-root rule that distinguishes Core-owned outputs from legitimately materialized external Gastown pack content.
- Implementation-plan proof that binding override precedence and binding optionality are data-driven without Go-side role special casing.

**Required changes:**
- Reconcile AC8 and AC9 around a single declared pack-data representation for the default executor binding, with all routes, formula bindings, prompt defaults, overlays, generated defaults, and Go fallbacks denied.
- Enumerate the controller-only SDK operation set in AC2/AC9 and require no-executor tests for that set.
- Require the denied-token set to derive from or be checked against the authoritative Gastown role and identity inventory.
- Scope rendered/generated/materialized scan roots so external Gastown pack role names are allowed in external pack content but denied in Core-owned output.
- Update the implementation plan to replace any hardcoded role-specific env var such as `GC_CORE_MAINTENANCE_WORKER` with generic data-driven binding overrides, fix override precedence, prevent legacy Core fallback during the folder move, and declare binding optionality in configuration data.
