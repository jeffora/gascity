---
plan_slug: core-gastown-pack-migration
phase: requirements
rig: gascity
rig_root: /data/projects/gascity
artifact_root: /data/projects/gascity/plans
status: questions
created_at: 2026-06-04T00:00:00Z
updated_at: 2026-06-09T01:20:00Z
---

# Requirements: Core and Gastown Pack Split

## Problem Statement

Gas City still carries SDK-required pack behavior, generic maintenance behavior,
and Gas Town-specific orchestration behavior across legacy source-tree pack
locations:

- `internal/bootstrap/packs/core`
- `examples/gastown/packs/maintenance`
- `examples/gastown/packs/gastown`

That layout makes ownership ambiguous. A non-Gastown Gas City deployment can
appear to depend on Maintenance or Gastown assets, while Gastown behavior can
appear to be supplied by in-tree examples instead of by the public Gastown pack.
It also obscures which behavior is required SDK infrastructure and which
behavior is user-supplied orchestration configuration.

The desired product outcome is a role-neutral Gas City runtime in which Core is
the only Gas City-owned pack required for normal `gc` operation, Maintenance is
retired as a standalone pack, and Gastown behavior is loaded explicitly from the
public `gascity-packs/gastown` pack. Existing operators must receive actionable,
non-interactive diagnostics and repair guidance for legacy pack references.
Gastown operators must not lose current supported behavior when Core assets are
generalized.

<!-- REVIEW: added per requirements-schema-compliance-officer -->

The requirements remain in `questions` status because the next planning step
must attach or generate two supporting artifacts before implementation planning
can be treated as complete: a validated asset migration ledger and a
behavior-preservation proof manifest. The requirements below define the product
contracts those artifacts must satisfy; they do not embed a file-by-file
implementation map in this requirements document.

## W6H

<!-- REVIEW: added per requirements-schema-compliance-officer -->

| Dimension | Requirement |
| --- | --- |
| Who | Gas City maintainers need an auditable Core pack boundary; Gas City operators need stable non-Gastown runtime behavior; Gastown operators need explicit public Gastown imports; pack authors need clear ownership rules; existing-city upgraders need safe migration diagnostics. |
| What | Split the current mixed pack behavior into required Core SDK behavior, retired Maintenance behavior, and explicit external Gastown behavior. Core must stay role-neutral; Gastown roles, workflows, prompts, commands, overlays, and role-specific recovery behavior belong to the public Gastown pack. |
| When | Fresh `gc init` output must use the new model immediately after the migration ships. Existing cities must be diagnosable before startup or operator workflows fail from retired pack paths. The public Gastown pack and Gas City release must be version-compatible before the migration is accepted. |
| Where | The migration applies to Gas City source-tree packs, generated/materialized system packs, public pack imports, city and rig config, docs, examples, doctor/import-state output, and tests. It also applies to external `gascity-packs/gastown` validation because that pack becomes the Gastown authority. |
| Why | Gas City's SDK must not encode Gastown roles or Maintenance as hidden framework behavior. Role behavior must be user-supplied pack configuration, while SDK infrastructure remains self-sufficient with only the controller and required Core runtime support. |
| How | Core is required in resolved config for real cities. Gastown is an explicit import using `https://github.com/gastownhall/gascity-packs.git//gastown` with the pinned version used by fresh Gastown init. Maintenance is neither auto-included nor silently materialized. Diagnostics report missing Core and retired imports with exact source attribution and explicit repair actions. |
| How much / scale | The migration covers every active pack asset under the three legacy pack roots, every generated/materialized system-pack source that can load those assets, every supported fresh-init template, all existing-city legacy import shapes, and all docs/tests that present retired paths as authoritative. |

## Example Mapping

<!-- REVIEW: added per requirements-schema-compliance-officer -->

| Type | Concrete example | Expected behavior | Evidence |
| --- | --- | --- | --- |
| Happy path | A new non-Gastown city is initialized and run with no Gastown import. | Resolved config includes required Core behavior and any provider-conditioned support packs, but no Gastown roles, Gastown workflows, or Maintenance pack import. The controller can execute SDK infrastructure operations without a configured Gastown agent. | Unit or integration test for fresh non-Gastown init plus resolved-config inspection. |
| Happy path | `gc init --template gastown` creates a Gastown city. | The root `pack.toml` and default rig import declare `[imports.gastown]` from `https://github.com/gastownhall/gascity-packs.git//gastown` with the current pinned version (`sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` as of this requirements update). No in-tree `examples/gastown/packs/gastown`, `.gc/system/packs/gastown`, or implicit Maintenance fallback is used. | Command verification or existing init test extended to check source, version, lock/cache provenance, and absence of legacy fallback. |
| Negative path | A city config or pack import graph omits Core. | `gc doctor` and `gc import-state` can still report the problem instead of failing before diagnostics. The report identifies the config source/layer that omitted Core, states that Core is required for real cities, and offers an explicit idempotent repair action. | Golden-output and JSON-output tests for missing-Core diagnostics. |
| Negative path | An existing city imports `packs/maintenance`, `packs/gastown`, `examples/gastown/packs/gastown`, or `.gc/system/packs/maintenance`. | Diagnostics classify the reference as retired or legacy, name the source that introduced it, distinguish required Core from optional Gastown, and do not silently redirect to a hidden fallback. | Import-state and doctor tests covering each legacy path shape. |
| Negative path | A Core asset or generated Core metadata routes work to a literal Gastown role name such as Mayor, Deacon, Polecat, Refinery, Witness, Boot, Crew, or Gastown. | Role-neutrality checks fail unless the occurrence is in allowed migration documentation, generated review artifacts, or a negative/absence test fixture. | Absence-scan test with positive controls and allowed-path exceptions. |
| Edge case | A user renamed or disabled the default Core maintenance executor. | SDK infrastructure still operates with only the controller. Agent-executed Core maintenance work is bound through configurable pack data; if no executor is available, the work remains visible or diagnosable rather than causing Go-side role special-casing. | Unit/integration test using a renamed executor and a no-executor city. |
| Edge case | The operator is offline while a Gastown city uses a lockfile for the public Gastown pack. | If the exact pinned external pack is already present in the repo cache, resolution uses that cache and reports provenance. If not, resolution fails with an actionable missing-cache diagnostic and never falls back to in-tree examples or system packs. | Offline cache-hit and cache-miss command tests. |
| Edge case | Stale `.gc/system/packs/maintenance` or old synthetic cache state exists after upgrade. | The stale state is ignored, pruned by an explicit repair action, or reported as retired state. It is never loaded as an active Maintenance pack. | Upgrade-state integration test or command verification. |

## Acceptance Criteria

<!-- REVIEW: added per acceptance-traceability -->

| ID | Criterion | Verification |
| --- | --- | --- |
| AC1 | The requirements artifact follows `gc.mayor.requirements.v1`: required front matter only, required top-level section order, W6H, example mapping, consolidated acceptance criteria, out-of-scope, and open questions. | Design-review schema check or manual review against `/data/projects/gascity-packs-worktrees/gc-plan-pack/gascity/assets/skills/mayor/requirements.schema.md`. |
| AC2 | Core is represented as required runtime behavior in resolved config for real cities, with a clear dev/test escape hatch if tests need to construct partial configs. Normal SDK infrastructure operations must not depend on any Gastown role existing. | Resolved-config tests for fresh cities, partial test configs, and controller-only operation. |
| AC3 | Pack resolution is deterministic across required Core, provider-conditioned support packs such as `bd` and `dolt`, root pack imports, city imports, default rig imports, locked remotes, cached remotes, system packs, and local overlays. User imports may not shadow required pack identity without an explicit collision diagnostic. | Pack-resolution matrix tests covering precedence, collisions, lock/cache provenance, and provider-conditioned support-pack cases. |
| AC4 | Fresh Gastown init imports the public Gastown pack explicitly from `https://github.com/gastownhall/gascity-packs.git//gastown` with a pinned version and default rig import. It must not rely on in-tree examples, `.gc/system/packs/gastown`, or implicit Maintenance. | Command verification for `gc init --template gastown`, lock/cache inspection, and absence checks for retired fallback paths. |
| AC5 | Maintenance is retired as a standalone active pack. It is not bundled, public-source recognized, auto-included, materialized as an active system pack, or presented as an implicit dependency. Legitimate non-pack uses of the word "maintenance" remain allowed. | Builtin/materialization tests, import-state checks, source alias tests, and docs/terminology audit. |
| AC6 | A validated asset migration ledger exists outside this requirements document before implementation approval. The ledger records current path, provenance, target owner, target output path or retirement action, split boundary, fallback classification, rationale, and proof command. It fails on missing current paths, unrepresented active source files, unresolved `review` rows, and duplicated or orphaned split behavior. | Ledger generation/validation command plus checked artifact in the plan or implementation context. |
| AC7 | A behavior-preservation manifest or harness exists before implementation approval. It covers supported Gastown formulas, orders, scripts, prompts, template variables, notification targets, requester/detector relationships, success/warning/failure/escalation paths, and recovery flows. | Behavior manifest validation plus runtime or command checks that external Gastown workflows load, render, trigger, and deliver expected notifications. |
| AC8 | Core role neutrality is enforced across Go production code, Core assets, formulas, orders, prompts, provider overlays, generated/materialized metadata, and route targets. Literal Gastown role names and literal `dog` routing are prohibited except for documented configured-default data, migration docs, generated review artifacts, and absence-test fixtures. | Absence-scan test with explicit denied tokens, scan roots, allowed paths, positive controls, and negative controls. |
| AC9 | The Core maintenance executor contract is configurable. Core may ship a default maintenance executor named `dog` as pack configuration, but Go code and SDK infrastructure must treat it as user-supplied config. Renamed, replaced, omitted, or disabled executors are handled through configuration and diagnostics, not hardcoded role logic. | Config tests for default, renamed, omitted, and disabled executor cases; Go-source scan for role special-casing. |
| AC10 | Existing-city migration diagnostics cover legacy local imports, stale system packs, public pins/caches, custom local overlays, duplicate Core, retired Maintenance, missing Core, version skew between Gas City and public Gastown, rollback expectations, and in-flight runtime state. Repair is report-only by default; any mutation requires an explicit non-interactive operator action and is idempotent. | Upgrade matrix tests, JSON/text golden outputs, repair idempotence checks, read-only config tests, and post-repair verification. |
| AC11 | Doctor/import-state output identifies the exact resolved config source or nested import that caused a missing-Core or retired-path condition. Text output is operator-readable; JSON output is structured for automation; neither mode prompts interactively or hides errors by substituting defaults. | Golden-output tests for text and JSON, nested-import attribution tests, read-only handling tests, and duplicate-Core prevention tests. |
| AC12 | Documentation, examples, CLI help, doctor messages, and import-state output consistently describe Core as required, Gastown as external/optional unless the user chooses the Gastown template, and the Maintenance pack as retired. They do not present `packs/maintenance`, `packs/gastown`, `examples/gastown/packs/*`, or `.gc/system/packs/*` as authoritative current sources except in migration/history contexts. | Docs/examples/help grep or generated audit artifact with allowed archive and migration-history exceptions. |
| AC13 | Legacy tests that asserted implicit Maintenance or in-tree Gastown behavior are either removed with explicit replacement coverage or rewritten to prove required Core plus explicit external Gastown imports. | Coverage-transfer table mapping old tests to new tests, command checks, absence scans, or manual checks. |
| AC14 | Public Gastown validation is part of acceptance. A local in-tree copy cannot mask a broken external pack: the public pack checkout or pinned cache must prove the roles, prompts, commands, formulas, orders, overlays, and checks needed by supported Gastown templates. | Cross-repo or pinned-cache validation command documented in the implementation plan and run in CI or as an explicit release gate. |

## Out Of Scope

- Redesigning MEOW primitives or adding new SDK primitives.
- Reworking the Gastown role model beyond moving Gastown behavior to the
  explicit public pack authority.
- Changing unrelated third-party packs except where their import metadata,
  tests, or documentation point at retired Maintenance or in-tree Gastown
  authority paths.
- Renaming unrelated Go packages, config fields, docs, or operational concepts
  that use "maintenance" for supervisor, store, cleanup, or reliability work
  rather than the retired Maintenance pack.
- Implementing file moves, code changes, or bead creation during the
  requirements phase.
- Replacing the public pack registry/import mechanism.
- Keeping source-code comments solely to preserve the history of removed pack
  locations.
- Treating Core as a new role system or adding framework-side decision logic
  about role behavior.

## Open Questions

<!-- REVIEW: added per requirements-schema-compliance-officer -->

1. Who owns generating and maintaining the validated asset migration ledger
   required by AC6: Gas City only, `gascity-packs/gastown` only, or a shared
   release artifact across both repositories?
2. What command or release gate will produce the behavior-preservation manifest
   required by AC7, and where will the proof artifact live so both Gas City and
   public Gastown releases can consume it?
3. What is the supported version-skew window between the Gas City release that
   requires public Gastown and the `gascity-packs/gastown` commit pinned by
   fresh init?
4. For existing cities with in-flight sessions using prompts or formulas from
   retired paths, should the migration allow those sessions to finish with old
   materialized content, require an immediate restart after repair, or expose a
   separate operator decision?
5. What exact repair command or workflow will perform explicit config mutation
   after doctor/import-state reports missing Core or retired imports?
