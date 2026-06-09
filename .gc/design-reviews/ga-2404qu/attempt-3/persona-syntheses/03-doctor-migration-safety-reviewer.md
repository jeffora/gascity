# Sofia Khoury

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash (available third review artifact is named with a Gemini suffix but identifies itself as DeepSeek V4 Flash)

**Consensus findings:**
- [Blocker] The legacy Gastown rewrite path is not safely preflighted or failure-atomic. All reviewers point at the same unsafe shape: current code rewrites `pack.toml` and possibly `city.toml` before remote resolution, lockfile writing, and install complete. A failed registry lookup, unreachable public source, lockfile write, or install can leave a working local import replaced by a partially migrated public import. Per-file temp rename is not a transaction across manifests, lockfiles, and installed pack directories, and `pack.toml` has no rollback path today.
- [Blocker] Automatic rewrite/removal provenance is underspecified and can affect operator-owned data. The design says generated legacy Gastown/Maintenance/Core imports can be fixed automatically while forks and edited packs are manual, but it does not define proof of generated, unmodified provenance. The existing suffix-based matcher cannot distinguish `.gc/system/packs/*`, copied examples, development forks, edited generated directories, or deliberate Core pins.
- [Blocker] Scoped TOML editing is unresolved. The reviewers agree the current manifest writers decode into structs and re-encode whole files, which can drop comments, unknown keys, unknown tables, ordering, formatting, and custom import options. The design's "preserve or refuse" language must become a concrete mechanism: either use a scoped TOML editor with golden tests for operator-authored manifests, or refuse the automated fix when preservation cannot be proven.
- [Blocker] The mutating doctor path lacks a concurrency policy. Codex and DeepSeek block on the absence of controller-active and concurrent-writer safety for `gc doctor --fix`, `gc import install`, lockfiles, and `.gc/system/packs/core`; Claude also asks for the live-controller/Core behavior to be specified. The design must either refuse mutating fixes while the controller or another pack mutation is active, or define a tested staging/locking/swap protocol that keeps readers seeing a complete old or complete new state.
- [Major] Core import removal and Core materialization ordering are not specified. The design allows redundant generated Core imports to be removed after required Core is present, but the doctor framework runs checks independently in registration order. The import-state fix needs an explicit precondition, dependency, or self-contained Core materialization step, with a test where Core materialization fails and the Core import remains untouched.
- [Major] The Core presence fix needs a no-op and atomic directory contract. Calling `MaterializeBuiltinPacks` currently implies a remove-and-regenerate path for required packs; under concurrent reads this can expose a missing or partial Core directory. The design should require zero content changes for a healthy Core pack and an atomic temp-directory swap or equivalent when repair is needed.
- [Major] Retired Maintenance runtime state is under-specified. Codex and DeepSeek both call out `.gc/runtime/packs/maintenance` handling, including JSONL archive/export state. The design must classify each retired runtime path as migrated, ignored legacy, or manual diagnostic before docs or fix behavior imply that state is gone.

**Disagreements:**
- Claude's verdict is `approve-with-risks`; Codex and DeepSeek both block. Assessment: this persona verdict is `block` because the unresolved items are not merely implementation details. They govern data preservation, partial migration, and controller-active safety for an automated doctor fix.
- The reviewers vary in severity for Core ordering/materialization. DeepSeek treats Core import ordering as a blocker, Codex lists it as a major risk, and Claude records it as missing evidence. Assessment: it is required design work, but the broader block is already justified by provenance, atomicity, and scoped-edit failures.
- DeepSeek proposes deferring all manifest writes until after preflight as the simplest safe route, while Codex allows a transaction/staging design, and Claude asks whether the implementation should use a preserving TOML editor or refuse auto-fix. Assessment: the design can choose the mechanism, but it must choose one testable mechanism before implementation.
- Runtime-state migration appears at different severities: Codex treats it as major under-classification, DeepSeek names the JSONL archive doctor as a concrete gap, and Claude focuses on preserving stale system/runtime pack directories. Assessment: this is not the leading blocker, but it is required evidence before declaring the Maintenance migration operator-safe.

**Missing evidence:**
- A provenance source proving a legacy Gastown, Maintenance, or Core import is generated and unmodified rather than an operator fork, copied example, edited system pack, or deliberate Core pin.
- A concrete immutable-pin check for `PublicGastownPackVersion`, including whether tags/branches are resolved to commits before any rewrite.
- An air-gapped or registry-unreachable failure contract that leaves `city.toml`, `pack.toml`, lockfiles, and installed pack directories byte-identical and prints manual recovery steps.
- Fault-injection coverage for failures after manifest staging, after any manifest write, after lockfile write, during install, and during installed-pack/Core directory replacement.
- Golden tests showing comments, import ordering, unknown TOML tables, unknown fields, custom import options, and unrelated imports survive import rewrite/removal.
- Tests for repeated healthy `gc doctor --fix`, two concurrent mutating doctor/import operations, controller-active reads during Core repair, and failed Core materialization before redundant Core import removal.
- A runtime-state migration table for `.gc/runtime/packs/maintenance`, including the JSONL archive doctor/export state and any doctor/export/reaper state files.

**Required changes:**
- Define the doctor fix safety contract as scoped, idempotent, preflight-before-mutation, failure-atomic or mutation-deferred, and explicit about allowed concurrency.
- Change legacy Gastown rewrite so reachability, immutable commit resolution, lockability, installability, and pack-subdir validation succeed before any existing-city manifest is written. On failure, leave files byte-identical and emit actionable manual guidance.
- Replace suffix-only legacy source classification with exact canonical matching plus generated/unmodified content provenance; route ambiguous or edited sources to manual diagnostics.
- Replace whole-file TOML re-encoding for doctor import fixes with a scoped preserving edit path, or require automated fix refusal when comments/unknown content/custom options cannot be preserved.
- Specify Core import removal ordering: require successful Core presence validation/materialization before removing a redundant generated Core import, and preserve the import if that precondition fails.
- Specify Core materialization under live-controller and repeated-run conditions, including zero-write healthy behavior and atomic directory replacement for repair.
- Add a runtime-state migration table for retired Maintenance paths, including explicit JSONL archive doctor behavior.
- Add the missing rollback, provenance, preservation, air-gap, concurrency, runtime-state, and Core-ordering tests as acceptance criteria for the doctor migration slice.
