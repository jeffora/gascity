# Lena Driscoll - Codex

**Verdict:** block

**Top strengths:**
- The design separates Phase 8 first-party requires-only conversion from Phase 9 parser alias removal, with distinct rollback units for restoring dual declarations versus restoring alias parser support.
- Built-in and example graph formulas are kept dual-declared through the compatibility window, and requires-only conversion is gated on minimum-floor, active-root, external-support, provenance, and stale-guidance evidence.
- Caller migration is split into producer sub-phases with zero-write fixtures, which gives the plan a workable path for keeping `main` green while dispatch, orders, API, convergence, molecule execution, and fanout move at different speeds.

**Critical risks:**
- [Blocker] Phase 3's executable gate does not include the old-reader pack-load and exact-floor proof that the design says is mandatory before first-party pack floors are written. Lines 4403-4418 require full `pack.toml` old-reader fixtures, exact old-reader binaries, exact floor values, and explicit accept/ignore/reject behavior before writing first-party `[pack] requires_gc` floors. But the operable phase table at lines 5220-5223 says Phase 3 is proven by `gc formula validate --all-packs --legacy-contract-report --json`, whose documented shape is a formula alias inventory, not a pack-floor old-reader compatibility gate. That creates a path to ship built-in pack floor edits with a passing legacy-contract report while a supported older binary might reject or misread the pack before formula selection.
- [Major] Phase 4 rollback is directionally correct but not operationally pinned. Lines 5235-5241 require each migrated producer to have a switch that disables only the new producer surface before durable writes, but the phase table only requires "package tests plus row-specific zero-write fixtures" and the sub-phase rows describe rollback in prose. For `gc sling`, API sling, orders, convergence, molecule execution, and fanout, the design needs named controls and command-level evidence that a regression can be reverted without also reverting docs, first-party dual declarations, or unrelated producer migrations.

**Missing evidence:**
- A Phase 3 gate artifact that combines first-party dual/floor edits, old-reader pack-load corpus results, exact minimum-floor comparator output, and release-floor JSON validation.
- A statement that the Phase 3 required command consumes the old-reader pack-load fixtures by digest, or a separate required command that does.
- Per-producer rollback-control artifacts for Phase 4b through 4g: control name, owner, disabled-mode behavior, command/config surface, and a zero-write assertion for the disabled path.

**Required changes:**
- Replace the Phase 3 required local command with an explicit floor-declaration gate, for example `gc formula validate --all-packs --first-party-floor-declaration-gate --json`, or require `--legacy-contract-report` plus a named old-reader/probe corpus and min-floor validation command. The artifact must fail closed on placeholder floors, missing old-reader binaries, stale corpus output, or any supported reader that cannot safely handle the declared floor.
- Update the release checklist language so Phase 3 cannot be approved from formula alias inventory alone. It must cite the old-reader pack-load fixture artifact required by lines 4403-4418.
- Add concrete rollback controls for each Phase 4 producer sub-phase. Each row should name how the surface is disabled, what happens when it is disabled, which test proves no durable writes occur, and which rollback artifact is attached to the phase evidence.

**Questions:**
- Is `--legacy-contract-report` intended to invoke the old-reader pack-load corpus implicitly? If yes, rename or document the command contract so Phase 3's gate is not narrower than the prerequisite it is meant to prove.
- What is the concrete disabled mode for a regressed `gc sling` or API sling migration after typed diagnostic docs and dual-declared first-party formulas have already shipped?
