# Faisal Khoury - Codex

**Verdict:** block

**Top strengths:**
- The draft makes operator-facing diagnostics a first-class product outcome, not an implementation afterthought. The Problem Statement requires actionable, non-interactive diagnostics and repair guidance for legacy pack references, and W6H requires exact source attribution plus explicit repair actions.
- The Example Mapping covers the main diagnostic entry points for this lane: missing Core, retired local pack paths, offline cache behavior, and stale `.gc/system/packs/maintenance` state.
- AC10-AC12 correctly require read-only defaults, non-interactive mutation, JSON/text golden outputs, nested-import attribution, and consistent Core/Gastown/Maintenance terminology across doctor, import-state, docs, help, and examples.

**Critical risks:**
- [Blocker] The repair contract is simultaneously required and unresolved. W6H says diagnostics must provide "explicit repair actions", the missing-Core example requires an "explicit idempotent repair action", and AC10 requires idempotent non-interactive mutation, but Open Question 5 still asks what exact repair command or workflow will perform that mutation. An implementation plan cannot safely proceed without either resolving that product decision or weakening the requirement to report-only candidate remediation until the repair workflow is designed.
- [Major] "Exact source attribution" is not defined tightly enough for doctor/import-state parity. AC11 says to identify the resolved config source or nested import, but the requirements do not specify the minimum diagnostic shape: stable code, severity, config file path or remote source, import chain, pack identity, resolved version or cache provenance, retired/missing/duplicate classification, and remediation text/command. Without that, JSON and text output can pass vague golden tests while still being unsafe for operators.
- [Major] Existing-city cases need a safety classification before they share one repair story. AC10 groups legacy local imports, stale system packs, public pin/cache issues, duplicate Core, missing Core, version skew, rollback expectations, and in-flight runtime state. Those do not have the same safe remediation surface: some are read-only warnings, some are explicit config edits, some are cache pruning, and in-flight runtime state may require an operator decision before restart. The requirements need to separate these action classes.
- [Minor] The terminology requirement is good, but it lacks at least one canonical message example. This migration is specifically about avoiding confusion between required Core, optional external Gastown, and retired Maintenance; one text and one JSON example would reduce downstream drift.

**Missing evidence:**
- A sample `gc doctor` and `gc import-state` diagnostic for missing Core, including the exact source attribution and repair guidance expected by AC11.
- A sample retired-path diagnostic for at least one nested import and one stale `.gc/system/packs/maintenance` state.
- A chosen repair command/workflow, or an explicit product decision that mutation is out of scope for this requirements artifact.
- A diagnostic classification matrix mapping each AC10 condition to report-only, config mutation, cache/system-state cleanup, version repair, or runtime/session operator decision.
- Evidence that doctor/import-state can enter a degraded diagnostic path when normal config resolution would fail because Core is missing or a retired import path is present.

**Required changes:**
- Resolve Open Question 5 before approval, or change the diagnostics requirements to say the commands report exact candidate remediation only and do not mutate config in this migration.
- Add a minimum diagnostic contract shared by doctor and import-state: diagnostic code, classification, severity, source location, import chain, pack identity, required-vs-optional status, text message, JSON fields, and remediation.
- Add a remediation safety table for missing Core, duplicate Core, retired Maintenance import, legacy Gastown import, stale system pack, custom local overlay, public pack cache miss, version skew, rollback, and in-flight runtime state.
- Add acceptance coverage that proves doctor and import-state use the same diagnostic model and differ only in presentation or command scope.
- Add at least one canonical text-output snippet and one canonical JSON-output shape for required Core versus optional Gastown versus retired Maintenance.

**Questions:**
- Is explicit config mutation intended to ship in this migration, or should this requirements artifact stop at report-only diagnostics with a separately planned repair workflow?
- For nested imports, what source identity is operator-visible: the top-level city file, the importing pack, the remote URL and version, the nested import name, or all of them?
- Should stale `.gc/system/packs/maintenance` be ignored automatically, pruned only by explicit command, or reported until the operator runs a cleanup workflow?
- What should doctor say when an in-flight session still references materialized content from retired paths: finish, restart, block, or ask the operator through a separate decision flow?
