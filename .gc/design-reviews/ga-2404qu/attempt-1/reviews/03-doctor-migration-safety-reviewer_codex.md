# Sofia Khoury - Codex

**Verdict:** block

**Top strengths:**
- The design treats `gc doctor --fix` as a staged mutation coordinator, with preflight, post-lock digest refresh, compare-before-rename, final validation, and byte-identical healthy no-op requirements.
- Legacy Gastown and Maintenance import rewrites are limited to generated/example provenance, while custom forks, edited retired directories, and air-gapped missing pins are diagnostic/manual.
- Stale `.gc/system/packs/*` and `.gc/runtime/packs/maintenance` paths are explicitly preserved, ignored, or copied with markers; the design does not rely on deleting operator-owned state.

**Critical risks:**
- [Blocker] The Core Presence Doctor section contradicts the read-only doctor boundary. The design says plain `gc doctor` must not materialize, repair, promote cache entries, write locks, quarantine/prune files, migrate runtime state, or rewrite TOML, and may use only no-refresh validation/diagnostic APIs (`design.md:2550`, `design.md:2106`). But the later Core Presence Doctor check says to load resolved config through `internal/systempacks.LoadRuntimeCity` (`design.md:3061`), which earlier sections define as the runtime path that can repair before publish. An implementer following the later section could make plain `gc doctor` mutate `.gc/system/packs/core` or cache/lock state before `--fix`, violating the operator-safety contract.
- [Major] The custom-source escape hatch is not designed. The doctor safety contract says operator forks, edited local packs, or custom public sources are diagnostic/manual unless the user explicitly opts into a separate migration command (`design.md:3135`). That separate command has no ownership, preflight model, rollback behavior, provenance proof, or UX contract in this document. Leaving it as an undefined exception risks implementers smuggling non-generated custom migrations into `gc doctor --fix`.
- [Minor] Live-controller and concurrent-doctor detection is directionally right but still under-specified at the evidence boundary. The design requires runtime facts and live process state rather than status files (`design.md:1667`, `design.md:1953`), but it does not pin how "same city" is identified across symlinks, moved directories, multiple rigs, or stale process arguments. That could produce either unsafe false negatives or unnecessary refusal loops.

**Missing evidence:**
- A named plain-doctor golden/failure-injection test proving `gc doctor` on missing, corrupt, partial, and extra-file Core leaves `.gc/system/packs`, lockfiles, manifests, and caches byte-identical.
- A loader/API distinction in the Core Presence Doctor section that maps report-only checks to `ValidateRequiredFileSetsNoRefresh` / `LoadRuntimeCityNoRefresh` and reserves materializing `LoadRuntimeCity` for staged `doctor.MutationCoordinator` validation under `--fix`.
- A concrete same-city runtime-discovery contract for controller-active and concurrent-doctor refusal.
- Either a full spec for the separate custom-source migration command or an explicit statement that it is out of scope and unavailable in this migration.

**Required changes:**
- Change the Core Presence Doctor implementation text to use no-refresh/read-only APIs for plain `gc doctor`; state that materializing/repairing loaders may run only inside the `doctor.MutationCoordinator` after `--fix`, staging, and preflight.
- Add explicit tests for plain `gc doctor` byte-identical behavior on missing/corrupt/stale/partial Core, retired active imports, stale locks, and preserved stale directories.
- Remove the undefined "separate migration command" exception from the doctor fix contract, or add a separate command contract with provenance, preflight, rollback, and operator-confirmation requirements outside `gc doctor --fix`.
- Define the runtime facts used to detect a controller for the same city and a concurrent doctor mutation process, including path canonicalization and at least one symlinked-city test.

**Questions:**
- Should plain `gc doctor` ever be allowed to repair Core materialization implicitly, or is the intended invariant strictly "no bytes change without `--fix`" for every doctor mode in this migration?
- If custom public sources need migration support, is that a new command in this scope, or should the migration explicitly defer it to manual operator instructions?
