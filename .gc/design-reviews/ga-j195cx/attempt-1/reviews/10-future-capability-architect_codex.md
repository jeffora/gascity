# Ibrahim Park - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The v0 grammar is deliberately closed: omitted, `>=1`, and `>=2` are the only accepted public forms, while `>=3` and larger monotonic minima are recognized as unsupported future capability requests rather than parsed optimistically or treated as generic semver.
- Unknown `[requires]` axes fail closed with a named diagnostic, and the design reserves a future-axis shape: flat scalar strings, byte-exact lowercase ASCII axis ids, one owner package, one accepted grammar, host-capability fields, diagnostics, provenance, docs, matrix rows, old-reader fixtures, and metadata/artifact schema bumps when persisted behavior changes.
- The default-capability rules are unusually explicit: omitted `[requires]`, empty `[requires]`, and `formula_compiler = ">=1"` normalize to compiler capability `1`, share behavioral identity and artifact deduplication, and differ only in diagnostic/provenance display.

**Critical risks:**
- [Major] The canonical v0 API sketch still models `NormalizedRequirements` with singular `source`, `sourcePath`, `Source()`, and `SourcePath()` fields, while the forward-compatibility section later requires each future axis to record source path/key/value independently. If implementers treat the v0 shape as the durable internal contract, the first second-axis change can either lose per-axis provenance or require a broad authority-shape refactor at exactly the migration point this design is trying to make boring.
- [Minor] The design intentionally rejects user-defined namespaces and raw maps in v0, which is the right minimal abstraction, but it makes the future-axis checklist the only escape hatch. That checklist is comprehensive in prose; it should be treated as a required implementation template, not a release-note reminder, or a future axis such as `state_store` could land with parser support before docs, old-reader fixtures, and metadata schema handling are complete.
- [Minor] Current context docs such as `docs/reference/formula.md` still teach legacy `version` and do not yet expose `[requires]`. The design has a hard docs gate, so this is not a design blocker, but user-visible diagnostics must remain blocked until that gate is executable and passing.

**Missing evidence:**
- Literal generated count locks for the grammar, raw-shape, future-boundary, and caller-preflight suites are still represented as design placeholders in some snippets. The design says the matrix seed PR must replace them, but the review cannot yet verify the actual row counts.
- There is no checked example yet showing the second-axis addition process end to end: source parse, unknown-axis old-reader behavior, normalized typed state, host capability, accepted artifact metadata, API/dashboard projection, and release checklist output. The worked `state_store` example is clear but illustrative.
- The docs/reference rewrite, stale-guidance gate, and first-party inventory are specified but not present in the context files.

**Required changes:**
- Clarify in the canonical compile-result/API section that the singular `NormalizedRequirements` source fields and accessors are v0-only shorthand for the single `formula_compiler` axis. State that adding any second requirement axis must add typed per-axis requirement/provenance accessors and must not reuse the single `Source()`/`SourcePath()` pair as generic requirement authority.
- Make the future-axis checklist an explicit CI/release gate artifact, tied to the same owner and phase as the parser matrix, so a future axis cannot ship parser acceptance without old-reader fixtures, persisted metadata schema decisions, generated docs, and diagnostic projections.

**Questions:**
- When a future axis is added, should the v0 `RequirementSource` enum be split into per-axis provenance objects immediately, or should `formula_compiler` keep the legacy singular accessor as a compatibility projection over a new internal per-axis shape?
- Should the old-reader fixture matrix include one case where the source contains a syntactically valid unknown axis plus no `formula_compiler`, to prove the binary does not default to capability `1` after emitting `formula.requirement_unknown_axis`?
