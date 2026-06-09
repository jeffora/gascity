# Elias Sato - Codex

**Verdict:** block

**Top strengths:**
- The design now makes required Core proof typed and resolver-produced rather than path/count based, including digest, layer, import-edge, collision, load-mode, and diagnostic fields.
- The default-deny loader inventory is scoped broadly enough to catch production `cmd/gc` and behavior-driving `internal/` bypasses, including aliases, wrappers, raw TOML pre-reads, no-refresh paths, API/controller helpers, packman, cache/lock validation, completion, stop, and exported resolved-config helpers.
- Runtime failure semantics are mostly explicit: required-pack integrity is checked before behavior discovery, post-resolution participation is fatal, no-refresh paths cannot repair, and partial reads cannot consume agents, formulas, orders, prompts, hooks, overlays, scripts, patches, or route metadata.

**Critical risks:**
- [Blocker] The Core Presence Doctor subsection still instructs the plain doctor check to call `internal/systempacks.LoadRuntimeCity` while the later read-only doctor boundary forbids materializing runtime loaders in plain `gc doctor`. In `.gc/design-review-inputs/core-gastown-pack-migration/design.md`, lines 2109-2122 say plain doctor may only use no-refresh/read-only APIs and must not call materializing runtime loaders, repair helpers, or writers. Lines 3055-3068 then tell `cmd/gc/core_pack_doctor_check.go` to verify the file set and "Load resolved config through `internal/systempacks.LoadRuntimeCity`." That is not a harmless wording issue: it gives implementers two incompatible contracts for the same check, and the later concrete file-level instructions can hide missing/corrupt Core by repairing or refreshing during report-only diagnostics.
- [Major] Provider host-pack selection is still not concrete enough to implement without accidental behavior reads. Lines 1883-1888 require Core plus provider packs to be selected after reading the final effective beads provider but before any behavior-bearing config is read. Lines 2429-2436 name `internal/systempacks.RequiredPackPlan`, but the design does not specify the prepass inputs, how fragments/includes/env overrides are handled, what happens when the prepass and final resolved provider disagree, or how malformed partial config is classified. This is adjacent to the Core invariant because the participation record must cover `bd` and `dolt` with the same fatal gates as Core.
- [Major] The generated loader inventory is the binding source of truth, but its implementation contract lacks a concrete generator/test entrypoint and scanner fixture list. Lines 2411-2427 describe what it scans and rows it emits, and lines 2586-2595 list a "loader scanner test", but the design should name the command/package/test and fixture cases for aliases, wrapper methods, function values, package aliases, exported config-returning helpers, raw TOML decoders, hand-built include lists, no-refresh variants, and newly exported loaders. Without that, the default-deny promise can regress into a hand-maintained allowlist.

**Missing evidence:**
- A corrected plain-doctor flow showing `core_pack_doctor_check.go` uses only `ValidateRequiredFileSetsNoRefresh`, `LoadRuntimeCityNoRefresh`, raw import parsing, and classifier APIs, with `LoadRuntimeCity` reserved for runtime commands or post-`--fix` final validation.
- The `RequiredPackPlan` algorithm: exact fields consumed in the provider prepass, how layered city/rig config, fragments/includes, env overrides, missing TOML, malformed TOML, and final/prepass mismatch are handled, and which diagnostics are fatal.
- Named scanner generator and CI tests for `loader-inventory.generated.yaml`, including negative fixtures for bypass classes that are easy to miss.

**Required changes:**
- Rewrite the Core Presence Doctor section so plain `gc doctor` is read-only/no-refresh and cannot materialize, repair, promote cache entries, prune/quarantine, or rewrite imports. Keep staged repair and `LoadRuntimeCity` final validation only under `gc doctor --fix` through `doctor.MutationCoordinator`.
- Add the provider-pack selection algorithm to the System Pack Loading design, including prepass/final mismatch handling and tests for root config, rig config, fragments/includes, env overrides, malformed partial config, `bd`, `dolt`, exec/no-provider cases, and stale retired sources named like provider packs.
- Name the loader-inventory generator entrypoint, schema test, freshness test, and negative fixture suite. Make stale rows and newly exported config-returning helpers fail CI by default.

**Questions:**
- Should `LoadRuntimeCityNoRefresh` return a `RequiredSystemPackParticipation` record for an already-invalid tree, or should it return only structured diagnostics without a resolved config?
- Is provider selection allowed to parse imported fragments before required host-pack validation, or must provider selection be limited to a stricter raw city/rig TOML prepass?
