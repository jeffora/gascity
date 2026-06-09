# Leah Okafor - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan replaces direct `Check.Fix(ctx)` mutation with `FixIntent` plus a single mutation coordinator before public-pin, import-rewrite, or runtime-state fixes are enabled (`design-before.md:246`-`design-before.md:251`).
- The coordinator is explicitly lock-first for mutation safety: it acquires a crash-released city advisory lock, repeats digest and provenance validation after the lock, and refuses automatic fix when a same-city controller is live (`design-before.md:253`-`design-before.md:257`).
- Multi-file mutation is not hand-waved as per-file rename safety. The plan requires durable recovery state before publish, staged edits, target digest re-reads before each temp-file rename, a single commit point, and deterministic rerun/rollback after crash (`design-before.md:259`-`design-before.md:264`).

**Critical risks:**
- [Major] Byte-preserving TOML behavior is present in the Testing section, but not yet part of the coordinator's design contract. The plan should state in Proposed Implementation that every TOML edit backend must preserve comments, unknown tables, unknown fields, array order, formatting, and unrelated lock entries, and must refuse mutation if it cannot prove that preservation before staging.
- [Major] The advisory-lock story does not say how lock loss is detected between staging and per-file publish steps. A crash-released lock protects against dead owners, but the coordinator should also carry an owner token and verify it before each publish/rename, especially for network filesystems or competing new binaries.
- [Minor] Recovery state is well described, but the commit-point semantics need one concrete invariant. For example: before commit, rerun either restores all preflight bytes or completes the staged publish; after commit, rerun only validates/finishes post-commit markers. Without that invariant, implementers may encode a publish-order log that still leaves ambiguous half-migrated cities.

**Missing evidence:**
- Coordinator API shape for `FixIntent`, staged mutation plans, and preservation proofs.
- The TOML edit strategy that proves byte-preserving scoped edits instead of rewrite-and-hope behavior.
- Lock owner-token or lease validation behavior across stage, validate, compare-before-rename, publish, and recovery.
- A sample recovery-state record showing pre-commit versus post-commit replay decisions.

**Required changes:**
- Move the byte-preserving TOML/refuse-on-uncertainty rule from test expectations into the mutation coordinator design contract.
- Add lock ownership validation before every publish step and during recovery, not only at initial acquisition.
- Define the single commit-point invariant and recovery replay rules concretely enough that failure-injection tests can assert them.
- State that import rewrites for custom local forks or operator-edited paths require provenance proof of system-generated state; otherwise the fixer must return manual guidance.

**Questions:**
- What TOML editing library or structured edit strategy will the coordinator use to prove no unrelated bytes were rewritten?
- Does the advisory lock have an owner token or generation that can be checked before every staged rename?
- Which file marks the commit point for cross-file fixes, and is it written before or after the final content rename?
