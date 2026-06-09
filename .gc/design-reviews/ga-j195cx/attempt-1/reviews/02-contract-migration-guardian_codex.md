# Elias Vega - Codex

**Verdict:** block

**Top strengths:**
- Legacy `contract = "graph.v2"` remains an accepted alias during the migration window, including current whitespace/case spellings accepted by the released compiler.
- First-party requires-only conversion and parser alias removal are separated into Phase 8 and Phase 9, with release-captain ownership, minimum binary floors, old-reader evidence, external-support evidence, and rollback units.
- SHA-pinned external packs are treated as immutable source objects, so the design does not pretend maintainers can edit a pinned revision in place.

**Critical risks:**
- [Major] The external-support evidence has two incompatible schema/status contracts. `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:2724` defines the canonical artifact as `alias_window` plus `rows[]`, with row-level `status`, `support_window_closes_at`, and `blocks_alias_removal`. Later, `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:3085` says the same artifact contains top-level `owner`, `status`, `supported_old_binaries`, `support_strategy`, and `expires_after_release`, and that CI blocks on top-level `status: active`. The release table at `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:3117` then describes success as "all rows expired/not-needed." Those are not the same contract. One implementation could pass alias removal by checking a top-level status while row-level SHA-pinned support still blocks, while another could reject the row-based artifact because top-level status is absent.
- [Minor] The release-evidence format contract says migration evidence is JSON with `schema_version`, `generated_at`, owner, command, exit contract, and consumer fields at `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:95` and again says every migration gate consumes JSON at `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:3103`. The compatibility artifact is specified as YAML at `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:3049` and listed as `formula-compiler-compatibility.yaml` at `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:3114`. That may be intentional, but the current wording makes the release evidence rule self-contradictory.

**Missing evidence:**
- No seeded `formula-compiler-external-support.json` example shows how unknown external packs initialize as blocking and later transition to expired or not-needed.
- No alias-removal gate sample demonstrates row-level behavior for an active SHA-pinned legacy formula alongside expired and not-needed rows.
- No digest flow is shown for the YAML compatibility artifact if it remains an exception to the JSON-evidence rule.

**Required changes:**
- Pick one canonical external-support schema. Prefer the `alias_window` plus `rows[]` schema, remove the top-level `status` wording, and define any aggregate status as a derived field in the alias-removal gate report.
- Add fixtures for `external-support.json` covering `unknown`, `unreachable`, `sha_pinned_legacy`, `expired`, and `not_needed` rows, with expected exit codes for `--external-support`, `--requires-only-conversion-gate`, and `--alias-removal-gate`.
- Either convert `formula-compiler-compatibility.yaml` to JSON or explicitly exempt it from the "JSON artifacts" rule and state how its schema and digest are validated by the JSON gate reports.

**Questions:**
- Is external-support status meant to be a cached top-level aggregate, or are support decisions exclusively per-row?
- Does `gc formula validate --all-packs --external-support --json` write the canonical support artifact, validate an existing artifact, or produce a separate report consumed by the gate?
