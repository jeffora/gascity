# Yuki Hayashi

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash content present in the `_gemini.md` artifact

**Consensus findings:**
- [Blocker] Activation sequencing is internally unsafe or contradictory. Claude and Codex both describe the intended safe boundary as an activation pin consumed by the same candidate tree that removes bundled Maintenance, while the rollout prose still leaves room for an intermediate state where `PublicGastownPackVersion` switches to activation before Maintenance is removed. If that intermediate can run, fresh `gc init --template gastown` can load activation public Gastown while bundled Maintenance remains active, recreating the duplicate-active state the rollout is meant to avoid.
- [Blocker] Old-binary fail-closed behavior depends on `RequiresGC` enforcement that the DeepSeek artifact says is not present. The activation pack can declare a minimum Gas City version, but if the loader only parses `requires_gc` and never semver-validates it, an old binary can load the activation-pinned public pack while still force-including bundled Maintenance. That turns downgrade guidance into silent duplicate behavior.
- [Blocker] Rollback and downgrade behavior is not executable enough for the riskiest path. Claude requires a decided answer on whether activation is reversible or one-way; Codex requires final matrix rows and rollback evidence for old binary plus activation; DeepSeek identifies a downgrade-then-reupgrade split-brain where legacy Maintenance state can advance after the first Core copy, then the new binary skips re-copying because Core state already exists.
- [Major] The staged rollout depends on a compatibility public Gastown commit that omits every asset colliding with bundled Maintenance while still preserving behavior. Claude identifies this as the load-bearing cross-repo precondition for avoiding a flag day, and DeepSeek adds that duplicate checks must be paired with no-gap completeness checks over the union of compatibility public Gastown, bundled Maintenance, and host Core.
- [Major] The `v1.2.1 + activation pin` matrix result remains unresolved. The design allows either support or a named one-way boundary, but the lane needs the observed result, tested downgrade or recovery procedure, and mandatory ledger evidence before activation can ship.
- [Major] Public pin evidence is strong in intent but inconsistent in final shape and durability. Codex finds stale two-row ledger prose conflicting with the required three-row model (`current_baseline`, `compatibility`, `activation`), while Claude and DeepSeek both require durable reachability proof so pinned commits cannot be orphaned by squash/rebase merges or branch deletion.
- [Major] Existing-city availability changes need operator-facing treatment, not only fresh-init notes. Moving Gastown from embedded bytes to a fetched public pack means cache-miss, fresh-host, lock-refresh, doctor repair, and air-gapped upgrade paths depend on a reachable or pre-populated ordinary remote cache.
- [Minor] Fresh-init behavior during rollout windows is implied but not stated. Codex and Claude both ask for per-window assertions covering pre-pin-update baseline, compatibility-pin adoption, activation-pin adoption, and rollback.

**Disagreements:**
- All three source verdicts are `approve-with-risks`, but DeepSeek labels two risks as blockers and all reviewers require activation/rollback evidence before the risky slice. Assessment: this persona synthesis blocks design approval because the open items affect release sequencing and downgrade correctness, not just implementation follow-through.
- Claude and Codex read the desired activation boundary as atomic, while later rollout wording appears to leave activation pin update and Maintenance removal separable. Assessment: the design must remove the ambiguity. Either the pin switch and Maintenance removal are one candidate gate, or the activation pin must prove safe coexistence with bundled Maintenance.
- DeepSeek says binary downgrade is structurally supported because old binaries reintroduce bundled Maintenance and legacy state is preserved; Claude and Codex still find the operator recovery contract incomplete. Assessment: preservation is not enough unless the design proves doctor-mutated locks, manifests, and runtime state remain readable or recoverable after downgrade and re-upgrade.
- DeepSeek proposes timestamp, monotonic counter, or epoch-based state reconciliation; this may be more implementation-specific than the other reviews require. Assessment: the design must require a monotonic reconciliation mechanism, but the exact representation can remain an implementation choice if it proves no state is lost.
- The third available artifact is named `_gemini.md` but identifies itself as DeepSeek V4 Flash. Assessment: include it as DeepSeek V4 Flash, since its content and upstream dependency identify that source; no separate `_deepseek.md` artifact is present for this attempt.

**Missing evidence:**
- A sample `public-gastown-pins.yaml` with all three phase rows: `current_baseline`, `compatibility`, and `activation`, plus exact validator-required fields.
- A sample `slice-gates.generated.yaml` activation row showing whether `PublicGastownPackVersion` update and Maintenance removal are one gate or two.
- A negative test that fails if the activation pin is selected while the production loader still includes bundled Maintenance, unless that activation pin has explicit current-loader duplicate and completeness evidence.
- Proof that `requires_gc` is semver-enforced during pack resolution and produces fail-closed downgrade guidance before unsupported binaries load activation-pack assets.
- Evidence that the compatibility public Gastown commit can omit colliding active assets while the union of compatibility public Gastown, bundled Maintenance, and host Core still preserves the target behavior manifest.
- The observed `v1.2.1 + activation pin` result, including content-digest evidence, duplicate/completeness matrix evidence, and the decision on whether activation is one-way.
- A generated downgrade or rollback transcript for old `v1.2.1` after `gc doctor --fix` has rewritten manifests, imports, locks, or runtime state.
- State migration evidence proving downgrade-then-upgrade does not lose JSONL export cursor progress, spawn-storm counts, message-delivery offsets, or other state advanced by the old binary.
- CI evidence that every consumed or historical public Gastown pin remains reachable from the default branch or a protected immutable tag in `gascity-packs`.
- A no-gap behavior manifest assertion proving the union of active assets across compatibility public Gastown, bundled Maintenance, and host Core matches the target behavior set during slices 2-4.
- Per-window fresh-init deployability evidence, including baseline-pinned pre-slice-2 and compatibility-pinned pre-slice-5 states.
- Operator guidance for existing embedded-to-fetched Gastown cities in air-gapped or cache-miss environments, including the pre-populated ordinary remote cache workaround and failure mode for unreachable pins.

**Required changes:**
- Rewrite the activation commit table so the activation `PublicGastownPackVersion` switch and Maintenance removal are a single candidate gate, or add a current-loader compatibility gate proving the activation pin can safely coexist with bundled Maintenance before the constant changes.
- Enforce `requires_gc` semver constraints in pack resolution and fail closed with downgrade guidance before unsupported binaries load activation-pack assets.
- Resolve the activation boundary explicitly: either prove `v1.2.1 + activation pin` loads safely or declare activation a one-way boundary with tested manual recovery steps and release-note/doctor guidance.
- Make compatibility-pin feasibility a slice-1 gate: prove colliding assets can be omitted without behavior gaps, or convert the plan to the documented paired cross-repo activation/removal boundary.
- Require `public-gastown-pins.yaml` to carry mandatory `v1.2.1 + activation` result evidence and non-empty manual recovery fields whenever rollback is one-way or manual.
- Normalize all ledger prose and validator behavior to the three-row model, and reject missing `current_baseline`, `compatibility`, or `activation` evidence where applicable.
- Replace copy-on-absence state migration with an epoch, timestamp, monotonic version, or equivalent reconciliation mechanism that detects legacy Maintenance state advanced during downgrade and resynchronizes it on re-upgrade.
- Add a generated rollback artifact requirement covering exact files touched, old-binary readability results, and operator commands for doctor-mutated manifests, locks, and runtime state.
- Define durable public ref requirements and enforce them in CI so all pinned commits remain reachable from default branch or protected immutable tags.
- Add a no-gap completeness gate for slices 2-4, symmetric with the zero-duplicate gate.
- Add an operator-communications and release-gate row for the offline fresh-init capability reduction, including the pre-populated ordinary remote cache workaround.
- Add per-window fresh-init deployability assertions so no intermediate slice can ship a city that initializes successfully but is not deployable.
