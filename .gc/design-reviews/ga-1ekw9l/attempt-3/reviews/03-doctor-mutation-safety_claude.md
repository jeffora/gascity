# Leah Okafor - Claude

**Verdict:** approve-with-risks

> Lane: doctor fix-coordinator atomicity, byte-preserving TOML, concurrency with
> live controllers, advisory locks, idempotent recovery. Reviewed against the
> current `implementation-plan.md` (528 lines, `updated_at: 2026-06-09T01:20:00Z`)
> — §"Doctor And Runtime-State Mutation Safety" (243–273), the Doctor/runtime
> recovery records in §"Data And State" (372–381), and the Doctor tests in
> §"Testing" (429–439). I verified the repo anchors rather than trust the prose.
>
> Output written to the live iteration-3 reviews dir (`attempt-3/`) beside the
> Codex sibling; my routed bead `ga-0ufwnr` carries `gc.attempt=1` while its
> iteration is 3 (logical `ga-pm6y5f` is `attempt=3`), so the literal
> `attempt-${gc.attempt}` path would overwrite the unrelated iteration-1 review.
> Reported via `design_review.output_path`.

**Schema conformance:** Conforms. Required front matter and the eight ordered
top-level sections are present; `Open Questions` is `None`. The doctor-safety
content is appropriately split across Proposed Implementation, Data And State,
and Testing rather than appended as review prose.

**Top strengths:**
- The live-controller race is closed the right way (red flag #3): the coordinator
  "acquires a crash-released city advisory lock before digest preflight," re-runs
  digest+provenance validation *after* the lock when a report-only phase read
  earlier (a real TOCTOU mitigation), and refuses automatic fix when a controller
  for the same city is running, "discovered from live runtime state rather than
  status files" (253–257) — consistent with the project's no-status-files rule.
  A crash-released advisory flock already exists to build on
  (`internal/events/recorder.go`, bounded-wait cross-process flock).
- Cross-file atomicity is *not* claimed from per-file renames (directly answering
  red flag #1): multi-file fixes "write durable recovery state before the first
  publish step, stage all edits, re-read target digests before each temp-file
  rename, and define a single commit point. A process death before commit reruns
  deterministically or rolls back from recovery state. A process death after
  commit converges by revalidating Core participation, public-pin installability,
  lock contents, and runtime-state marker state" (259–264). The recovery record
  carries "publish order, commit point, completed steps, rollback instructions"
  (372–375), which is a genuine idempotent-resume contract.
- Concurrent old/new binaries are handled (lane Q3): the runtime-state migration
  marker records "old-binary post-marker write detection," and an old-binary
  write after the marker triggers a version-skew diagnostic requiring manual
  reconciliation or a deterministic re-upgrade (266–273). Byte-preservation has a
  refuse-rather-than-rewrite contract in Testing: "Scoped TOML edits preserve
  comments, unknown tables, unknown fields, array order, formatting, and
  unrelated lock entries; otherwise automatic fix refuses" (432–434), with
  per-step failure-injection rerun/rollback tests (435–436).

**Critical risks:**

- **[Major] The "single commit point" for a multi-file rename set is asserted but
  its mechanism is unnamed.** POSIX renames are atomic per file, not across
  files; the plan says "define a single commit point" and records a "commit
  point" field (262, 374) but never says *which* durable write linearizes the
  multi-file change — e.g. "all data temps are written and fsync'd, then a single
  marker rename is the commit, and all post-marker renames are idempotent
  roll-forward." Without naming the linearizing operation and the ordering
  invariant, an implementer can satisfy "recovery state + per-file renames"
  while still having a window where a crash leaves some target files published
  and others not, with no single before/after boundary — re-introducing exactly
  red flag #1. Pin the commit artifact and the strict before-commit (rollback) /
  after-commit (roll-forward) classification of every staged rename.

- **[Major] "Proof the target is system-generated before rewriting" is not an
  explicit fix-refusal rule (red flag #2).** The coordinator is the sole writer
  of import rewrites (250–251) and `internal/packsource` can classify a source as
  "retired custom/fork" vs "retired generated/example" (222–228), so the
  capability exists — but the plan never states that doctor *refuses* to rewrite
  an import the classifier marks custom/fork or cannot prove is system-generated.
  Today there are 8+ `Check.Fix(ctx *doctor.CheckContext)` implementations in
  `cmd/gc/doctor_v2_checks.go` (e.g. `v2ImportFormatCheck`,
  `v2DefaultRigImportFormatCheck`, `v2PackSourcesCheck`) that mutate directly;
  the plan must say which migrate to FixIntent and that each import/pack rewrite
  consults the classifier and refuses on custom/fork or unproven provenance.

- **[Minor] Commit granularity across a multi-check `gc doctor --fix` run is
  unspecified.** The coordinator gives per-fix atomicity, but a single doctor run
  may apply several coordinator fixes plus several still-direct checks. The plan
  should state whether one run is one transaction or N independent idempotent
  fixes, and confirm a crash mid-run leaves each completed fix independently
  consistent (NDI) rather than a half-applied run.

- **[Minor] The byte-preserving edit approach is only a test assertion, not a
  named implementation.** §432–434 requires preserve-or-refuse, but Proposed
  Implementation never names a trivia/CST-preserving TOML editor (vs full
  re-serialization). Naming it would make the "refuse rather than rewrite whole
  files" guarantee (lane Q2) decomposition-ready instead of test-discovered.

**Missing evidence:**
- The concrete commit-point artifact and the rule classifying each staged rename
  as pre- or post-commit for deterministic rerun/rollback.
- The list of existing `doctor_v2_checks.go` `Fix` methods that migrate to
  FixIntent vs are refused for automatic fix, and which writes route through the
  coordinator.
- The provenance test that proves an import/pack target is system-generated
  before a rewrite, and the refuse path for custom/fork.
- Whether a `gc doctor --fix` run is one transaction or a sequence of independent
  idempotent fixes, and the crash semantics between fixes.
- The named byte-preserving TOML edit mechanism behind the preserve-or-refuse
  test.

**Required changes:**
- Specify the single commit point: name the linearizing durable write (e.g. a
  marker rename), require all data temps fsync'd before it, and classify every
  staged rename as pre-commit (rollback on crash) or post-commit (idempotent
  roll-forward). Reconcile "single commit point" with multi-file rename reality
  explicitly.
- Add an explicit refusal rule: doctor refuses to rewrite any import/pack source
  the `internal/packsource` classifier marks custom/fork or cannot prove
  system-generated; enumerate which current `Check.Fix` methods become FixIntent.
- State the doctor-run transaction granularity and the inter-fix crash
  semantics.
- Name the trivia-preserving TOML editor that backs the preserve-or-refuse
  contract.

**Questions:**
- What single durable operation is the commit point for a multi-file fix, and
  which renames are pre- vs post-commit?
- Does doctor consult `internal/packsource` and refuse before rewriting a
  custom/fork import, or can it rewrite any matching import?
- Is one `gc doctor --fix` invocation one transaction or N independent
  idempotent coordinator fixes?
