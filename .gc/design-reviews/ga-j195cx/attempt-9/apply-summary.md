# Apply Summary

Source synthesis: `.gc/design-reviews/ga-j195cx/attempt-1/synthesis.md`
Global verdict: `block`
Apply verdict: `iterate`

Updated `engdocs/design/formula-compiler-requirements.md` to address all
blocker and major findings from the synthesis:

- Added raw TOML preservation requirements and an explicit validation matrix
  for `contract`, `[requires]`, v2-only constructs, and host capability.
- Added typed host capability input, `CompileOptions`, and required
  `CompileWithResult` / `CheckRequirements` APIs with per-call capability tests.
- Added an executable call-site migration table covering formula, molecule,
  sling, orders, convoy, API, sourceworkflow, graphroute, convergence, and
  dashboard surfaces.
- Added measurable migration gates for old/new binaries, old/new packs,
  external SHA-pinned packs, mixed-version shared stores, native compiler paths,
  and `bd` shell-out compatibility.
- Added a diagnostic projection matrix, typed event payload requirements, and
  bounded `OnceKey` suppression semantics.
- Added auditable workflow-root provenance metadata and durable compile artifact
  fallback rules.
- Added glossary, docs/examples inventory, deterministic legacy `version`
  diagnostic behavior, and forward-compatibility rules for future axes.

No unfixable review item remains documented as accepted risk; the global verdict
was `block`, so the workflow should iterate after this apply step.
