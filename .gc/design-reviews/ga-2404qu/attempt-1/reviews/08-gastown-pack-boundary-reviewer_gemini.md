# Avery McAllister — DeepSeek V4 Flash Perspective Independent Review (Iteration 20 / Attempt 20)

**Verdict:** BLOCK

**Lane:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, and the public pack contract.

---

## Executive Summary

As **Avery McAllister**, the **08-gastown-pack-boundary-reviewer**, I have conducted an independent DeepSeek V4 Flash-styled evaluation of the Iteration 20 / Attempt 20 design (`design.md` and `requirements.md` active draft updated at `2026-06-09T08:40:42Z`). I have grounded every finding in the physical workspace tree of `/data/projects/gascity` and audited the design's structural safety from a pack boundary, dependency isolation, and rollout contract perspective.

While the current design snapshot incorporates exceptional architectural advancements—including the default-deny loader inventory (§2408), the ownership-aware scanner (§2483), and the symbolic `dog` maintenance worker bindings (§2456)—it contains critical internal contradictions, unmitigated version-skew hazards, and self-defeating diagnostics that threaten the integrity of the Gas City / Gastown boundary. 

Most notably, we have identified a **direct architectural collision** between the report-only doctor diagnostic boundary and the materializing loader invoked by the new `Core Presence Doctor` check (§3061). This materializing loader performs automatic on-disk repairs and quarantines outside of the `MutationCoordinator` transaction, introducing a severe risk of destroying stale-but-preserved legacy directories (`.gc/system/packs/maintenance`), thereby completely undermining our rollback safety invariants (§2975). Furthermore, critical gaps regarding relative `PromptTemplate` path resolution in patching layers remain unresolved.

Consequently, my verdict for Attempt 20 is a firm **BLOCK** until these boundary, safety, and compatibility contradictions are formally reconciled.

---

## Top Strengths of Attempt 20

1. **Symbolic Maintenance-Worker Binding (§1787, §2468):** Utilizing symbolic bindings (`[gc.bindings.maintenance_worker]`) rather than hardcoded Go role constants successfully isolates de-roling logic in configuration. It represents an outstanding application of Zero Framework Cognition (ZFC) and ensures full SDK self-sufficiency even if the default agent is renamed or omitted.
2. **Defensive Source-Classifier Guardian (§2177, §2577):** The introduction of the `internal/packsource` classifier to actively identify, ignore, and reject retired directories (e.g., `.gc/system/packs/maintenance`, legacy `../maintenance` imports) provides robust defense-in-depth, preventing stale or generated code from being transitively loaded.
3. **Rigid File-Set and Digest Verification (§2580):** Requiring generated, checked-in, and freshness-tested YAML artifacts (such as `role-surface.generated.yaml` and `loader-inventory.generated.yaml`) before implementation beads move behavior ensures the entire migration is auditable, deterministic, and independent of VCS history.
4. **Preservation of Ignored Legacy Paths (§2975):** Banning aggressive startup deletion of legacy system-pack directories is an excellent rollback safety invariant. It preserves operator-modified files and ensures that downgraded binaries immediately recover access to their required assets.

---

## Critical Risks & Architectural Gaps (The Blockers)

### 1. The `PromptTemplate` Path Resolution Void (Blocker)
* **The Code Evidence:** `internal/config/config.go` defines `AgentOverride` with `PromptTemplate *string`. Attempt 17 introduces symbolic patching of the maintenance worker via `[[patches.agent]]` (§2470).
* **The Gap:** When the public Gastown pack patches the host-Core `maintenance_worker` to use its custom prompt template, it must declare a relative path (e.g., `prompt_template = "assets/prompts/deacon.template.md"`). However, the design fails to specify the base directory against which this relative path is resolved. If the loader resolves it relative to the city's root, it will fail because the template only exists inside the public Gastown pack's directory. If the loader resolves it relative to the patching pack's root, this behavior must be explicitly specified and implemented in the config loader layer.
* **Why it matters:** Without explicit resolution rules, cross-pack patching of agent prompts is unimplementable and will result in fatal "file not found" errors or silent omissions during prompt compilation.
* **Recommendation:** Update §2470 to mandate that the config layer resolves relative `prompt_template` paths inside imported patch layers against the root directory of the patching pack (`gascity-packs/gastown`), rather than the city's root directory or the host-Core pack's root directory.

### 2. Version-Skew TOML Parse Crash on older binaries with Activation Pin (Blocker)
* **The Code Evidence:** The Release Compatibility matrix (§3465–3472) says:
  `| old binary | new pack | Public pack remains compatible ... no reliance on new Gas City-only loader behavior |`
* **The Gap:** This row holds true for the compatibility pin, but is fundamentally false for the activation pin. Under Attempt 11/17, the activation pin requires active public Gastown assets to utilize `target_binding = "core.maintenance_worker"` configuration keys (§2470). Older binaries do not implement this configuration field. Because the older config loader parses TOML with strict undecoded-key warnings (`fatalUndecodedWarnings`), loading the activation-pin pack on a legacy binary will cause an immediate crash on startup.
* **Why it matters:** Version-skew safety is compromised. Operators running older binaries against a city that attempts to load the public Gastown activation pin will suffer immediate, unmitigated startup crashes.
* **Recommendation:** Split the "old binary | new pack" compatibility matrix row into two: one for the compatibility pin (which remains fully compatible), and one for the activation pin (which triggers a version-skew diagnostic or warning, with manual recovery instructions).

### 3. Core Presence Doctor Check invoking Materializing Loader (Blocker)
* **The Code Evidence:** §3061 specifies:
  `- Load resolved config through internal/systempacks.LoadRuntimeCity and verify typed RequiredSystemPackParticipation includes Core.`
* **The Gap:** The `Read-Only Doctor Diagnostic Boundary` (§2106–2122) explicitly mandates that plain `gc doctor` is strictly report-only and "must not call materializing runtime loaders, ... repair helpers, ... or quarantine/prune writers." It permits only non-refresh API calls. Calling `LoadRuntimeCity` directly violates this invariant.
* **Why it matters:** `LoadRuntimeCity` is the materializing/writing entrypoint that performs on-disk repairs and quarantines. If a report-only `gc doctor` run does this, it will perform mutations on disk *outside* of the `MutationCoordinator` transaction. This can result in unexpected file pruning/quarantines (including the accidental deletion or quarantine of `packs/maintenance` or `packs/gastown` before the operator is ready to transition), destroying rollback capability and violating the "preservation of ignored legacy paths" invariant.
* **Recommendation:** Update §3061 to require that the Core Presence Doctor check loads the config strictly via `LoadRuntimeCityNoRefresh` and validates required file sets via `ValidateRequiredFileSetsNoRefresh`, leaving all materialization, repair, and quarantine side effects exclusively to `gc doctor --fix` under the control of the `MutationCoordinator`.

### 4. Ghost Asset Remnants in the Cross-Pack Table (Major)
* **The Code Evidence:** The Cross-Pack Ownership Decisions table (§3275) includes the `Gastown Codex overlay` as a pending `review` decision.
* **The Gap:** The in-tree Gastown Codex overlay has already been deleted in the live codebase (as verified in §19–21 of sibling reviews). Retaining this row in the design document creates an implementation hazard, as developers will waste cycles trying to audit and move a ghost asset.
* **Recommendation:** Completely delete the `Gastown Codex overlay` row from the Cross-Pack Ownership Decisions table (§3275).

### 5. Undefined Overlay-File Collision Semantics (Major)
* **The Gap:** While the design details strict collision detection for duplicate formulas and orders, it does not define merge or precedence rules for file-set overlays (e.g. `overlay/per-provider/codex/.codex/hooks.json`). If both Core and an imported pack attempt to supply the same destination overlay file, the behavior is undefined. Will the imported pack overwrite the Core overlay, or will the loader raise a fatal duplicate error?
* **Recommendation:** Establish a clear precedence rule for overlays: imported pack overlays merge with or override host Core overlays of the same destination path, and any fatal collision checks are confined to exact duplicate content.

---

## Technical Evaluation of Invariant Questions

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

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 20 Design | Status |
| :--- | :--- | :--- |
| **Gastown still imports or documents Maintenance** | **Pass.** Scrubbed completely. `internal/packsource` rejects explicit, transitive, and generated Maintenance imports (§2198). | **Pass** |
| **Shared Core asset retains Gastown-specific conditional branches** | **Pass.** AST scanner's de-roling scope covers Go, prompts, TOML defaults, and script texts (§2483–2503). | **Pass** |
| **Gas City removes in-tree Gastown tests before equivalent public-pack coverage exists** | **Pass.** Handled via `TestPinnedPublicGastownBehavior` and the strict "Behavior Witness Floor" (§2922–2930). | **Pass** |

---

## Actionable Requirements & Proposed Adjustments

1. **Mandate Read-Only Doctor Check:** Update §3061 to mandate that `cmd/gc/core_pack_doctor_check.go` loads the config strictly via `LoadRuntimeCityNoRefresh` and validates via `ValidateRequiredFileSetsNoRefresh`, ensuring that the report-only `gc doctor` command never triggers materialization or prune side effects.
2. **Specify Cross-Pack Template Resolution:** Update §2470 to mandate that the config layer resolves relative `prompt_template` paths inside imported patch layers against the root directory of the importing/patching pack.
3. **Correct the Version Skew Matrix:** Split the "old binary × new pack" row in §3465 to explicitly separate the `compatibility` pin from the `activation` pin, detailing the expected diagnostic failures on the activation pin for old binaries.
4. **Purge Ghost Codex Overlay:** Remove the stale `Gastown Codex overlay` row from the Cross-Pack Ownership Decisions table (§3275).
5. **Define Overlay Collision Rules:** Add an explicit invariant to the loader design: "Two active packs shipping the same destination overlay file must merge their configurations if JSON/JSONL, or let the imported pack overlay take precedence over Core, without raising a fatal loading error."

---

## Questions for the Implementation Team

1. Will the generated release notes for the activation pin explicitly warn operators about the one-way binary compatibility boundary?
2. If the AST scanner encounters role literals within comments of public Gastown assets, is it configured to only record them as inventory rather than triggering role-neutrality failures?
3. In `test/packcompat`, is the offline remote-cache pre-populated before running CI to ensure a completely deterministic test gate?
