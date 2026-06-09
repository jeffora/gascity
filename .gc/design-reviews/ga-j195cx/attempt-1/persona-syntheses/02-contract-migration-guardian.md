# Elias Vega

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The migration does not yet guarantee that existing `contract = "graph.v2"` formulas keep compiling on first upgrade. Claude identifies an unresolved default or migration policy for `[daemon] formula_v2`; if the first requirements-normalizing binary defaults this gate to `false`, unchanged legacy graph formulas can fail with `formula.compiler_requirement_unsatisfied` before the alias window has a chance to collect evidence.
- [Major] The external-support evidence contract is internally inconsistent. Codex finds two incompatible descriptions for `formula-compiler-external-support.json`: one row-based schema under `alias_window` and `rows[]`, and another top-level status contract. Claude separately flags stale, unknown, SHA-pinned, and active-support rows as release-gate inputs whose freshness and blocking semantics are not precise enough. This can make alias removal pass or fail depending on which implementation reads the artifact.
- [Major] Accepted-alias drain evidence is not tied to a durable, complete writer contract. Claude reports that runtime/background accepted-alias counts are required for alias removal while warning emission, producer-local LRUs, process restarts, and config-generation resets can drop evidence. Codex also asks where this count belongs and what command writes it. Without producer coverage, a zero count can mean missing evidence rather than no remaining legacy usage.
- [Major] The compatibility corpus is load-bearing but not reproducibly seeded from supported old binaries. Claude finds no named artifact or generator that enumerates every spelling the currently released compiler accepts at alias-window start. Codex's artifact-format concern reinforces that the corpus and its digest path must be explicit if release gates depend on it.
- [Minor] Evidence freshness, waiver expiry, and rollback boundaries remain ambiguous. Claude identifies undefined `last_validated_at` staleness, unbounded active-root waivers, and drain-window wording that can accidentally span a Phase 8 rollback boundary.
- [Minor] The release-evidence format rule conflicts with the compatibility artifact format. Codex notes that migration evidence is generally required to be JSON, while `formula-compiler-compatibility.yaml` is named as YAML without an explicit exception, schema validation, or digest flow.

**Disagreements:**
- Claude's verdict is `approve-with-risks`; Codex's verdict is `block`. Assessment: choose `block` because the unresolved `formula_v2` upgrade default can break legacy formulas immediately, and the external-support schema contradiction can make the alias-removal gate fail open.
- Claude emphasizes upgrade compatibility, durable accepted-alias evidence, corpus seeding, freshness, waivers, and rollback windows. Codex emphasizes the external-support schema conflict and evidence-format contradiction. Assessment: these are related contract-migration failures; the design needs one authoritative compatibility and release-evidence contract before this lane can approve.
- Codex treats the YAML-versus-JSON compatibility artifact issue as minor; Claude does not mention it. Assessment: keep it as a required cleanup unless the design explicitly exempts that artifact and defines equivalent validation and digest behavior.

**Missing evidence:**
- Kimi 2.6 was not present; this synthesis uses the required Claude and Codex reviews only.
- A pinned default and upgrade-migration policy for `[daemon] formula_v2`, plus a release-gate test proving default-config legacy `contract = "graph.v2"` formulas compile with warnings after upgrade.
- One canonical `formula-compiler-external-support.json` schema, including row states, aggregate status if any, blocking semantics, and examples for `unknown`, `unreachable`, `sha_pinned_legacy`, `expired`, and `not_needed`.
- A durable producer contract for every accepted-alias compile path, including storage location, retention, reset behavior, owner, and tests that each caller path writes queryable evidence.
- A corpus-seeding command and checked-in artifact generated from each supported old binary's actual accepted and rejected spellings at alias-window start.
- Numeric freshness thresholds for external-support discovery and alias-drain evidence, expiry or re-review rules for active-root waivers, and post-rollback drain-window digest rules.
- A validation and digest path for `formula-compiler-compatibility.yaml`, or a decision to convert it to JSON with the other migration evidence artifacts.

**Required changes:**
- Add a Phase 2 or Phase 5 rollout requirement that pins `[daemon] formula_v2` upgrade behavior. Either default it to `true` for binaries that ship requirements normalization or define an automatic promotion/migration when existing inventory contains legacy graph formulas or formula-compiler requirements. Add a gate test for default-config legacy formula compile success with a deprecation warning.
- Collapse external-support evidence into one authoritative schema and one gate interpretation. Prefer the row-based `alias_window` plus `rows[]` contract, with any aggregate status derived by the gate report rather than stored as a competing source of truth.
- Add fixtures and exit-code expectations for external-support and alias-removal gates covering `unknown`, `unreachable`, `sha_pinned_legacy`, `expired`, and `not_needed` rows, including a mixed active SHA-pinned legacy case that must block removal.
- Define the accepted-alias evidence writer contract and register every accepting producer path against it. CI should fail when a caller can accept `contract = "graph.v2"` without writing durable evidence consumed by the drain gate.
- Add the compatibility-corpus seeding command, record the generated corpus digest in `formula-compiler-alias-window-start.json`, and require gate reports to fail on digest mismatch or unsupported narrowing during the alias window.
- Define `discovery_freshness_days`, waiver `expires_at` and `next_review_at`, and rollback-aware drain-window rules requiring qualifying reports to reference the post-rollback alias-window-start artifact or rollback marker.
- Resolve the JSON/YAML evidence contradiction by converting `formula-compiler-compatibility.yaml` to JSON or explicitly exempting it while defining its schema validation and digest inclusion in JSON gate reports.
