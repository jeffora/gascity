# Sofia Khoury

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash (present in the `_gemini.md` artifact; no separate `_deepseek.md` artifact was present)

**Consensus findings:**
- [Blocker] The Core Presence Doctor section still contradicts the read-only `gc doctor` boundary. Claude, Codex, and DeepSeek all found that the plain diagnostic path is specified to call `internal/systempacks.LoadRuntimeCity`, even though that loader is described elsewhere as materializing, repairing, pruning, or quarantining. Plain `gc doctor` must use no-refresh/read-only APIs and must not mutate before `--fix` or outside `doctor.MutationCoordinator`.
- [Major] Controller and doctor concurrency are underspecified. The design names advisory locking and live process discovery, but does not define the same-city runtime facts, path canonicalization, controller participation in the lock, or the behavior when a controller starts during a multi-file publish.
- [Major] Import and custom-source migration safety needs a firmer provenance boundary. Claude flags that generated-looking import edges can be operator-authored; Codex flags an undefined separate custom-source migration command; DeepSeek flags custom fork classification sensitivity. Automated rewrite/removal needs proof that the import edge or pack content is generator-owned, with manual diagnostics on uncertainty.
- [Major] Reverse version-skew remains unaddressed. Claude and DeepSeek both identify the unsafe case where an older already-released `gc doctor --fix` runs against a newly migrated city and may whole-file rewrite manifests or ignore migration markers.
- [Major] Healthy `gc doctor --fix` idempotency must be a filesystem zero-write, not only byte-identical final content. DeepSeek calls out mtime/inode churn and watcher reload storms; the design should require digest comparison before publish and skip write/rename syscalls when live bytes already match.
- [Minor] Preflight ordering for public pin and cache validation needs explicit safety language. DeepSeek argues the network/cache/installability checks must complete before any edit begins; Claude agrees the remote/cache gate is conceptually strong but asks for exact operator-safe behavior on offline/cache failure.

**Disagreements:**
- Severity differs on concurrency. DeepSeek treats controller lock non-participation as a blocker; Claude and Codex rate it below the shared Core Presence Doctor blocker. Assessment: the design is blocked by the plain-doctor mutation contradiction regardless; concurrency is a required major fix before approval.
- DeepSeek elevates zero-write idempotency to a major safety issue, while Claude treats the healthy zero-write guarantee as missing evidence and Codex mostly frames it as a test/API distinction. Assessment: require zero-write semantics and tests because byte-identical rewrites can still affect live controllers.
- Codex focuses on the undefined custom-source migration command; Claude focuses on import-edge authorship; DeepSeek focuses on custom fork classification. Assessment: these are the same operator-data-preservation boundary and should be resolved by explicit provenance proof plus a scoped or deferred custom-source command.
- Claude suggests old-binary inertness or release-note/manual-recovery guidance; DeepSeek suggests a fail-fast migrated manifest marker; Codex does not foreground this risk. Assessment: add the reverse version-skew fixture first, then choose an implementable mitigation based on what the old binary can actually do.

**Missing evidence:**
- A named plain-`gc doctor` golden/failure-injection test proving no filesystem changes for missing, corrupt, stale, partial, and extra-effective Core states.
- A loader/API contract proving report-only checks use `LoadRuntimeCityNoRefresh`, `ValidateRequiredFileSetsNoRefresh`, raw import classification, or equivalent read-only APIs, while materializing loaders are `--fix`-only.
- A concrete same-city runtime discovery contract for controller-active and concurrent-doctor refusal, including symlinked/moved city paths and restricted process namespaces.
- A healthy-city zero-write test that detects mtime/inode-changing republish behavior across manifests, lockfiles, installed pack metadata, and runtime-state paths.
- A reverse version-skew fixture for an older `gc doctor --fix` against a new-binary-migrated city.
- A generator-owned import-edge or pack-content provenance mechanism, plus behavior for unverifiable same-path directories, edited local forks, and custom public sources.
- A complete custom-source migration command contract, or an explicit statement that such migration is out of scope and must remain manual.

**Required changes:**
- Rewrite the Core Presence Doctor contract so plain `gc doctor` uses only no-refresh/read-only validation and diagnostic APIs. If a diagnostic cannot be computed read-only, report it as unavailable without `--fix` instead of calling a materializing loader.
- Split report-only and fix-intent behavior explicitly. All materialization, repair, cache promotion, lock generation, quarantine/prune, runtime-state migration, and config writes must run only through `doctor.MutationCoordinator` after `--fix` preflight.
- Add zero-write plain-doctor tests for missing/corrupt/stale/partial Core and healthy-`--fix` tests that prove byte-identical targets are not rewritten.
- Define the concurrency contract: advisory lock path and scope, whether the controller takes a shared lock or is refused by live process discovery, how city paths are canonicalized, and how torn multi-file reads are prevented or bounded.
- Define import-edge and pack-content provenance for automated rewrite/removal. Operator-authored, unverifiable, or diverged inputs must become manual diagnostics, not automated fixes.
- Remove the undefined custom-source migration exception from `gc doctor --fix`, or specify a separate command with ownership, preflight, rollback/resume, confirmation, and test requirements.
- Add a reverse version-skew matrix row and fixture. If old-binary mutation cannot be made inert, state the downgrade limitation and manual recovery path in release notes and shape migrated manifests to fail clearly where possible.
- State that remote public pin reachability, digest-validated cache availability, installability, lockability, and dry-run resolution complete before any filesystem write or temporary target is created.
