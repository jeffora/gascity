# Yuki Patel - Caller Integration Inventory

**Persona verdict:** block

**Review target:** `.gc/design-reviews/ga-j195cx/attempt-102/design-before.md`

**Scope:** caller rewrite, bd/native path alignment, in-flight molecule behavior, and dead-code policy.

## Findings

### [Blocker] Mandatory caller-manifest input includes ignored runtime state

The design makes the caller manifest repository-wide and says the explicit scan inputs include `.gc/system/packs`; it also says omitting a tree is a schema error. That cannot be a stable implementation gate in this repository: `.gc/` is ignored by `.gitignore`, and `git ls-files .gc/system/packs` returns no tracked files. In a clean CI checkout, a conforming generator would either fail because the mandatory tree is absent or silently lose the materialized pack prompts/formulas that the no-bypass gate is supposed to police.

This is not theoretical for this persona. The local materialized `.gc/system/packs/gastown/agents/*/prompt.template.md` files contain `gc bd mol wisp` / `gc bd mol burn` producer instructions, and the design correctly treats those as first-class producer candidates. But a release gate cannot depend on ignored local runtime state unless the design specifies how that state is deterministically generated and snapshotted from tracked sources.

Required change: remove `.gc/system/packs` as a mandatory CI scan input, or define a deterministic fixture-generation step that materializes it from tracked assets in a temp city and records a checked-in manifest artifact. The normative scan should name tracked sources such as `internal/bootstrap/packs`, examples, docs, and a packman/lockfile-proven external workflow-pack source. Runtime `.gc` scans can be optional release/operator evidence, not a schema-required repository input.

### [Major] Fanout parent acceptance is ambiguous after host downgrade

The in-flight section says an artifact-stamped same-identity graph root may continue graph-specific mutations after `[daemon] formula_v2` is disabled by validating the persisted accepted artifact. Later, the fanout preflight section says a fanout operation first "compiles and accepts the parent and every selected fragment" using the same host-capability snapshot. Those are different contracts.

If the host has been downgraded, recompiling the parent would fail and block a root that the in-flight rules otherwise allow to continue. If the parent source has changed, recompiling can also select a different identity than the persisted root artifact. The design is clear that new or changed formulas must compile against the current host; it is not clear whether a fanout parent is recompiled or loaded from the root artifact, nor whether inline fragments captured by the parent artifact differ from separately resolved fragment formulas.

Required change: state the fanout parent rule in one place. For same-identity existing roots, load and validate the persisted parent artifact; compile/accept only new or changed fragment formulas against the current host snapshot. If every fragment always recompiles, say that explicitly and include the host-disabled behavior. Add a zero-write fixture for an artifact-stamped root with pending fanout after `formula_v2` is disabled.

### [Major] The allowlist example undercuts the stated dead-code policy

The design says allowlist entries are narrow and that package-level or directory-level allowlists are invalid because they can hide a second raw consumer in the same file. The sample allowlist then uses `file: internal/sourceworkflow/*.go`. If that wildcard is accepted as a file scope, it weakens the raw-consumer guard exactly where the shared workflow-root predicate is supposed to be the only remaining compatibility owner.

Required change: make the sample match the policy. Use exact files plus `line_pattern`/symbol entries, or explicitly specify that file globs are allowed only when every match is further constrained by an exact symbol and the scanner fails on any raw read outside that symbol body.

## What Is Solid

The design now covers the right surfaces: per-occurrence caller rows, zero-write fixtures for durable boundaries, typed `CompileResult`/`AcceptedCompileArtifact` handoff, bd probe demotion, API/dashboard projection parity, convergence projection, active-root repair, and explicit helper expiry. Once the runtime-state scan and fanout-parent ambiguity are fixed, this persona's caller-integration concerns are mostly addressed.
