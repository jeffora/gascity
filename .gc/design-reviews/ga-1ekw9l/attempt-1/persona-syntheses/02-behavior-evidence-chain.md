# Oleg Marchetti

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The behavior inventory can still pass at asset/file granularity instead of discrete-trigger granularity. Multi-trigger formulas, orders, scripts, prompt fragments, requester/detector paths, notifications, mail/nudge behavior, and escalation gates must each have rows, or the migration can silently drop behavior while satisfying a broad per-asset completeness check.
- [Blocker] The witness contract remains too weak for executable behavior. Source assertions and static scanners can still stand in for old or new behavior proof where reviewers expect normal-runtime execution witnesses, packcompat assertions that actually fire the trigger, golden transcripts, or explicit approved-delta records.
- [Blocker] Full implementation approval is not justified while required AC6, AC7, and AC14-AC17 evidence artifacts are absent. The current repository has only the requirements and implementation plan for this work, and the current public pack pin is pre-evidence; downstream work must be limited to prerequisite evidence production until the proof artifacts validate.
- [Major] Cross-repo manifests can drift. Gas City-side ledgers/manifests and public `gascity-packs/gastown` evidence files need a shared behavior-id contract and bidirectional parity validator over row identity, old source digest, public commit, new witness digest, owner, semantic delta, and packcompat transcript.
- [Major] Packcompat needs stronger per-row proof. Compatibility-pin and activation-pin modes must prove resolution provenance from the pinned public pack, reject in-tree or synthetic fallback, and assert observable effects such as route metadata, mail/nudge recipients and payload shape, rendered prompt variables, script exit codes, diagnostics, runtime-state mutations, and success/warning/failure/recovery branches.
- [Major] The historical old baseline is unsafe unless it is pinned to an immutable pre-migration commit or tag, recorded in the manifest, and validated by CI to predate any deletion.
- [Major] The schema and extractor taxonomy omit required AC7 dimensions. Outcome paths, recovery flows, template fragments, template variables, prompt/render variables, runtime-state paths, and side-effect envelopes need explicit row fields, extraction rules, and witness rules.
- [Major] The explicit old-witness mapping is too narrow. Reviewers identified additional `examples/gastown/*_test.go` files beyond the two named in the plan; each legacy witness surface needs a new owner, executable witness, or approved removal/delta record.
- [Major] The deletion slice does not re-assert the core behavior proof at the exact point of source deletion. Behavior-evidence freshness and activation-mode packcompat must be must-pass gates for the deletion merge.

**Disagreements:**
- Claude and DeepSeek V4 Flash return `block`; Codex returns `approve-with-risks`. Assessment: the persona verdict is `block` because the unresolved issues are the central safety contract for this lane, not peripheral polish.
- Claude treated source assertions as possible lower-assurance old-side evidence only with an approved record; Codex and DeepSeek pushed to disallow source/static proof for executable triggers. Assessment: executable behavior should require execution-level witnesses where execution is possible; source/static evidence should be limited to non-executable prose, generated references, approved retirement, or explicit semantic-delta records.
- DeepSeek framed the cross-repo bootstrap as a circular deadlock, while Claude and Codex emphasized manifest drift, provenance, and public-pin evidence. Assessment: these are compatible risks; the plan needs both a local/draft promotion workflow and a final immutable reconciliation gate.
- DeepSeek credited durable packcompat isolation more strongly than Claude and Codex. Assessment: the design direction is sound, but compatibility-pin mode still needs explicit per-assertion provenance, digest, subpath, and fallback-negative checks.
- Codex treated the absence of current proof artifacts as acceptable only for prerequisite-producing decomposition, while Claude and DeepSeek treated weak evidence contracts as blocking before behavior deletion. Assessment: decomposition may proceed only for proof-substrate work unless the artifact clearly says full behavior-changing implementation is blocked until the evidence exists.

**Missing evidence:**
- Checked-in Gas City support artifacts such as `behavior-preservation-manifest.yaml`, `asset-migration-ledger.yaml`, `public-gastown-pin-ledger.yaml`, version-skew data, and acceptance proof matrices.
- Public `gascity-packs/gastown` evidence artifacts at the exact consumed commit, including behavior preservation, pin ledger, ownership rows, pack digest, behavior-manifest digest, and packcompat transcripts.
- A fixed immutable old-baseline commit or tag and CI proof that the generator uses that exact baseline for historical scans.
- A worked end-to-end behavior row showing old trigger, old witness, behavior id, new owner, public-pack path, new executable witness, public commit, digest, packcompat assertion, pin-ledger entry, and approved delta if any.
- Trigger extraction parser details, behavior-bearing predicates, positive/negative fixtures, and row-to-trigger reconciliation for formulas, orders, prompt/template fragments, shell branches, doctor checks, route metadata, notifications, mail/nudge behavior, requester/detector fields, hook overlays, and runtime-state paths.
- A witness-strength matrix mapping behavior kinds to minimum admissible proof types.
- Cross-repo bootstrap commands for local multi-workspace or draft-pin validation before immutable public SHAs are finalized.
- A validator that compares Gas City and public-pack manifests bidirectionally and fails on missing rows or drifted source digests, witness digests, owners, deltas, or packcompat transcript references.
- Packcompat transcript assertions for side effects and branch behavior, not just formula/script/prompt existence.
- Explicit mapping for all legacy `examples/gastown/*_test.go` witness files or approved removal/delta records.

**Required changes:**
- Define behavior row identity as one row per discrete trigger or behavior-bearing branch, with CI-enforced row-to-trigger reconciliation so multi-trigger assets cannot collapse into one row.
- Add outcome-path, recovery-flow, template-fragment, template-variable, prompt/render-variable, runtime-state, requester, detector, notification, mail, nudge, route-metadata, escalation, and script-branch categories to the behavior taxonomy.
- Add a witness-strength matrix. Active executable behavior must use execution-level witnesses; source assertion and static scanner evidence must be limited to non-executable or explicitly approved retirement/equivalence cases.
- Strengthen `pack-evidence validate` and `test/packcompat` so every row asserts provenance and observable execution effects, including negative fallback checks against in-tree or synthetic sources.
- Add a shared behavior-id contract and reconciliation gate across Gas City ledgers, public Gastown manifests, public ownership files, pin ledgers, and packcompat transcripts.
- Document the cross-repo bootstrap and immutable promotion flow from frozen old Gas City baseline to draft public Gastown evidence, draft transcripts, final public commit, digest recalculation, pin-ledger update, and fail-closed row-change handling.
- Pin the historical baseline to an immutable pre-migration commit or tag, record it in the manifest, and require CI to prove it predates source deletion.
- Expand old-to-new witness mapping to all legacy `examples/gastown/*_test.go` files relevant to this lane.
- Add behavior-evidence freshness and activation-mode packcompat to Slice 7's must-pass-before-merge deletion gates.
- Clarify readiness in the artifact: full implementation remains blocked until AC6, AC7, and AC14-AC17 evidence artifacts exist and validate; only prerequisite evidence-substrate work may be decomposed before then.
