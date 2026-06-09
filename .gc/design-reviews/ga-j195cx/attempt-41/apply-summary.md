# Apply Summary

Verdict applied: `block` -> `iterate`.

Updated `engdocs/design/formula-compiler-requirements.md` to address all
attempt-41 blocker and major findings:

- Added a generated workflow-control metadata registry and expanded the
  validation-matrix schema with suite ownership, coverage intent, row kinds,
  count locks, dimension ownership, and required zero-write assertions.
- Fixed the raw-consumer rollout contradiction: new production raw consumers
  become CI-blocking at phase 3a once the canonical compile result, accepted
  artifact, and shared workflow-root predicate exist; existing consumers must
  live on owned expiring allowlist rows.
- Extended caller coverage to retry, `on_complete`, Ralph continuation,
  workflow-control metadata, fanout fragments, `molecule.Instantiate`, API
  globals, and legacy graph-contract helpers.
- Made alias removal fail closed on zero first-party `legacy_only`, zero
  first-party `dual_declared`, expired external-support rows, exact old-reader
  versions/SHAs, and explicit visibility for accepted legacy-alias compiles
  across CLI/API, orders, convergence, fanout, retry, controller, dashboard, and
  release validation.
- Added a v0-readable durable root requirements contract with schema version,
  accepted artifact version, minimum reader capability, axis manifest, and
  old-reader fail-closed lifecycle rules for future axes.
- Pinned graph semantics: `requires.formula_compiler = ">=2"` means graph
  workflow topology, exposed only through compiler-owned compile results,
  accepted artifacts, typed step/runtime-var projections, and sourceworkflow
  facts.
- Tightened diagnostic projection with canonical warning/fatal keys, dual
  formula/host attribution for unsatisfied requirements, typed wire fields, and
  updated parity fixtures.
- Coupled rollout, docs, generated help/schema, OpenAPI/dashboard types, and
  rollback rules so user-visible diagnostics have their docs bundle first and
  rollback does not revive runtime `bd` or `GC_NATIVE_FORMULA=false` paths.
- Named convergence projection ownership, enumerated `ValidateProjection`
  diagnostic codes, and defined active legacy-root migration and no-write
  behavior for retry, next iteration, missing-child repair, and speculative
  wisps.
- Defined PackV2 lockfile ownership, schema, immutable `LockedRevision`
  semantics, CLI validation behavior for remote/local/direct/transitive inputs,
  transitive ref collision handling, and advisory-only `safe_automatic_edit`.
- Converted minor edge semantics into fixture-level decisions for `>=1`,
  integer boundaries, Unicode and NUL strings, test-only constructors,
  `schema` terminology, dashboard remediation fields, transitive ref
  collisions, and safe edit hints.

No unfixable blocker or major item remains documented as accepted risk.
