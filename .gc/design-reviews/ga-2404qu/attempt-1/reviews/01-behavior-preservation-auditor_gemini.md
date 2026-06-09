# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 18 / Attempt 18)

**Verdict:** approve

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

---

### Executive Summary

The Iteration 18 design represents the definitive, highly matured, and structurally complete specification for the Core and Gastown Pack Split. It successfully translates all behavioral safety requirements into concrete, enforceable contracts, executable scanner rules, and rigid CI/CD validation gates.

This iteration incorporates the highly detailed resolution contracts established in Attempt 17, which systematically address and resolve all critical risks of silent capability loss, boundary leakage, and untraced asset drift. 

Specifically, this review evaluates the design's effectiveness against the four major behavior-preservation risks identified in previous iterations:
1. **The AST Parser vs. Dynamic Shell Eval Discovery Gap:** Resolved by the mandate under §2597–2600 requiring independent baseline transcript comparison from VCS history, ensuring that dynamically evaluated commands or notification targets are never pruned or omitted from the behavior manifest.
2. **Multi-Layer Configuration Override Ambiguities:** Mitigated through explicit collision rules in §2471–2472. Conflicting duplicate global names across host Core and public Gastown evaluate cleanly as a fatal error unless an explicit alias is declared in the behavior manifest.
3. **Stale Synthetic Cache Risks:** Resolved by §2538–2542, ordering legacy synthetic cache handling before git validation, so retired synthetic public Gastown/Maintenance entries are cleanly detected and ignored rather than causing partial-materialization version skew.
4. **Provider-Pack Terminology and Post-Cleanup Regression:** Prevented via continuous linting under §2638–2640, where the wording linter remains a permanent CI gate spanning Markdown, MDX, JSON, TXT, TOML, Go strings, and TypeScript files.

Furthermore, the global blockers from the Attempt 17 overall synthesis are elegantly and thoroughly addressed from the perspective of behavior preservation. The creation of the importable system-pack boundary (`internal/systempacks`) in §2425 and the CST-preserving doctor mutation coordinator in §2556 ensure that both runtime loading and offline repairs remain fully deterministic, protecting existing operator state.

---

### Top Strengths

- **VCS-Independent Manifest Oracle (§2597–2600):** By comparing generated behavior manifest rows with old-tree VCS moves, deletions, and old baseline transcripts, Gas City prevents the static discovery tool from silently missing dynamically constructed script paths or dynamic evaluations (e.g. `eval "gc session nudge $target"`).
- **Comprehensive Provider Continuity Witness Matrix (§2670–2685):** Enforces explicit, execution-level witnesses for non-trivial provider file rewrites. This includes configured and empty recipient handling, `pool = "dog"` symbolic overrides, `DOG_DONE` nudges, and health JSON consumer schemas.
- **Hermetic Packcompat Offline Execution (§2009–2010):** Prevents external network dependencies in restricted or air-gapped CI pipelines by guaranteeing that `test/packcompat` can run using local-fixture or pre-populated remote-cache entries.
- **The Behavior Witness Floor (§1393–1401):** Mandates that any behavior possessing old execution-level coverage must receive execution-level replacement coverage. Count-only or path-presence validations are strictly rejected.

---

### Prior Risks & Gaps Resolution

#### 1. Discovery Blind Spots (AST vs. Dynamic Eval)
- **The Prior Risk:** Discovery tools using regex or AST parsing could miss dynamically generated notification paths or dynamic target lookups in shell scripts, allowing behaviors to be deleted from Core without corresponding replacement in public Gastown.
- **The Resolution:** Under §2597–2600, the manifest verification is decoupled from current-tree generation. It enforces cross-referencing against independent old-tree file moves, deletions, and baseline execution transcripts, failing the build if an executed path is unrepresented in the manifest.

#### 2. Multi-Layer Configuration Precedence
- **The Prior Risk:** Multi-layer config overrides (such as user-provided `city.toml`, system packs, and host includes) could collide on symbolic target bindings, silently misrouting alerts or escalations.
- **The Resolution:** Section §2471–2472 establishes that duplicate active global names across Core and public Gastown are fatal unless a manifest row explicitly defines a safe alias, preventing silent resolution drift.

#### 3. Stale Synthetic Cache Pollution
- **The Prior Risk:** A stale, partially updated synthetic cache could load outdated prompt templates alongside new Core logic, resulting in silent runtime errors.
- **The Resolution:** Section §2538–2542 orders legacy synthetic cache handling before ordinary git validation, ensuring retired synthetic entries are detected, classified, and ignored cleanly.

#### 4. Post-Cleanup Provider Terminology Drift
- **The Prior Risk:** Future updates to independent provider packs could accidentally re-introduce legacy Maintenance references or violate role neutrality after the split.
- **The Resolution:** The docs and wording linter under §2638–2640 is codified as a permanent CI gate. It continuously scans all operator-facing code, scripts, CLI help, schemas, and provider pack assets to ensure role names (`dog`, `mayor`, etc.) remain confined post-split.

---

### Evaluation of the Three Key Questions

#### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding: Yes.**
- **Evidence & Justification:**
  - The behavior manifest is anchored as an executable gate under the `Source-Derived Behavior Manifest` contract (§88–121) and the `Generated Artifact Contracts And Independent Completeness` table (§2580–2596).
  - This is reinforced by §2144–2175 (`Role-Surface Ownership And Host-Core Patch Model`), which maps all legacy surfaces (including Go literals, prompt fragments, and provider-pack `dog` routes) to distinct final owners.
  - No source file can be deleted or moved until every old behavior has a designated manifest row (§2479–2481).

#### 2. Which concrete gascity-packs/gastown tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding: Yes.**
- **Evidence & Justification:**
  - The design requires `go test ./test/packcompat` with sharded targets running the `TestPinnedPublicGastownBehavior` suite (§2627, §3108).
  - This is validated under the `Behavior Evidence Witness Floor` (§1393–1401), which forbids count-only or file-existence proofs, requiring full execution-level old and final witnesses.
  - Section §2654–2685 (`Witness Rows For Behavior And Provider Continuity`) specifies exact witness requirements for branch pruning, Polecat formulas, shutdown-dance requesters/detectors, and `DOG_DONE` nudges under simulation.

#### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding: Yes.**
- **Evidence & Justification:**
  - Reviewers can trace all asset relocations through `plans/core-gastown-pack-migration/slice-gates.generated.yaml` (§2591), which serves as the single source of truth for rollout gates.
  - The sequential 7-slice rollout plan (§2627–2685) and the 6-phase green intermediate commit ladder (§2261–2271) guarantee a continuous chain of custody. Every intermediate commit is verified test-green before destructive file removals occur.

---

### Required Changes for Finalization

**None.** All critical behavior preservation risks, terminal boundary edge cases, and provider-pack rewrite exceptions have been fully and robustly codified into the design as concrete, non-negotiable contracts and test gates. The design is fully approved from the Behavior Preservation perspective.
