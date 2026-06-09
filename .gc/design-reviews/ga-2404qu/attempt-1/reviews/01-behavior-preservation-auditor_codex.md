# Nadia Volkov - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now treats Gastown behavior preservation as a blocking prerequisite, with replacement public Gastown commits pinned before Gas City stops owning behavior. The invariant at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:34` is the right policy: behavior must land in public Gastown at an immutable commit and be proven by CI before deletion or generalization.
- The behavior evidence contract is no longer file-list based. The pilot-row table at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2279` requires old and final execution witnesses for requester/detector semantics, nudges, notifications, Polecat handoffs, branch pruning, prompt fragments, hooks, runtime state, and authorship.
- The rollout order is defensible. Slice 1 must land the public Gastown preservation branch and generated manifest before either pin is consumed, and slice 5 couples activation pin adoption with Maintenance removal and no-Maintenance production-loader packcompat.

**Critical risks:**
- [Major] The normalized high-risk list still says `mol-shutdown-dance` detector examples can be preserved in "Deacon/Witness/Boot/Gastown documentation" at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2885`. That is weaker than the later execution-witness contract and could be misread as permitting docs-only preservation for requester/detector behavior that used to affect prompts or formulas. Tighten this bullet so docs-only preservation is allowed only for rows classified as docs-only; active requester/detector semantics need public Gastown prompt/formula/rendered-prompt witnesses or an approved semantic-delta row.
- [Minor] The final proposed-design section at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2866` lists "old Gas City test" and "new gascity-packs test" as required manifest fields, while the stronger generated-artifact sections allow old witnesses to be transcripts, source assertions, prompt renders, script side effects, or other execution evidence. If implemented literally, untested legacy behavior could be dropped as "no old test" instead of receiving an old-tree transcript. Update the final list to say "old witness" and "final witness", with tests required when an execution-level old witness exists.
- [Minor] The historical attempt sections still contain stale "iterate/global verdict block" prose, while the final design says no shared open questions remain. That is not a behavior-preservation flaw, but it is a handoff risk: implementation beads should rely on the generated artifacts and final rollout gates, not the historical verdict labels.

**Missing evidence:**
- The generated artifacts do not exist in this review artifact yet: `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, `test-migration.generated.yaml`, and `public-gastown-pins.yaml` are specified as first-slice deliverables rather than present inputs. That is acceptable for design approval only because the design blocks all destructive work until those artifacts are generated, validated, and cited by row id.
- No concrete public Gastown compatibility or activation commit is named yet. The design covers this through the pin ledger and immutable commit gates, but implementation must not substitute branch names, local worktrees, or synthetic cache bytes for the actual public pin.

**Required changes:**
- Revise the `mol-shutdown-dance` high-risk bullet so active detector/requester behavior cannot be "preserved" only in documentation unless the behavior row is explicitly docs-only or an approved semantic delta.
- Change the final manifest field wording from "old Gas City test" / "new gascity-packs test" to "old witness" / "final witness", while retaining test-function/subtest mappings where tests exist or are removed.
- Make implementation bead templates require affected behavior row ids before approving source moves, source deletion, Core generalization, public pin consumption, or Maintenance removal. This is already implied by the generated-artifact contract; make it operational so the first destructive bead cannot proceed from prose alone.

**Questions:**
- For `mol-shutdown-dance`, which old detector/requester examples are active prompt/formula behavior versus docs-only explanatory text? The answer determines whether the final owner needs a public Gastown prompt/formula witness or only a docs classification row.
- Will the first generator pass run against a clean old-tree baseline plus the exact pinned public Gastown checkout, or only against the current migration branch? The design says independent old-tree transcripts are required; the implementation should name the baseline commit in the generated artifact.
