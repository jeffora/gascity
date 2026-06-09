# Oleg Marchetti

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] The requirements now treat Gastown behavior preservation as an acceptance proof, not a file-move promise. Claude, Codex, and DeepSeek all point to AC6/AC7 plus external Gastown validation as the intended proof chain: asset ledger, behavior-preservation manifest or equivalent harness, public-pack validation, pin/cache evidence, and release proof gates.
- [Info] The behavior surface is broad enough for this lane. The reviews agree that AC7 covers formulas, orders, scripts including `assets/scripts`, prompts and prompt fragments, template variables, notification targets, requester/detector relationships, identity side effects, success/warning/failure/escalation paths, and recovery flows.
- [Info] External Gastown validation with in-tree fallback disabled is correctly load-bearing. Codex and DeepSeek both identify AC14/AC15/AC16/AC17 as the guard against accidentally depending on the old in-tree copy or on live network access in normal CI.
- [Major] Cross-boundary trigger and notification continuity still needs an explicit composed witness. Claude's central risk is that endpoint proofs can pass while the real seam fails: a Core-owned detector or script invokes an override-resolution mechanism and a Gastown-owned notification/escalation target must actually fire. Examples include `spawn-storm-detect` -> `mayor/` escalation and `reaper` -> `mayor/`/`deacon/` notification.
- [Major] The AC5 script-dependency example must be corrected before it is used as behavior evidence. Claude credits the `port_resolve.sh`/`dolt-target.sh` closure as a useful worked example, while DeepSeek reports the actual sourcing direction is `dolt-target.sh` sourcing `port_resolve.sh`, not the reverse. That makes the example and downstream rehoming task risky until docs, design, and tests match the real dependency direction.
- [Major] Legacy sister-directory test coupling is a behavior-preservation release risk. DeepSeek reports `examples/dolt/port_resolve_test.go` still reads `examples/gastown/packs/maintenance/assets/scripts/dolt-target.sh`; if that path is deleted or isolated before the test is rehomed, CI can fail or preserve the wrong dependency shape.
- [Minor] The Core/non-Gastown baseline denominator should be explicitly anchored to the same AC6 frozen-before-deletion snapshot that now anchors the Gastown denominator. Claude sees this as an asymmetry; Codex considers the future proof artifacts sufficient if they preserve row-level granularity.
- [Minor] The implementation plan must keep behavior proof at row/call-site level. Codex's main caution is that a future manifest or harness must not collapse the inventory into file-level presence checks.
- [Minor] Public Gastown validation remains a release-coordination risk because it depends on pinned caches, public-pack availability, and cross-repository ordering, even though the requirements now name those gates.

**Disagreements:**
- Claude requires document-level tightening for AC7 cross-boundary witness coverage and the Core baseline denominator. Codex says no change is required before requirements approval as long as the proof artifacts remain mandatory. My assessment: approve with risks, but make the cross-boundary witness explicit before implementation slices so the seam cannot be reduced to endpoint checks.
- Claude treats the `port_resolve.sh` -> `dolt-target.sh` example as a closed worked side-effecting closure. DeepSeek says the code reality is the opposite direction. My assessment: treat this as unresolved evidence until the dependency direction is verified and the docs/design are corrected.
- DeepSeek treats the `test/packcompat` harness and behavior manifest as sufficient proof. Claude and Codex treat them as future acceptance artifacts rather than present evidence. My assessment: they are valid gates, but they remain missing evidence until generated and run.
- DeepSeek raises a dev/test escape-hatch hardening concern. That is credible for Core-loading safety, but it is outside this persona lane unless the escape hatch can affect behavior-preservation witnesses or production validation.

**Missing evidence:**
- The concrete AC6 asset ledger, AC7 behavior-preservation manifest or equivalent harness output, AC14 public Gastown validation output, AC15/AC16 pin/cache/version evidence, and AC17 acceptance-proof matrix.
- An end-to-end witness for a cross-boundary notification, escalation, requester/detector, trigger, or recovery path with in-tree fallback disabled.
- Proof that the override-resolution mechanism used by a Core caller can re-home stripped Gastown notifications and escalation targets and still fire the Gastown target.
- The explicit frozen source set for the Core/non-Gastown baseline.
- Verified source-of-truth documentation for the `dolt-target.sh` and `port_resolve.sh` dependency direction, plus updated test coverage that no longer depends on the retired Maintenance path.
- Evidence that an "equivalent harness" path, if used instead of a named manifest file, enforces the same stable behavior IDs, AC6 row equality, call-site witnesses, trigger checks, notification checks, and recovery-flow checks.
- Release-order evidence showing public Gastown publication, cache promotion, and Gas City validation happen in the required sequence.

**Required changes:**
- Make AC7 require an explicit cross-boundary row class for Core-owned detectors/scripts whose notification or escalation is stripped and re-homed to a Gastown-owned target. The witness must exercise the composed path: Core caller -> real override resolution -> Gastown target fires, with in-tree fallback disabled.
- Add at least one Example-Mapping row for a notification/escalation/trigger seam such as `spawn-storm-detect` -> `mayor/`, `reaper` -> `deacon/`, or `mol-shutdown-dance` warrant routing.
- Anchor the Core/non-Gastown baseline denominator to the AC6 frozen-before-deletion snapshot, matching the supported-Gastown denominator discipline.
- Correct the AC5/docs/design treatment of `dolt-target.sh` and `port_resolve.sh` so it matches the real sourcing direction, and ensure the implementation plan includes the actual rehoming/inlining work.
- Remove or rehome the `examples/dolt/port_resolve_test.go` dependency on the legacy Maintenance path before deleting or isolating that path.
- Require any manifest or equivalent harness to preserve row-level and call-site-level traceability, stable behavior IDs, trigger/notification/recovery witnesses, and fail-closed coverage rather than file-level presence checks.
