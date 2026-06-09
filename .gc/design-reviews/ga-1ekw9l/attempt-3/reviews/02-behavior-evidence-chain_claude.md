# Oleg Marchetti - Claude

**Verdict:** approve-with-risks

> Lane: Gastown behavior inventory completeness, execution-level witnesses,
> cross-repo packcompat, old→new traceability, source-deletion gate. Reviewed
> against the current `implementation-plan.md` (528 lines,
> `updated_at: 2026-06-09T01:20:00Z`) — §"External Public Gastown Prerequisite"
> (104–128), §"Behavior Evidence Contract" (130–166), §"Data And State"
> (359–364), and the Rollout source-deletion gates (453–518). I verified the
> repo anchors rather than trust the prose.
>
> Output written to this iteration's reviews dir (`attempt-3/`) beside the
> iteration-3 Codex sibling. My routed bead `ga-0yz16y` is a dynamic expansion
> instance with `gc.attempt=1` while its iteration is 3 (logical bead
> `ga-tk219g` is `attempt=3`); the literal `attempt-${gc.attempt}` path would
> overwrite the unrelated iteration-1 review, so I used the live iteration-3 dir
> and report it via `design_review.output_path`.

**Schema conformance:** Conforms. Eight required top-level sections present once
each in order; front matter complete with `phase: implementation-plan`, no
`design_file`; `Open Questions` is `None` with prose clarifying that the
behavior-evidence artifacts are external prerequisites rather than open design
questions. No appended attempt/review prose in the artifact.

**Top strengths:**
- Completeness is generation-driven, not hand-curated — directly answering red
  flag #1. A generator walks old behavior-bearing sources (Core, Maintenance,
  Gastown examples, hook overlays, formulas, orders, prompts, skills, shell
  scripts, doctor strings, route metadata, notification templates, runtime-state
  helpers, tests, helper refs) and "CI fails if a moved, split, generalized,
  deleted, or helper-dependent asset lacks a row, if a row lacks old and new
  witnesses, or if a semantic delta lacks an approved record" (155–160). Every
  row also requires an immutable public Gastown commit and the consuming
  `PublicGastownPackVersion` value (150–151), which is real old→new traceability.
- `test/packcompat` verifies the **pinned public checkout**, not copied assets —
  directly answering red flag #3: it "installs public Gastown through the
  ordinary remote-pack path or a validated ordinary remote cache, composes moved
  formulas and orders, resolves moved scripts using pack-relative paths, verifies
  hook overlays and configured agents, and checks one assertion per manifest row"
  (162–166), and runs first in compatibility-pin mode, then no-Maintenance
  production-loader mode in the activation slice (416–418).
- The two-meaning pin is a precise deletion gate: the activation pin "is consumed
  only in the same candidate branch that removes Maintenance from required host
  packs and proves no-Maintenance production loading" (124–128), and Slice 1
  blocks all Gas City source deletion / Maintenance removal until the
  `gascity-packs` prerequisite exists at immutable commits (107–109, 455–459).
  The existing old witnesses are substantial — `examples/gastown/gastown_test.go`
  (3,322 lines) and `examples/gastown/maintenance_scripts_test.go` (8,607 lines)
  — so the inventory has real coverage to map from.

**Critical risks:**

- **[Major] "Source assertion" as an allowed old witness undercuts the
  execution-level mandate (red flag #2).** The row schema lets the old witness be
  a "source assertion, test, fixture, golden output, or command transcript"
  (148). A *source assertion* is static — it proves the old tree *contained* a
  string, not that the trigger *fired*. For any behavior that had no old test, a
  row can satisfy "lacks old and new witnesses → CI fail" with a source assertion
  and a correspondingly static new check, so the behavior is "covered" on paper
  without an executing witness on either side. The mandate is execution-level
  witnesses; the plan should require that every row whose old witness is a source
  assertion be upgraded to an executing `test/packcompat` assertion (trigger
  fires through normal resolution) or carry an explicit, approved
  no-execution-possible record.

- **[Major] The generator's trigger-detection method is unspecified, so
  asset-level row presence does not guarantee trigger-level completeness (red
  flag #1, depth).** The CI completeness gate keys on *assets* ("if a … asset
  lacks a row", 158–159), but the row fields ("trigger, requester, detector,
  route metadata, mail/nudge target, prompt fragment, script branch,
  runtime-state path, or named-session behavior", 142–144) are singular. A single
  formula/order/script can declare multiple requesters/detectors/notifications;
  the plan never says a multi-trigger asset must emit one row per trigger, nor how
  the generator *identifies* each trigger kind inside a behavior-bearing source
  (heuristic? parser? grep?). Without a defined detection method and a
  one-row-per-trigger rule, an asset can be "covered" while individual triggers
  inside it go un-enumerated — exactly the lane-Q1 gap.

- **[Major] Single mutable `PublicGastownPackVersion` plus an unresolved
  version-skew window leaves an "exact public pin" hole.** The pin is one
  constant (`internal/config/public_packs.go:11`,
  `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`) already consumed by fresh init
  (`config_test.go:865`). The design treats it as two pins (compatibility vs
  activation) updated across Slices 2 and 5, so rows proven against the
  compatibility commit must be re-proven against the activation commit. The plan
  re-runs packcompat in activation mode (416–418), which helps, but requirements
  Open Question 3 (supported skew window between the Gas City release that
  requires public Gastown and the commit fresh init pins) is still unresolved.
  Until that window is pinned, a row's recorded "consuming `PublicGastownPackVersion`
  value" can drift from what fresh-init cities actually resolve.

- **[Minor] Cross-repo deletion-gate atomicity is a fallback, not a defined
  mechanism.** The plan correctly forbids deletion before the landing commit
  exists, but the Gas-City-deletes ↔ gascity-packs-lands coupling is handled by
  "stop and convert the slice into a paired cross-repo activation boundary" if
  compatibility can't be satisfied (486–488). The actual CI mechanism that, at
  deletion time, proves the named immutable commit is landed *and* contains the
  trace row's behavior (beyond packcompat fetching it) is not pinned — which
  matters for lane-Q3 enforcement across two repositories.

**Missing evidence:**
- The generator's trigger-detection method and the one-row-per-trigger rule for
  multi-trigger assets; how "requester/detector/notification" relationships are
  recognized inside formulas, orders, scripts, and Go helpers.
- The list of behaviors whose only old witness is a source assertion (no old
  test), and the plan for giving each an executing new witness.
- `test/packcompat` does not exist yet (confirmed absent; the other `test/`
  suites — acceptance, agents, docsync, integration, packlint, tmuxtest — do).
  The execution-witness machinery is fully greenfield and an external
  prerequisite (AC7); no `behavior-preservation.yaml`, `public-gastown-pins.yaml`,
  or `plans/core-gastown-pack-migration/artifacts/` exists yet.
- The owner/command that generates the manifest and ledger (requirements Open
  Questions 1–2 are unresolved), and the exact supported version-skew window
  (requirements Open Question 3).
- The CI check that verifies a row's immutable `gascity-packs` commit is landed
  and contains the behavior at deletion time, independent of packcompat fetch.

**Required changes:**
- Require execution-level witnesses: every manifest row must carry a
  `test/packcompat` assertion that fires the trigger through normal resolution
  from the exact public pin; permit a static source-assertion witness only with
  an explicit approved "no executable witness possible" record naming owner and
  reason.
- Define the generator's trigger-detection method and mandate one row per
  trigger (not per asset) for multi-trigger formulas/orders/scripts; state how
  each trigger kind (requester, detector, notification, mail, nudge, script
  branch) is recognized.
- Pin the version-skew window (requirements OQ3) and state whether a transition
  needs both compatibility and activation pins representable at once, or a
  single-pin cutover with a defined diagnostic for cities on the old pin.
- Specify the cross-repo deletion gate as a concrete CI check: a Gas City
  deletion PR fails unless the row's named `gascity-packs` commit is immutable,
  reachable, and proven by packcompat against that exact pin.

**Questions:**
- For a behavior that had no old test, what executing witness proves equivalence
  — or is a static source assertion accepted? The mandate implies the former.
- Does one behavior-bearing asset with N triggers produce one row or N rows, and
  what enumerates the N triggers?
- Across Slices 2→5, is `PublicGastownPackVersion` a single cutover value, and
  what happens to cities that fresh-init against the compatibility pin before
  activation lands?
