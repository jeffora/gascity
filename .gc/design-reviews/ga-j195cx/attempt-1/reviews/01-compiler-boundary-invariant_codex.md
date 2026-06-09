# Nadia Sorenson - Codex

**Verdict:** approve

**Top strengths:**
- The design states the core boundary directly: `[requires]` declares a minimum compiler capability, the active Gas City binary decides whether it can satisfy that requirement, and formulas do not select a compiler implementation at runtime.
- Requirement parsing, legacy `contract` normalization, v2-only construct detection, host-capability checks, accepted artifact minting, and write-intent validation are all owned by `internal/formula`, with durable callers required to consume `CompileResult`, `AcceptedCompileArtifact`, or typed projections instead of raw TOML or root metadata.
- Caller drift is addressed as a first-class invariant: the design names the durable writer surfaces, the shared `internal/sourceworkflow` workflow-root predicate, explicit `HostCapabilities` plumbing, and static guards against bare `Compile`, `*Recipe`, `Recipe.GraphWorkflow`, direct host globals, and convergence subset parsers.

**Critical risks:**
- [Minor] The compiler-boundary design is now strong enough, but its safety depends on the generated manifests and static guards being treated as hard CI gates. If those checks land as advisory reports, CLI, API, convergence, order, or fanout callers could slowly reintroduce raw `contract` or root-metadata authority despite the prose contract.
- [Minor] The document allows preview APIs such as `CompileWithResult`, `ProjectFormula`, and legacy `Compile` to remain for display or compatibility paths. The design compensates with accepted-artifact and write-intent requirements, but the implementation must make preview-vs-durable types mechanically hard to confuse.

**Missing evidence:**
- The document specifies checked-in caller manifests, construct registries, projection equivalence fixtures, and zero-write fixture suites, but this review did not inspect implemented generated artifacts proving every current caller is represented.
- The document names `internal/sourceworkflow.ListWorkflowRoots` as the only workflow-root query boundary; the implementation still needs a scan/guard proving no CLI, API, order, convoy, dashboard, or convergence path keeps a parallel `beads.ListQuery` or metadata predicate.
- The design says `HostCapabilities` is explicit at every compile entry point; final implementation evidence should include concurrent host-capability tests proving no package-global formula-v2 flag can influence compile decisions.

**Required changes:**
- None before approval for this persona lane. The document satisfies the active-binary ownership, normalized-requirement authority, and typed-consumer boundaries.

**Questions:**
- Will the static guards run in the fast required gate (`make test` or equivalent CI), not only in an optional release-validation command?
- Will preview-only APIs use distinct accepted/preview types at function signatures so a durable writer cannot compile with preview data without a test-only or unsafe escape hatch?
