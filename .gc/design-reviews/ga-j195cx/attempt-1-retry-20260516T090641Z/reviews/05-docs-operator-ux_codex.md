# Sienna Operator UX - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The live design at `engdocs/design/formula-compiler-requirements.md` now treats docs and terminology as release-blocking artifacts, not cleanup work. It explicitly requires `docs/reference/formula.md`, generated CLI help, config/schema docs, architecture docs, PackV2 author docs, dashboard generated types, examples, and first-party inventory updates in the same branch that exposes user-visible diagnostics.
- The operator-facing distinctions are much clearer than the short attempt snapshot: formula `[requires]`, pack `requires_gc`, host `[daemon] formula_v2`, pack refs/SHAs, formula `version`, and legacy `contract` have a required glossary and a "which key do I edit" table.
- The compatibility and rollback story avoids the dangerous operator workaround of treating `GC_NATIVE_FORMULA=false` as a production fallback. The design instead requires dual-declared formulas, typed diagnostics, minimum binary floor evidence, and alias-window reports before source conversion or alias removal.

**Critical risks:**
- [Major] The review artifact path and persona contract are stale relative to the live design. This bead's contract points at `/.gc/bugs//runs//design-review/bugfix-plan.md` and asks default-rig documentation questions, while the current design artifact is the formula compiler requirements document. The live design already says review automation must fail closed if it snapshots the wrong design path, but this attempt is demonstrating the exact class of mismatch. Required change: make the design-review gate include an executable check that the prepared attempt directory, persona inventory, and raw review task descriptions all name the same canonical design path, source hash, and persona set before any review fanout starts.
- [Major] Existing-city/operator migration guidance is implied by release artifacts but not yet shaped as a direct operator runbook. The docs skeleton tells formula authors, pack authors, and operators which keys to edit, but operators also need a concrete "I have an existing city or workflow root that was created during the alias window" path: validate/report command, expected diagnostic, whether existing accepted artifacts are rewritten or only future compiles change, and what to do if `[daemon] formula_v2` is disabled.
- [Minor] Workaround language is mostly correct, but the design should require one canonical troubleshooting section for disabled host capability and stale `contract` sources. Without that, CLI, dashboard, release notes, and PackV2 docs can each be individually accurate while still leaving operators to stitch together the safe path from separate tables.

**Missing evidence:**
- I did not find a current raw review-task artifact that matches the live persona inventory for this attempt. The prepared `gc.output_json` in the normalized attempt directory contains `05-caller-integration-inventory`, while this bead and its synthesis task still expect `05-docs-operator-ux`.
- The live design requires `docs/reference/testdata/formula-requirements-doctest.yaml`, generated first-party inventory, stale-guidance checks, and generated-help refreshes, but this review only saw the design requirements, not evidence that those fixtures already exist.
- The design names active-root repair and release evidence, but the operator-facing docs acceptance criteria should explicitly say whether existing workflow roots are migrated, dual-stamped in place, left as historical records, or repaired only by an explicit command.

**Required changes:**
- Add a release-gating check for design-review input coherence: source design path, source hash, attempt directory, persona list, and per-review task descriptions must agree, and a mismatch must hard-fail before model review tasks are routed.
- Add an existing-city migration note to the required `docs/reference/formula.md` skeleton and generated CLI/help acceptance criteria. It should cover legacy `contract` formulas, dual-declared formulas, requires-only formulas, disabled `[daemon] formula_v2`, and preexisting workflow roots/accepted artifacts.
- Add a single operator troubleshooting section with copy-paste-safe remedies for `formula.compiler_requirement_unsatisfied`, `formula.contract_deprecated`, `formula.version_misuse`, and alias-window pack compatibility, with links or exact command names for validation/report artifacts.

**Questions:**
- Is the current workflow supposed to review the live `engdocs/design/formula-compiler-requirements.md` file or the copied attempt snapshot? If the snapshot is authoritative, it is missing the docs gates that make this persona approvable.
- Should design-review fanout fail when a generated persona slug from a prior run is routed against a different current `gc.output_json`, or should the synthesis step treat it as a malformed artifact?
