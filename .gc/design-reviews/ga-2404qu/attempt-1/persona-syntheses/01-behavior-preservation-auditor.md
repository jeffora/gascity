# Nadia Volkov

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash (present as the `_gemini.md` artifact)

**Consensus findings:**
- [Major] The design is directionally strong for behavior preservation. All reviewers approved with risks and credited the generated behavior manifest, role-surface inventory, old-tree evidence, public pin ledger, test-migration map, witness floor, and packcompat plan as the right blocking structure before destructive migration work.
- [Major] Behavior proof must be execution-level where old behavior had execution-level evidence. File lists, path-presence checks, source-count checks, docs-only preservation, or two independent synthetic renders are not enough for active requester, detector, notification, nudge, formula, prompt, script, or runtime-state behavior.
- [Major] The `mol-shutdown-dance` warrant seam is the most concrete behavior-loss risk. Gastown-owned filers emit `{target, reason, requester, warrant_id}` while Core extracts and acts on those fields, so the field contract must be frozen and proven by a public-Gastown-to-Core round trip, not by independent fixtures on each side.
- [Major] Required recipient handling needs explicit preservation. Shutdown-dance `requester` and other migrated notification/escalation targets must not degrade into optional empty recipients, slash targets, or silent skips unless that change is recorded as an approved semantic delta.
- [Major] Semantic-delta approval is underspecified. The design blocks unrecorded behavior loss, but it must define who can approve a behavior delta and make that approval independently auditable rather than implementer-settable.
- [Minor] Some high-risk and final-design wording is still weaker than the strongest witness contract. In particular, `mol-shutdown-dance` examples should not appear preservable through documentation unless classified docs-only, and final manifest fields should say old/final witness rather than old/new test where legacy behavior lacks a direct test.
- [Minor] `test/packcompat` hermeticity and local pin override states need explicit checks so cache-backed or offline-compatible tests do not accidentally depend on network fetches or bypass manifest validation under local redirects.

**Disagreements:**
- Claude requires concrete design additions for the warrant round trip, requester classification, semantic-delta approval authority, and `orphan-sweep` / cross-rig-deps pilot rows. Codex focuses more on tightening stale wording and making generated row ids operational in implementation beads. Assessment: both are valid; the operational bead criteria should carry the stricter concrete contracts where they protect live behavior.
- DeepSeek emphasizes dynamic recipient preflight and packcompat hermeticity, while Claude and Codex emphasize cross-pack warrant and witness wording risks. Assessment: these are complementary rather than conflicting; all affect whether old observable behavior can disappear without a failing gate.
- Codex accepts that generated artifacts are not present yet because the design blocks destructive work until they exist. Claude stresses that the guarantee is only contractual until slice 1 lands. Assessment: approve the design only with slice 1 as a hard evidence gate.

**Missing evidence:**
- The actual `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, `test-migration.generated.yaml`, `public-gastown-pins.yaml`, schema validator output, freshness checks, and packcompat result artifacts.
- A baseline transcript from the old tree reconciled against manifest rows, including executed commands, mail and nudge routes, formula/order defaults, prompt renders, script branches, runtime-state mutations, and approved removals.
- A packcompat witness where pinned public Gastown files a warrant and Core `mol-shutdown-dance` extracts and executes `{target, reason, requester, warrant_id}` unchanged.
- Required-vs-optional recipient classifications for shutdown-dance `requester` and other migrated notification routes, including negative tests for empty, malformed, or slash targets.
- Execution witnesses or approved semantic deltas for `orphan-sweep` and cross-rig-deps city-qualified and ephemeral assignee handling.
- Proof that `test/packcompat` runs in a genuinely offline/cache-backed mode and that manifest validation remains enforced when local redirect or overridden public-pack pins are active.

**Required changes:**
- Make the first implementation slice generate and gate on the behavior manifest, role-surface inventory, old-tree baseline evidence, public pin ledger, test-migration map, witness rows, schema validation, freshness checks, and packcompat artifacts before any destructive source move, role-neutral rewrite, public pin update, registry/cache retirement, docs claim, or Maintenance removal.
- Freeze the cross-pack warrant field contract `{target, reason, requester, warrant_id}` and require a packcompat round trip from pinned public Gastown warrant filer output to Core `mol-shutdown-dance` extraction and execution.
- Classify shutdown-dance `requester` and other behavior-bearing recipient fields as required or optional in the manifest. Required recipients must fail preflight with an explicit diagnostic when empty, malformed, unbound, or `/`.
- Define independent semantic-delta approval authority and make approved behavior loss auditable by someone other than the implementing bead.
- Promote `orphan-sweep` and cross-rig-deps to named high-risk preservation moves with first-pass witness rows, or require approved semantic-delta records that cite the original regression and operator impact.
- Tighten stale final-design wording so active detector/requester behavior cannot be preserved only in documentation, and replace "old Gas City test" / "new gascity-packs test" language with "old witness" / "final witness" while retaining test mappings where tests exist.
- Add a packcompat hermeticity assertion and ensure pin-coherence plus behavior-manifest validation still run when local redirect config overrides public pack pins.
