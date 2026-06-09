# Apply Summary

Verdict applied: `iterate`

Global synthesis verdict was `block`, so this attempt addressed all blocker
and major finding classes and leaves the design-review workflow in iteration.

Changes made to `engdocs/design/formula-compiler-requirements.md`:

- Replaced the raw source-workflow query sketch with typed
  `WorkflowRootCriteria`, `WorkflowRootSnapshot`, `ClassifyWorkflowRoot`, and
  `ListWorkflowRoots` contracts, plus Phase 0 parity fixtures for canonical,
  dual-stamped, legacy, closed, source-scoped, whitespace, and case-variant
  stores.
- Collapsed the convergence metadata boundary into a thin
  `ConvergenceFormulaView` over compiler-owned projections, with generated
  field-equivalence checks and guards against branching on requirement source,
  host capability, provenance, artifact refs, or root metadata.
- Strengthened durable-writer migration requirements with per-occurrence
  manifest rows, rollback controls, runner-based `bd` probe coverage, and
  zero-write assertions across roots, children, hooks, convoys, tracking,
  retry metadata, fanout state, convergence state, fired-order metadata, and
  artifact refs.
- Added named raw scanner, TOML/JSON duplicate, fuzz/property, source-position,
  and literal count-lock contracts for the requirement validation matrix.
- Defined producer-owned `FormulaDiagnosticGroupState` persistence, upsert and
  clear APIs, atomic state/event ordering, restart behavior, cleanup behavior,
  typed event payload constructors, and write-failure zero-write semantics.
- Added packman schema 2 as a hard prerequisite for release gates that rely on
  durable content hash, binding identity, transitive import identity, or
  lockfile-stored `requires_gc` evidence.
- Replaced the broad stale-guidance `version` matcher with parsed
  formula-context checks and scoped formula-version prose checks, and added
  the local `make formula-docs-check` report contract.
- Added phase owner/PR-home/local-command/rollback controls, and reinforced
  source/provenance fields as non-behavioral authority with canonical axis
  manifest encoding and default-capability equivalence fixtures.

Unfixable items: none.
