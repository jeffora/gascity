# Sofia Khoury - Claude

**Verdict:** block

I reviewed strictly in lane against the current design: `gc doctor` / `--fix`
idempotency, legacy import rewrites, custom-data preservation, and operator-safe
diagnostics. The intended mutation architecture is genuinely strong and several
of my prior-iteration concerns are now closed (the generated-vs-custom *pack*
discriminator is defined at 2577-2578; healthy `--fix` is a stated zero-write
golden at 2577). It still blocks on one self-contradiction: the section an
implementer actually builds the Core-presence check from wires **plain**
`gc doctor` (no `--fix`) to the materializing/pruning loader, contradicting the
read-only boundary the design itself establishes — on the exact no-mutation path
my lane protects.

**Top strengths:**
- Mutation discipline is correct and well-guarded: coordinator-only writes, a
  crash-released advisory city lock (an fd flock, not a PID/status file — honors
  the no-status-files rule), post-lock digest refresh, staged-outside-active,
  compare-before-rename on every target, and span/CST-preserving TOML edits that
  *refuse* with manual steps rather than whole-file re-encode (2556-2562,
  3103-3134). Generated-vs-custom *pack* detection is concretely defined (digest,
  generation marker, lock metadata, operator-edit classification; custom forks →
  manual) (2577-2578).
- Stale-directory containment and custom-data preservation are airtight:
  `.gc/system/packs/{maintenance,gastown}` are ignored, "never delete; optional
  report only" (2193-2194), enforced centrally through `internal/packsource`
  (2185-2189, 2001-2004); unexpected operator-edited/unclassifiable files are
  quarantined, not pruned (149-150, 1557-1562, 2091); runtime-state migration is
  copy-only with a completion marker and post-marker divergence ignored
  (1968-1978, 560-566). Custom Core, custom forks, and air-gapped migrations are
  routed to manual diagnostics, not treated as silently fixable (3098-3101,
  2197). Red flags 2 and 3 are cleanly closed.
- Rewrite-to-remote is preflight-gated and refuses rather than forces (lane Q3):
  the immutable `sha:` pin must be reachable via network or a *digest-validated*
  ordinary remote-cache hit, installable, and lockable before any write
  (1956-1959, 3112-3117, 2240); offline is an explicit failure with no embedded
  or synthetic fallback (2136, 2241, 2958), and the pin validator observes actual
  checkout/cache content digests, not load-success (2242-2244).

**Critical risks:**
- **[Blocker] The plain-doctor Core-presence check is specified to call the
  materializing/pruning loader, contradicting the read-only doctor boundary and
  the coordinator-only-mutation rule.** §Core Presence Doctor defines the *check*
  that runs on plain `gc doctor` — it returns Error/Warning (3063-3064) and sets
  `FixHint` to run `gc doctor --fix` (3068), i.e. the no-`--fix` report path — yet
  instructs it to "Load resolved config through `internal/systempacks.LoadRuntimeCity`"
  (3062) and validate "using the same strict integrity gate as runtime config
  loading" (3060). But `LoadRuntimeCity` is the *materializing* entrypoint
  (3004-3010), and required-pack materialization "repairs missing/corrupt expected
  files and prunes or quarantines unexpected effective files before validation"
  (3008-3010, 1557-1562); `LoadRuntimeCityNoRefresh` is its read-only twin. This
  contradicts the Read-Only Doctor Diagnostic Boundary — plain `gc doctor`
  "may call only" `ValidateRequiredFileSetsNoRefresh` / `LoadRuntimeCityNoRefresh`
  and "must not call materializing runtime loaders … repair helpers …
  quarantine/prune writers" (2109-2122) — reaffirmed at attempt-17 ("Plain
  `gc doctor` is report-only … must not materialize, repair … quarantine/prune
  files", 2550-2556) and the failure-semantics table ("Plain `gc doctor` → … no
  mutation", 2087-2094), with "All mutation happens only through
  `doctor.MutationCoordinator` and only after `--fix` is present" (2124-2125). As
  written, a no-`--fix` `gc doctor` on a city whose generated Core dir has drifted
  would prune/quarantine Core files — a mutation outside `--fix`, outside the
  coordinator, with no controller-active refusal (that is `--fix`-only, 3122-3123)
  and no advisory lock. This breaks lane Q1 (a diagnostic that prunes is not a
  no-op) and trips red flag 1 (mutation beyond scope — without even `--fix`). The
  prose-precedence rule (2604-2606) only demotes prose under *generated rows*;
  this is prose-vs-prose and stays unreconciled in the one section that concretely
  specifies the check.
- **[Major] Import removal/rewrite is gated on the *target source path*, not on the
  *import edge's* authorship.** Maintenance imports are auto-removed "when the
  source is `.gc/system/packs/maintenance` or `examples/gastown/packs/maintenance`"
  (3091); redundant Core imports are removed when they point at generated/legacy
  Core paths (3094-3097). The generated-vs-custom discriminator (digest, generation
  marker, lock metadata, operator-edit, 2577-2578) classifies *pack directories*;
  an `[[imports.*]]` table in `city.toml` / rig `pack.toml` carries no generation
  marker of its own. An operator who hand-authored an import whose target matches a
  generated/system pattern therefore has that edge silently removed/rewritten. Lane
  Q2 is half-answered: custom *pack content* is protected, but a custom *import
  edge* with a generated-looking target is not.
- **[Major] Downgrade-then-`--fix` with an already-released old binary is
  unaddressed.** The design freezes unsafe legacy fixes in *new* slice-2/3 binaries
  (3397-3399) and gives rollback *readability* guidance — migrated manifests "must
  remain readable by old binaries" and old binaries may keep reading retained
  legacy runtime state (566-568, 3471). It never addresses the inverse: an operator
  downgrades to the shipped pre-migration binary and runs *its* `gc doctor --fix`,
  which predates the coordinator, the span-preserving editor, the public-pin
  semantics, and the runtime-state marker. Any fix that old binary applies uses the
  old whole-file rewrite path, so it can lossily re-encode `city.toml` / rig
  `pack.toml` (destroying comments/formatting and potentially the migration's
  import wiring) while ignoring the marker and migrated state. The compatibility
  matrix (3463-3471) covers rollback *readability*, not old-binary *mutation*.
  Precondition is downgrade + run `--fix`, but the consequence is silent operator
  data loss, so it sits squarely in lane; the only mitigations (shape migrated
  manifests to be inert + release-notes guidance) are named nowhere.
- **[Minor] Controller concurrency leaves a torn multi-file read window.** Exclusion
  is detection-then-refuse from live runtime state (3122-3123, 180-181) plus
  compare-before-rename (3124-3125), but nothing states the *controller* (or its
  `MaterializeRequiredPacks` self-heal) acquires the coordinator's advisory city
  lock — "concurrent self-heal contention through the advisory directory lock"
  appears only as a fixture line (2007-2008), and the live fact that binds a
  controller to *this* city is unnamed. Multi-file atomicity is guaranteed only
  against *preflight* failure (3118-3121), not against a controller that *starts
  mid-publish* and reads a torn set (new `city.toml` + old rig `pack.toml`).
  No-refresh reload + NDI converges this, hence Minor, but lane Q1 explicitly names
  "concurrent runs with a controller active," so the bound should be stated.

**Missing evidence:**
- No named test asserts that plain `gc doctor` (without `--fix`) performs zero
  filesystem writes against a drifted / missing-file / extra-effective-file Core.
  The zero-write golden (2577, 3129) is scoped to `--fix`; the test that would have
  caught the Blocker is absent from the doctor failure-injection matrix (2130-2143).
- Whether the per-pack repair-status / "generation marker" (2577-2578) is an
  in-memory field of the resolver-produced `RequiredSystemPackParticipation` record
  or is persisted into `.gc/system/packs/core`. If persisted and rewritten per run,
  the "byte-identical after `gc doctor --fix`" guarantee for healthy cities
  (1977-1978, 3129) is unprovable. The design implies in-memory but never states the
  Core system-pack directory holds no per-run-mutated marker.
- The exact live runtime fact that binds a controller to this city, and whether the
  controller takes the same advisory lock as the coordinator.

**Required changes:**
- Reconcile 3060-3062: the plain-doctor Core-presence *check* must read via
  `LoadRuntimeCityNoRefresh` / `ValidateRequiredFileSetsNoRefresh` (2112-2113) and
  route every materialize/repair/prune/quarantine strictly through the coordinator
  under `--fix`. If a required-Core diagnostic cannot be computed read-only, report
  `diagnostic_unavailable_without_fix` (2121-2122) rather than calling the
  materializing loader. Delete the `LoadRuntimeCity` / "same strict integrity gate
  as runtime config loading" wording from this section so the implementer cannot
  build the unsafe path. Add a failure-injection row proving plain `gc doctor`
  against missing/corrupt/extra-file Core emits a diagnostic with zero mutation.
- Define import-*edge* provenance for removal/rewrite: auto-act only on edges proven
  generator-emitted (recorded in a generated-import ledger/lock, or matching an exact
  `gc init` template), and treat any operator-authored import — even one whose target
  matches a generated/system pattern — as manual/diagnostic. State this for both
  Maintenance removal (3091) and redundant-Core-import removal (3094-3097).
- Add a version-skew matrix row and fixture for old-binary `gc doctor --fix` against
  a new-binary-migrated city. If it cannot be made inert, name the explicit downgrade
  limitation and manual-recovery path in release notes, and shape migrated `city.toml`
  / rig `pack.toml` so the old `--fix` neither classifies the public import as a
  rewritable legacy local import nor whole-file re-encodes the manifest.
- State the controller-concurrency contract explicitly: either the controller
  acquires the same crash-released advisory city lock (closing the torn-read window),
  or document why detection + compare-before-rename + no-refresh reload is sufficient
  and bound the torn-multi-file-read window. Name the live fact that identifies a
  controller for this city.

**Questions:**
- Does the Core-presence check at 3062 intend the no-refresh loader, or is plain
  `gc doctor` expected to materialize Core? If the latter, how is that squared with
  "All mutation happens only through `doctor.MutationCoordinator` and only after
  `--fix` is present" (2124-2125)?
- The check must report "Core absent from typed resolved-config participation"
  (3066-3067), yet the no-refresh loader *fails closed* on a city missing required
  Core (table rows 2087, 2094). Does `LoadRuntimeCityNoRefresh` return a structured
  "missing participation" verdict the read-only check can surface (as 2094 implies),
  or only a load error — and is that ambiguity what drove this section to reach for
  the materializing `LoadRuntimeCity`?
- Is the per-pack generation marker / freshness state ever written under
  `.gc/system/packs/core`, or is it purely an in-memory `RequiredSystemPackParticipation`
  field? This determines whether healthy-city byte-identity is actually achievable.
