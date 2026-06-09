# Priya Zielinski

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Positive] Both reviewers agree the design has the right parser boundary posture: the v0 grammar is closed around omitted, `">=1"`, and `">=2"`, rejects best-effort parsing, and explicitly calls out signed, zero, leading-zero, decimal, whitespace, overflow, Unicode, control-byte, empty, and unsupported-future forms.
- [Positive] Both reviewers agree the matrix-based validation strategy is appropriate. Count locks, generated rows, golden diagnostics, parser-boundary fixtures, caller preflight, contribution traversal, and registry-driven v2 construct coverage are the right mechanisms to prevent hand-picked parser tests.
- [Positive] Both reviewers agree v2-only construct enforcement is directionally sound because the design ties enforcement to a registry, source contributions, caller paths, and zero-write fixtures rather than prose-only expectations.
- [Major] The caller-preflight `count_lock: 18` is not derivable from the design text. Claude reports that the listed caller paths and durable-write boundaries do not explain the number, and Codex independently notes that literal suite counts and generated fixtures are not yet present evidence.
- [Major] Several parser raw-shape and diagnostic-count edges still need fixture-level precision. Claude specifically flags top-level `requires` as a non-table, combined-defect `diagnostic_count`, duplicate scalar keys, array-of-tables, and JSON terminology; Codex also requires golden diagnostics for every named parser edge before implementation.
- [Minor] The introductory requirement wording and provenance model need tightening. Codex flags that `>=1` is accepted later but absent from the early mapping table, and that empty `[requires]` needs a named provenance/display field so runtime behavior remains equivalent to omission.

**Disagreements:**
- Claude verdict is `approve-with-risks`; Codex verdict is `approve`. Assessment: choose `approve-with-risks` because Claude's major findings identify matrix-audit gaps that can let coverage drift, even though neither reviewer found a design-blocking behavioral hole.
- Claude treats `caller-preflight count_lock: 18`, top-level non-table `requires`, combined-defect counts, and construct-registry cross-walk as required design changes. Codex does not make those design-blocking and asks mainly for implementation-time fixtures. Assessment: require the clarifications in this lane before relying on the matrix contract, but do not fail the persona because the overall design shape is sound.
- Claude expects explicit JSON/TOML shape vocabulary corrections, including avoiding TOML "dotted table" terminology for JSON. Codex focuses on identifying the active JSON loader owner/status. Assessment: both should be resolved together by naming the JSON reader path or deleting JSON support, then expressing JSON fixtures in JSON-native terms.

**Missing evidence:**
- No Kimi 2.6 review artifact is present for this persona in the attempt's `reviews/` directory.
- Generated matrix artifacts, literal count-lock outputs, generated Go test cases, and golden diagnostic files are not present in the review artifact, so exact row counts and diagnostic projections cannot be verified yet.
- The derivation for `caller-preflight count_lock: 18` is missing.
- The accepted/rejected matrix does not explicitly pin top-level `requires` as a non-table, `[[requires]]`, duplicate scalar keys, duplicate tables, dotted/nested table conflicts, inline tables, arrays, invalid scalar types, or combined-defect diagnostic counts.
- The construct registry has one apparent identity/count mismatch unless expansion/aspect contribution is intentionally represented through `contribution_path`.
- The active JSON formula loader owner/status and JSON source-attribution behavior are not demonstrated.
- The field or vocabulary that preserves "empty `[requires]` was supplied" for diagnostics/display while keeping runtime behavior identical to omission is not named.

**Required changes:**
- Add a derivation note or explicit enumerated row list for `caller-preflight count_lock: 18`.
- Add matrix rows for top-level `requires` as invalid non-table shapes, including strings, integers, floats, booleans, arrays, array-of-tables, inline tables, duplicate tables, duplicate scalar keys, and dotted/nested table conflicts where applicable.
- Add `diagnostic_count` to the combined-defect precedence table, or require fixture rows for every combined-defect case that pin count and ordering.
- Cross-walk the v2-only construct registry against the `construct_identity` dimension and either add an expansion/aspect contribution identity or document that it is represented through `contribution_path`.
- Clarify JSON formula support by naming the active loader and owner, or remove JSON rows with the loader; express JSON edge cases in JSON-native terms with pointer/source attribution expectations.
- Align introductory wording with the executable grammar by mentioning accepted `>=1`, and name the provenance/display field for an explicitly empty `[requires]`.
