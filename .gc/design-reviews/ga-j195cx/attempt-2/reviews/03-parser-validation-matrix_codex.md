# Priya Zielinski - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design treats the validation matrix as an executable contract, with named source files, generated tests, golden diagnostics, suite count locks, and zero-write assertions for caller paths.
- The raw requirement capture contract is much stronger than a direct decode into `Formula`; it explicitly preserves source path, key, raw value, shape, line, column, and duplicate state before typed normalization.
- The v2-only construct registry is tied to decoded field paths and generated workflow-control metadata rows, which directly addresses the risk of caller-specific bypasses.

**Critical risks:**
- [Major] TOML quoted keys are not pinned in the grammar or raw-shape matrix. TOML permits quoted keys such as `[requires] "formula_compiler" = ">=2"` and dotted quoted forms such as `requires."formula_compiler" = ">=2"`. The design says axis identifiers are byte-exact lowercase ASCII and rejects dotted/nested tables, but it does not state whether quoted spelling of the same logical key is accepted, rejected, or treated as a duplicate when mixed with the bare key. Without explicit rows, the raw scanner and `toml.MetaData` can diverge and either miss an unknown axis or accept a duplicate requirement.
- [Major] TOML string lexical forms are under-specified for a byte-exact value grammar. The document repeatedly says only byte-exact `">=1"` and `">=2"` are accepted, but TOML has basic strings, literal strings, multiline strings, and escapes. It is not clear whether `formula_compiler = '>=2'`, `formula_compiler = "\u003e=2"`, or a multiline string that decodes to `>=2` should be accepted based on decoded bytes or rejected based on raw lexical spelling. That choice affects source attribution, duplicate diagnostics, and old-reader compatibility.
- [Minor] The early decision table lists only omitted and `>=2`, while the grammar later accepts explicit `>=1`. The later rules are clear, but the summary table should include `>=1` or point to the grammar so implementers do not accidentally omit an accepted default-capability provenance case.

**Missing evidence:**
- Matrix rows for quoted TOML keys, escaped key spellings, quoted-key duplicates, and quoted unknown axes.
- Matrix rows for TOML basic, literal, multiline, and escaped string values whose decoded value is or resembles `>=1`, `>=2`, or `>=3`.
- A fixture-level decision for overflow-sized `>=<integer>` strings that chooses the exact diagnostic family instead of leaving it to "negative or unsupported according to parser boundary."

**Required changes:**
- Add raw-shape or grammar rows for `[requires] "formula_compiler" = ">=2"`, `[requires] "state_store" = ">=2"`, `requires."formula_compiler" = ">=2"`, `[requires."formula_compiler"]`, and a mixed bare-plus-quoted duplicate of `formula_compiler`. Each row should name the expected source key, diagnostic code, and duplicate behavior.
- State whether the requirement value grammar is evaluated over the decoded TOML string bytes or the raw TOML literal spelling. Then add rows for single-quoted, multiline, escaped ASCII, escaped Unicode, newline-containing, and control-character strings so the scanner cannot silently drift from the decoder.
- Update the introductory requirement mapping table to include explicit `>=1` as default compiler capability with `RequirementSource=requires`, or add a direct cross-reference to the later grammar section.
- Pin the overflow rule for `">=<huge integer>"` to one diagnostic in the fixture contract: either invalid syntax or unsupported future capability. The implementation can still explain parser-boundary limitations, but the generated golden diagnostics need one expected result.

**Questions:**
- Should quoted TOML key spelling of `formula_compiler` be allowed because TOML treats it as the same semantic key, or rejected because the design wants a single canonical authoring spelling?
- Should escaped TOML strings that decode to `>=2` be accepted, or must the raw source bytes literally contain `">=2"`?
- Is the raw scanner intended to preserve canonical TOML semantics for quoted keys and escaped strings, or is it intentionally a stricter authoring subset layered on top of the TOML decoder?
