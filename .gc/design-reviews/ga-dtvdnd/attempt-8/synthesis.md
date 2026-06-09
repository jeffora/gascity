# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins yields `block` because the embed/materialization persona synthesis blocks, even though the other nine persona syntheses are `approve-with-risks`. The requirements artifact conforms to the required `gc.mayor.requirements.v1` shape at a high level, but several remaining findings are document-fixable requirements gaps rather than merely missing downstream implementation detail. Proof artifacts and validators are also still external prerequisites before implementation approval.

## Consensus Strengths

- Reviewers agree the document has moved to the right product boundary: Core is required Gas City infrastructure, Gastown behavior is explicit external pack configuration, and Maintenance is retired rather than silently supplied.
- The requirements schema shape is substantially correct: front matter is present, required top-level sections are present and ordered, W6H is concrete, Example Mapping includes happy, negative, and edge paths, and Open Questions is `None`.
- AC6, AC7, AC14, AC15, AC16, and AC17 establish the right release-gate strategy: asset ledger, behavior preservation proof, public Gastown validation, pin/version policy, offline/cache behavior, and acceptance proof matrix.
- Diagnostics, doctor/import-state, and repair are treated as first-class operator surfaces with bootstrap-only behavior, stable condition codes, source attribution, non-interactive repair, and no silent fallback.
- The role-neutrality direction is strong: Go must not hardcode Gastown roles, Core-owned assets must stay role-neutral, and the default maintenance executor must be pack data rather than framework logic.

## Critical Findings

### [Blocker] Build and materialization contract still has unresolved acceptance decisions

**Sources:** Petra Novak; Priya Menon; Camille Okafor; Hugo Bautista

**Actionability:** document-fixable

**Issue:** The embed/materialization lane blocks because the requirements and implementation plan are not aligned on the canonical Core source root, and because surviving SDK/support-pack consumers still appear to depend on retired Maintenance assets. Reviewers specifically called out `examples/dolt/assets/scripts/port_resolve.sh` sourcing `.gc/system/packs/maintenance/assets/scripts/dolt-target.sh`, plus the need for first-class closure over Go embeds, builtin registry functions, public aliases, required-pack literals, synthetic layout/hash behavior, materialization commands, locks/caches, provider scripts, hooks, fixtures, docs, and tests.

**Required change:** Align the requirements and implementation plan on one canonical Core source root. Add a normative support-pack asset rule: any asset consumed by surviving Core/support packs but physically housed in retired Maintenance or in-tree Gastown must be rehomed, inlined, or explicitly retired with a replacement. Carry `dolt-target.sh` as a worked closure row and make the Go embed/registry/materialization layer an explicit source-consumer closure target.

### [Major] Required proof artifacts are still absent external prerequisites

**Sources:** Oleg Marchetti; Mara Voss; Mira Acharya; Simone Kaye; Camille Okafor; Hugo Bautista

**Actionability:** external-prerequisite

**Issue:** The requirements correctly name the asset migration ledger, behavior-preservation manifest or equivalent harness, public Gastown validation, pin ledger, version-skew matrix, diagnostics schema, docs-authority audit, role-neutrality scan, coverage-transfer map, source-consumer closure, and acceptance-proof matrix. Those artifacts are not present and executable yet, so they cannot serve as proof. This is acceptable for requirements review only if implementation approval remains blocked until the artifacts exist, run, and pass.

**Required change:** Keep requirements approval distinct from implementation approval. Before implementation approval, create and validate the named support artifacts, wire them into local/pre-commit/CI/release gates as AC17 requires, and fail closed when any AC lacks executable evidence.

### [Major] Pack-resolution and bootstrap behavior need sharper product policy

**Sources:** Priya Menon; Faisal Khoury; Mira Acharya; Petra Novak

**Actionability:** document-fixable

**Issue:** The requirements describe a broad resolution matrix but still need explicit policy for provider-conditioned `bd`/`dolt` activation, mutual exclusivity or co-activation, default-city cardinality, healthy no-conflict precedence, bootstrap-safe commands, behavior-changing commands blocked without Core, same-path/static asset collisions, and transitive diamond pin conflicts. Reviewers also warned that bootstrap-only commands such as `gc doctor` and `gc import-state` must not eagerly initialize normal pack resolution before diagnosing missing or corrupt Core.

**Required change:** Add or tighten AC3/AC11 matrix rows for provider-pack cardinality, operation classes, bootstrap-safe exceptions, static/path collisions, transitive diamond conflicts, and Core provenance/dedupe. Require init, doctor, CLI config load, and runtime resolution to agree on source attribution and condition codes for the same broken city state.

### [Major] Behavior preservation denominator can still shrink unless frozen first

**Sources:** Oleg Marchetti; Mira Acharya; Hugo Bautista; Simone Kaye

**Actionability:** document-fixable

**Issue:** AC7 needs to bind its supported-Gastown before-state denominator to the AC6 deterministic source snapshot captured before deletion or fixture isolation. Without that ordering and a frozen named supported-template/workflow list, validators could discover behavior after removal and accidentally produce an empty or narrowed preservation surface.

**Required change:** State that AC6 freezes the source denominator before any in-tree Gastown or Maintenance root is deleted or isolated. Require AC7 and AC14 to cover a frozen, named, version-pinned list of supported Gastown templates/workflows, and require behavior witnesses for side-effecting paths such as notifications, requester/detector links, identity side effects, escalation, and recovery.

### [Major] Existing-city migration needs a clean success path and stronger repair semantics

**Sources:** Camille Okafor; Faisal Khoury; Simone Kaye; Priya Menon

**Actionability:** document-fixable

**Issue:** Failure and diagnostic cases are well covered, but the central happy path for upgrading an existing Gastown city is missing. Repair semantics also need more product-level pinning: offline repair sequencing, local modified imports, read-only/transitive refusal, live-session refusal, interrupted/concurrent repair, inactive task-store references to retired paths, old-binary write prevention, stale legacy directory isolation, and rollback or restore expectations.

**Required change:** Add a happy-path Example Mapping row for successful existing-Gastown-city upgrade with operator-visible success evidence. Tighten AC10/AC15 around the repair state machine, doctor-written pin coherence, offline fail-before-mutation behavior, stale directory isolation, old-binary write prevention, inactive task-store policy, and rollback/restore outcome.

### [Major] Role-neutrality and configurable binding boundaries remain too easy to misread

**Sources:** Alistair Sterling; Mira Acharya; Faisal Khoury

**Actionability:** document-fixable

**Issue:** The allowed `dog` default is intentionally narrow, but the exact inert pack-data shape and denied executable uses still need precision. Reviewers also flagged implementation-plan risks around role-specific environment override names, override precedence, binding optionality, and doctor validation of user-configured bindings.

**Required change:** Clarify AC8/AC9 so the only permitted Core-owned `dog` occurrence is an inert configured-default binding data shape. Deny `dog` or Gastown roles as routes, notification targets, formula bindings, prompt defaults, overlays, generated defaults, or Go fallbacks. Require generic data-driven binding overrides, data-declared optionality/requiredness, rendered-output fixtures, substituted-executor fixtures, and doctor/import-state diagnostics for undefined or disabled bindings.

### [Major] Test strategy must prove active witnesses, not only mapped files

**Sources:** Mira Acharya; Oleg Marchetti; Hugo Bautista; Mara Voss

**Actionability:** document-fixable

**Issue:** AC7 and AC13 can pass too weakly if validators only check that files or test names exist. Reviewers require active execution evidence, negative/regression controls, assertion- or behavior-level coverage transfer, and failure on skipped, empty, no-op, or collapsed witnesses.

**Required change:** Require validators to parse active execution evidence such as `go test -json`, prove each mapped witness runs and fails when moved behavior is broken, and reject skipped/no-op/empty witnesses. Give the acceptance-proof matrix a named command plus a concrete local, pre-commit, or CI gate before decomposition.

### [Major] Public-pack authority audit is too Gas City-local

**Sources:** Simone Kaye; Oleg Marchetti; Camille Okafor

**Actionability:** document-fixable

**Issue:** AC12 and AC14 audit Gas City docs and behavior, but the public pinned `gascity-packs/gastown` pack is now part of the operator-facing product. Its `pack.toml` comments, command help, README/docs, prompts, prompt fragments, formula descriptions, doctor checks, generated docs, release notes, registry/catalog surfaces, and stale text references must be included.

**Required change:** Amend AC12/AC14/AC15 so docs-authority, public validation, registry/discovery proof, and pin coherence include the public Gastown pack at the pinned commit. Require row-level disposition for stale public-pack text and proof that docs, repair output, fresh-init output, registry discovery, source URL, subpath, commit, pack digest, and behavior-manifest digest agree.

### [Major] Diagnostics substrate and dev/test escape hatch need fail-closed boundaries

**Sources:** Faisal Khoury; Priya Menon; Mira Acharya; Camille Okafor

**Actionability:** document-fixable

**Issue:** The diagnostic direction is strong, but condition codes, severities, exit-code mapping, and schema rendering must be pack-independent compiled state so missing/corrupt Core cannot suppress its own diagnostic. The AC2 dev/test Core-less escape hatch is also unbounded and could either trip false positives or suppress real production missing-Core diagnostics.

**Required change:** State that diagnostic registry/schema rendering is pack-independent Go-side state and test it with no packs resolved. Define the dev/test escape hatch so it is available only to native test paths and cannot drive production CLI, doctor, controller, runtime, session, dispatch, formula expansion, or city-state mutation.

### [Major] Schema conformance passes, but traceability and artifact-shape refinements are required

**Sources:** Mara Voss; Mira Acharya; Hugo Bautista

**Actionability:** document-fixable

**Issue:** The requirements artifact conforms to the required output schema, but several acceptance criteria are compound enough that downstream decomposition can lose traceability. Some support artifact paths and formats also need to be clearly marked as normative acceptance artifacts rather than illustrative examples.

**Required change:** Split the largest ACs or add sub-IDs for material behaviors. Clarify which `support/*.yaml` and `support/*.json` artifacts are binding evidence contracts. Keep implementation assignments out of requirements, but make product policies and proof obligations independently traceable.

### [Minor] Persona synthesis artifacts were written to the wrong attempt path

**Sources:** Workflow artifacts; all persona syntheses

**Actionability:** workflow-defect

**Issue:** The active attempt directory `.gc/design-reviews/ga-dtvdnd/attempt-8/persona-syntheses/` is empty, while all 10 current persona syntheses were written to `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/` with timestamps from this iteration. Several raw reviews explicitly mention a known `gc.attempt=1` path defect. I used the current fallback persona syntheses to avoid dropping valid reviewer output, but this is a malformed review artifact shape.

**Required change:** Fix the persona-synthesis workflow path/attempt metadata so current-attempt syntheses are written under `$ATTEMPT_DIR/persona-syntheses/`. Add a control check that fails or repairs the path mismatch before global synthesis starts.

## Disagreements

- Petra Novak's persona synthesis escalates to `block` even though its three raw reviewers reportedly returned `approve-with-risks`. Assessment: the escalation is justified because the `dolt-target.sh` fresh-city risk and Core source-root mismatch are unresolved acceptance decisions for the build/materialization lane.
- DeepSeek/Gemini-perspective reviews often approved the direction while Claude/Codex framed the same areas as risks. Assessment: treat the direction as strong, but preserve `approve-with-risks` for every non-blocking persona because the executable proof gates are not present yet.
- Reviewers split on whether some gaps belong in requirements text or only in implementation-plan/support artifacts. Assessment: product policies belong in `requirements.md`; executable proofs can live in support artifacts, but the requirements must make those artifacts mandatory and traceable.
- Some reviewers accept the current artifact paths as recovery from a workflow defect, while the required contract says persona syntheses should be in the current attempt directory. Assessment: the content was usable because all 10 fallback syntheses are current, but the workflow path defect should be fixed separately.

## Missing Evidence

- Current-attempt persona syntheses under `.gc/design-reviews/ga-dtvdnd/attempt-8/persona-syntheses/`.
- Final Core source-root policy and proof that requirements and implementation plan agree.
- Source-consumer closure for Go embeds, builtin registry functions, public aliases, materialization commands, provider scripts, hooks, fixtures, docs, tests, locks, and caches.
- Concrete rehome, inline, or retirement decision for `dolt-target.sh` and other former Maintenance behaviors such as `_bd_trace.sh`, status-line tracing, dog molecules/orders, `jsonl-export.sh`, and `reaper.sh`.
- AC6 asset migration ledger, AC7 behavior-preservation manifest or equivalent harness, AC14 public Gastown validation, AC15 public pin ledger and version-skew matrix, AC16 offline/cache proof, and AC17 acceptance-proof matrix.
- Frozen supported-Gastown before-state denominator and version-pinned supported template/workflow list.
- Recursive import-chain tracing design, pack-independent diagnostic registry/schema, bootstrap-only command isolation, repair transaction state machine, and dev/test escape-hatch boundary.
- Existing-city upgrade happy-path evidence, custom local `packs/gastown` fixture, old-binary write-prevention proof, inactive task-store migration policy, stale directory isolation, and rollback/restore outcome.
- Public-pack docs authority audit covering public `gascity-packs/gastown` operator surfaces at the pinned commit.
- Active witness execution evidence and negative/regression controls for behavior preservation, coverage transfer, and role-neutrality scans.

## Convergence Assessment

- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: At least one remaining blocker or major finding is document-fixable. A targeted requirements edit can align the Core source-root policy, add the support-pack asset rule, freeze the behavior denominator, clarify resolution/repair/role-neutrality/test policies, and make support artifacts mandatory without needing to invent proof artifacts inline.
- Next non-design work: create and validate the named support artifacts, validators, public Gastown proof, diagnostics schema, pin/version/cache proofs, repair tests, and workflow path fix before implementation approval.

## Recommended Changes

1. Align `requirements.md` and `implementation-plan.md` on the canonical Core source root, and state whether any other Core-like path is deleted, retained as canonical, a compatibility shim, or a non-runtime fixture.
2. Add a support-pack asset closure rule and worked `dolt-target.sh` row so surviving SDK/support packs cannot depend on retired Maintenance paths.
3. Make Go embed, builtin registry, materialization, public alias, synthetic layout/hash, lock/cache, provider-script, hook, fixture, docs, and test consumers first-class closure targets.
4. Freeze the AC6 source denominator before deletion/isolation and require AC7/AC14 to cover a named, version-pinned supported Gastown template/workflow list.
5. Tighten AC3/AC11 with provider-pack cardinality, operation classes, bootstrap-safe exceptions, static/path collisions, diamond conflict policy, and cross-surface condition-code consistency.
6. Add an existing-city happy-path upgrade example and tighten repair semantics for offline, interrupted, concurrent, read-only, transitive, live-session, local-modification, stale-directory, inactive-task, old-binary, and rollback states.
7. Clarify AC8/AC9 around the exact allowed inert `dog` binding, denied executable uses, generic binding overrides, binding optionality metadata, rendered-output fixtures, and doctor binding diagnostics.
8. Require active witness execution and negative/regression controls for AC7/AC13, plus a named acceptance-proof validation command and gate before decomposition.
9. Extend docs-authority and public-pack validation to the pinned public `gascity-packs/gastown` pack and require docs/registry/repair/fresh-init/pin-ledger coherence.
10. Fix the design-review workflow so persona syntheses for attempt 8 are written under the current attempt directory before global synthesis.
