# Elias Vega - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design preserves `contract = "graph.v2"` as a deprecated alias and explicitly keeps first-party graph formulas dual-declared until old binaries and the bd shell-out path are no longer compatibility hazards.
- Contract/requirements disagreement is specified as a hard validation error with deterministic diagnostic ordering, so conflicting declarations should not silently compile.
- Alias removal is framed as a release-gated decision owned by the release captain, not as a calendar cutoff, and it includes a concrete first-party scan command plus stale-docs protection.

**Critical risks:**
- [Major] External pinned-pack compatibility is still too checklist-shaped to be enforceable. The criteria say externally pinned legacy formulas must remain supported by either the alias or a documented compatibility branch, but the design does not define that branch, the supported binary/pack matrix, its support duration, or an executable test/report that proves a new release will not strand users with SHA-pinned `contract = "graph.v2"` formulas. Because global external usage cannot be measured directly, the design needs a precise policy for what "supported" means after alias removal.
- [Major] Deprecation warning reach is underspecified for `GC_NATIVE_FORMULA=false` and the bd shell-out path. The design says callers should consume `CompileWithResult`, but the compatibility bridge still allows a path where the old bd compiler may accept `contract` without producing `formula.contract_deprecated`. If that fallback can run during the alias window, users on that path may never see the warning signal used to justify removal readiness.
- [Minor] `gc formula validate --all-packs --legacy-contract-report` has clear first-party exit semantics, but external/imported legacy-only formulas appear to be informational items. That is acceptable for first-party conversion, but alias removal needs a separate external policy signal so release decisions do not accidentally treat external breakage as non-blocking by omission.

**Missing evidence:**
- The exact scan scope of `--all-packs --legacy-contract-report` for imported Git packs, local path packs, dirty local packs, and externally pinned SHAs.
- Whether the bd fallback path runs the new native requirement preflight and projects diagnostics before shelling out.
- The release artifact or checklist location that records the minimum binary floor, compatibility branch, and release-captain approval.
- Test cases for new binary plus external SHA-pinned legacy-only formula during the alias window, and for the documented post-removal behavior.
- How users of legacy external formulas are guaranteed to receive `formula.contract_deprecated` before alias support is removed.

**Required changes:**
- Add a warning-reach requirement for fallback execution: when `GC_NATIVE_FORMULA=false` or bd shell-out is used, either run `CompileWithResult` first and project `formula.contract_deprecated` consistently, or declare alias removal blocked until that fallback path is gone. Add tests for warning projection on the fallback path.
- Strengthen the alias-removal gate with an executable external compatibility contract. Define the supported compatibility branch or binary line, support duration, release artifact path, and what operators should do when a SHA-pinned legacy formula is found after alias removal.
- Extend the legacy-contract report schema or release gate with external aggregate counts such as `external_legacy_only`, `external_dual_declared`, and an explicit `alias_removal_blocked` or equivalent decision field.
- Add a compatibility matrix row for "new binary after alias removal + external legacy-only formula" and state whether that is supported by mainline alias, a compatibility branch, or an intentional hard failure with migration instructions.

**Questions:**
- Is the "documented compatibility branch" a Gas City binary branch, a dual-declared pack branch, or guidance to pin an older Gas City release?
- Should `GC_NATIVE_FORMULA=false` still perform native validation solely for diagnostics and provenance before delegating instantiation to bd?
- Who is the named release captain for this migration, and where is that approval recorded?
- What support promise does Gas City make for external SHA-pinned `contract = "graph.v2"` formulas after first-party formulas become requires-only?
