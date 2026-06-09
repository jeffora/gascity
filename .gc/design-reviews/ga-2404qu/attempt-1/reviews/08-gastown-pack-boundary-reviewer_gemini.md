# Avery McAllister — DeepSeek V4 Flash (Gastown Pack Boundary Reviewer, Iteration 18)

**Verdict:** approve-with-risks

**Lane:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, and the public pack contract.

---

## Executive Summary

As **Avery McAllister**, the **08-gastown-pack-boundary-reviewer**, I have conducted an independent DeepSeek V4 Flash-styled evaluation of the Iteration 18 / Attempt 18 design (`design.md` and `requirements.md` active draft updated at `2026-06-09T08:40:42Z`). I have grounded every claim against the physical workspace tree and verified the architectural soundness of the split.

In Iteration 17, this lane was blocked by severe, load-bearing contradictions regarding the template fragment collision strategy, the lack of an explicit cross-pack re-skinning mechanism for the Core-owned `dog` agent, and the potential for config-loading pathways to transitively leak retired Maintenance assets on disk. 

The Iteration 18 design has systematically resolved these blocking concerns with high-quality architectural invariants:
1. **Dog Prompt Separation (§2470–2471):** Duplicate prompt fragments across Core and public Gastown are declared fatal. Global names are strictly pack-scoped, and there is no implicit cross-pack lookup. Gastown overrides default host-Core maintenance behavior via symbolic configuration patches (`[[patches.agent]] name = "dog"`) rather than fragile same-name fragment shadowing.
2. **Explicit Retired-Source Rejection (§2177–2208):** The `internal/packsource` classifier is established as the sole, rigid guardian. It runs pre-resolution and actively ignores or rejects retired directories (`.gc/system/packs/maintenance`, `../maintenance`, etc.), completely closing the transitive config loading leak.
3. **Role Neutrality Scanner Coverage (§2854–2859):** The AST role-token scanner has been expanded to cover TOML default values (such as `mol-shutdown-dance.toml` defaults) and active script texts, preventing accidental leakage of Gastown roles into Core.

While the design is now structurally sound and ready for implementation, minor risks regarding version-skew TOML parse behaviors and developer dirty-checkout stale-path warnings remain. These are detailed below and should be documented as accepted boundary conditions before merge.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Does `gascity-packs/gastown` own all Gastown roles, formulas, orders, scripts, prompts, overlays, doctor checks, and commands after the split?
**Yes, completely.**
* All Gastown-specific role behaviors (Mayor, Deacon, Witness, Refinery, Polecat, Boot, Crew), Polecat formulas (`mol-polecat-*`), branch pruning, custom overlay hooks, and custom diagnostics are fully owned by `gascity-packs/gastown`.
* Prompt fragments are separated (§2471) with zero implicit cross-pack lookup, preventing any boundary bleed.
* Rather than importing Core, public Gastown layers its custom behavior onto the default host-Core maintenance agent by declaring symbolic agent patches using `target_binding = "core.maintenance_worker"` (§2171–2172).
* Active command/script texts that previously lived in Maintenance are either generalized in Core or migrated to Gastown.

### Q2: Do Gastown pack comments and imports describe explicit Gastown behavior plus host-required Core without implying a standalone Maintenance layer?
**Yes.**
* With the complete retirement of the Maintenance pack and the introduction of the `internal/packsource` classifier (§2185), any explicit or transitive import of `../maintenance` or `.gc/system/packs/maintenance` in Gastown pack configurations is strictly rejected (§2198).
* Comments and documentation in `gascity-packs/gastown/pack.toml` have been fully scrubbed of legacy references to Maintenance. Public Gastown's configuration describes explicit, pure Gastown orchestration behavior layering cleanly over host-provided Core without assuming or referencing an intermediate, standalone Maintenance pack.

### Q3: Which test proves Core plus public Gastown, with no Maintenance pack, preserves Polecat, branch pruning, detector, requester, and review workflow behavior?
**The Dual-Mode Integration Gate in `test/packcompat`.**
* This is proven via `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` (§2922–2930, §3438–3441).
* During the **compatibility pin stage**, the test runs with the new loader but with Maintenance remaining active on disk, asserting green status.
* During the **activation pin stage**, the test runs in a pure, no-Maintenance production-loader environment using the exact public Gastown pack commit pinned in `public-gastown-pins.yaml` (§2216). This test explicitly acts as the witness proving that Polecat, branch pruning, shutdown-dance detector/requester, and review workflow behaviors are preserved without relying on legacy in-tree Maintenance code.

---

## Top Strengths of the Design

1. **The `internal/packsource` Classifier Invariant (§2185–2189):** Establishing a single, unified source-classifier before any config loading, lock validation, or doctor check is exceptional systems engineering. It prevents stale, left-behind, or generated legacy directories from being accidentally globbed or imported.
2. **Strict Template Fragment Ownership (§2471):** Banning same-name template shadowing and implicit cross-pack lookups forces a clean break between Core and Gastown prompt assets, ensuring the boundary remains hermetic.
3. **Multi-Phase Pin Ledger (§2520–2529):** Tracking the public pins across three explicit phases (`current_baseline`, `compatibility`, `activation`) in `public-gastown-pins.yaml` ensures that rollout steps can be verified incrementally.
4. **Expanded AST Scanner Scope (§2483–2503):** Extending role-token checking to TOML default values and active script text ensures role neutrality across non-Go assets (e.g., formula variables in `mol-shutdown-dance`).

---

## Critical Risks & Architectural Gaps

### 1. [Major] Version-Skew TOML Parse Crash on older binaries with Activation Pin
* **The Risk:** Under §2520–2529, the activation pin adopts `target_binding = "core.maintenance_worker"` in the public Gastown pack's active assets.
* **The Hazard:** Older binaries (such as `v1.2.1`) do not understand the `target_binding` configuration key. Because the older config loader parses TOML with strict undecoded-key warnings (`fatalUndecodedWarnings`), loading the activation-pin pack on a legacy binary will cause an immediate crash on startup.
* **The Verdict:** This is an accepted but sharp one-way boundary. Legacy binaries cannot load the activation pin and must remain on the compatibility pin (§2233). This boundary must be explicitly documented in the release notes.

### 2. [Minor] Stale Source Tree Cleanup of examples/ path
* **The Risk:** Removing `examples/gastown/packs/maintenance` only in the activation candidate commit (§3207).
* **The Hazard:** In dirty developer checkouts, cached build directories, or local worktrees, stale files under `examples/gastown/packs/maintenance` may remain untracked. If a local run glob-discovers these stale files, the `internal/packsource` classifier will trigger retired-source diagnostic warnings.
* **Recommendation:** Mandate a `make clean-stale-packs` helper or integration in the pre-commit hook that actively purges untracked legacy directories under `examples/` to avoid spurious local developer warnings.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 18 Design | Status |
| :--- | :--- | :--- |
| **Gastown still imports or documents Maintenance** | **Pass.** Scrubbed completely. `internal/packsource` rejects explicit, transitive, and generated Maintenance imports (§2198). | **Pass** |
| **Shared Core asset retains Gastown-specific conditional branches** | **Pass.** AST scanner's de-roling scope covers Go, prompts, TOML defaults, and script texts (§2483–2503). | **Pass** |
| **Gas City removes in-tree Gastown tests before equivalent public-pack coverage exists** | **Pass.** Handled via `TestPinnedPublicGastownBehavior` and the strict "Behavior Witness Floor" (§2922–2930). | **Pass** |

---

## Actionable Requirements & Proposed Adjustments

1. **Document Activation-Pin Version Skew:** Explicitly state in `public-gastown-pins.yaml` and the release notes that the transition from the compatibility pin to the activation pin represents a one-way version-skew boundary. Legacy binaries must remain on the compatibility pin (§2233).
2. **Purge Stale Examples Locally:** Introduce a quick clean-up script or add a task in the developer instructions to run `git clean -fdx examples/gastown/packs` after checking out the activation branch to prevent stale local files from triggering retired-source diagnostics.

---

## Questions for the Implementation Team

1. Will the generated release notes for the activation pin explicitly warn operators about the one-way binary compatibility boundary?
2. Does the AST scanner check for role literals in comments within the public Gastown pack to ensure no lingering legacy documentation is shipped?
3. In `test/packcompat`, is the offline remote-cache pre-populated before running CI to ensure a completely deterministic test gate?
