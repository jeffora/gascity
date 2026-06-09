# Mira Acharya

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] The requirements have a strong evidence-first testing posture. Claude, Codex, and DeepSeek all identify AC17's acceptance-proof matrix, AC13's coverage-transfer validator, AC8's role-neutrality absence scans, and AC14/AC16 release/offline gates as the right backbone for this migration.
- [Info] The role-neutrality and legacy-test-transfer contracts are directionally strong. AC8 includes denied-token matching, generated/materialized/rendered output coverage, allowlists, positive and negative controls, and rot guards; AC13 requires frozen baseline coverage, active execution evidence, and fail-closed handling for skipped, empty, or no-op witnesses.
- [Major] Active execution and no-op defenses must apply to every invariant gate, not just AC13. Claude and DeepSeek both warn that absence scans, public-pack validation, executor-binding tests, and behavior-preservation witnesses can falsely pass if they run zero assertions, skip fixtures, scan empty inputs, or only prove file presence.
- [Major] Fresh generation/materialization must be part of the gates. Claude flags a stale-snapshot false pass: scanning checked-in generated files or stale `.gc/system/packs/*` output is not enough if regeneration can reintroduce a role token or retired Maintenance import.
- [Major] AC2/AC9 need a minimum controller-only operation denominator. Codex flags that "normal SDK infrastructure operations" and "no-executor controller-only test" can be satisfied by shallow config loading unless the requirements name the Core-only operations that must work with Gastown roles absent and the maintenance executor renamed, omitted, undefined, or disabled.
- [Major] AC7 needs per-witness negative or mutation controls. Codex notes that a broad public-pack behavior check can pass while a moved prompt variable, route, notification target, script side effect, formula/order trigger, or recovery branch is broken.
- [Major] The AC2 dev/test escape hatch is a production-safety test risk. DeepSeek warns that the implementation plan does not define it; if it is loose, production can bypass required Core, and if it is absent, partial-config tests may fail.
- [Major] Bootstrap-only doctor/import-state tests must guard against eager loader dependencies. DeepSeek identifies the doctor boot-loop failure mode: if CLI startup resolves packs before recognizing bootstrap-only commands, operators cannot run the diagnostic that would repair missing Core.
- [Major] Multi-repository SHA-pin validation needs a local override path. DeepSeek flags a circular pin-lock: public Gastown cannot fully validate activation against final Gas City before the final `sha:` exists unless packcompat/pin-coherence tests support local cross-repo override inputs.
- [Minor] AC13's baseline should be digest-pinned, not merely "frozen" by convention. Claude recommends reusing the AC6 digest-frozen snapshot discipline so retired assertions cannot disappear from the denominator.
- [Minor] AC17 needs bidirectional AC-to-proof-matrix completeness. Claude notes that the matrix should fail both when an AC lacks evidence and when a matrix row references a nonexistent or changed AC.
- [Minor] AC17 should define a minimum acceptable validation command and gate shape so the proof matrix cannot be satisfied by an ad hoc weak command.
- [Minor] Tmux safety should be enforced by static scan. DeepSeek recommends denied-token coverage for raw `tmux kill-server` or unsafe `tmux kill-session` usage without isolated sockets.
- [Minor] Doctor repair needs crash-recovery tests that simulate unclean process exit mid-mutation and prove the next run recovers or rolls back deterministically.

**Disagreements:**
- Claude treats fresh-generation scanning and generalized no-op defenses as required requirements changes. Codex focuses on controller-only denominators and per-witness negative controls. DeepSeek adds CI mechanics such as local pin overrides, bootstrap isolation, and crash recovery. My assessment: these are complementary, not conflicting; the lane verdict remains approve-with-risks because the proof strategy is strong but still needs these guardrails to avoid false positives.
- Codex is comfortable with AC17 leaving some command shape to later design, while Claude wants AC-to-matrix drift guards and DeepSeek shows a detailed target matrix. My assessment: the requirements should at least require a stable validation command and bidirectional completeness, without needing to freeze every file path now.
- DeepSeek suggests specific mechanisms for the dev/test escape hatch and tmux scanning. My assessment: exact mechanisms can vary, but tests must prove production cannot activate the escape hatch and unsafe tmux cleanup cannot re-enter scripts/tests.

**Missing evidence:**
- A shared execution-evidence primitive or validator proving each mapped witness ran, was not skipped, was non-empty, and was not a no-op across AC7, AC8, AC9, AC13, AC14, and release gates.
- Fresh-generation/materialization controls for absence and behavior scans, including planted role-token and retired-Maintenance-import controls in regenerated output.
- A controller-only/no-executor operation matrix covering the concrete SDK surfaces that must work with only Core present and optional executors renamed, omitted, undefined, or disabled.
- Per-row or per-call-site negative/mutation controls for AC7 behavior-preservation witnesses.
- A digest-pinned AC13 retired-test/assertion/fixture/behavior-witness baseline complete against the live retired-test set.
- Bidirectional AC17 proof-matrix validation and a stable command/gate shape.
- A safe, test-only dev/test escape-hatch proof.
- Bootstrap-only CLI tests proving `gc doctor`, `gc import-state`, and `gc version` do not trigger normal pack resolution before command selection.
- Local override support for packcompat/pin-coherence validation during multi-repository CI and pre-final SHA testing.
- Static tmux cleanup scans and doctorfix crash-recovery simulations.

**Required changes:**
- Amend AC7 and AC8 verification so generated/materialized/rendered artifacts are produced fresh inside the gate and scanned live, with controls proving planted role tokens and retired Maintenance imports in regenerated output fail the gate.
- Generalize AC13's active-execution contract to AC7, AC8, AC9, AC14, and release gates: parse `go test -json` or equivalent execution logs, prove each witness ran, and fail on skipped, empty, or no-op witnesses.
- Amend AC2 and AC9 to list the minimum controller-owned operations that must pass with only Core present and with the maintenance executor renamed, omitted, undefined, or disabled.
- Amend AC7 to require negative or mutation controls for each behavior-row or call-site witness, especially prompt variables, routes, notification targets, script side effects, formula/order triggers, and recovery branches.
- Require AC13's baseline to be digest-pinned and complete against the live retired-test/assertion/fixture set.
- Add AC17 bidirectional completeness checks and require a stable proof-matrix validation command invoked by a named local/pre-commit/CI gate before decomposition or implementation approval.
- Define the dev/test Core-less escape hatch as strictly test-only and prove production `gc` commands cannot activate it.
- Add bootstrap-only CLI tests ensuring doctor/import-state/version are dispatched before normal pack resolution or required-system-pack validation can run.
- Add local override support to pin-coherence and packcompat validators so candidate Gas City and public Gastown changes can be tested together before the final public `sha:` pin exists.
- Add denied-token static scans for unsafe tmux cleanup and add crash-recovery tests for interrupted mutating doctor repairs.
