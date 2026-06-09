# Oleg Marchetti - Claude

**Verdict:** approve-with-risks

Reviewed `plans/core-gastown-pack-migration/implementation-plan.md` (identical to
the `attempt-8/design-before.md` snapshot) against `requirements.md` (AC6, AC7,
AC13, AC14, AC15) and `implementation-plan.schema.md`, grounded in the live
behavior surface under `examples/gastown/packs/{maintenance,gastown}` and the
existing old-witness tests `examples/gastown/maintenance_scripts_test.go` (126
test funcs) and `examples/gastown/gastown_test.go` (70 test funcs). Lane:
Gastown behavior-inventory completeness, execution-level witnesses, cross-repo
packcompat, old-to-new traceability, source-deletion gate.

**Top strengths:**
- Git historical-baseline comparison (lines 205-206) defeats the
  delete-then-workspace-scan-misses-it failure mode: the generator compares
  against an immutable old commit, so deleted/moved legacy behavior cannot
  silently escape a manifest row.
- Activation-pin packcompat mode fails if any assertion resolves from
  `examples/gastown`, `.gc/system/packs/{gastown,maintenance}`, or a synthetic
  alias (lines 214-219), and installs only via the ordinary remote-pack
  path / commit+subpath+digest-keyed cache (lines 211, 329-334). This directly
  closes the "consumes copied assets instead of verifying the pinned public
  checkout" red flag.
- Source deletion is genuinely gated: a row whose new witness lives in public
  Gastown cannot unblock deletion until public commit + pack digest +
  behavior-manifest digest + packcompat transcript are cited in the pin ledger
  (lines 195-197), with Slice 1a/1b/1c sequenced ahead of Slice 5b/7.

**Critical risks:**

- **[Major] Manifest granularity is internally contradictory (per-asset vs.
  per-trigger), so the inventory cannot be shown to enumerate every
  requester/detector/notification/nudge.** The CI failure condition is
  asset-level — "CI fails if a moved, split, generalized, deleted, or
  helper-dependent **asset** lacks a row" (lines 203-204) — but Testing requires
  "one-row-per-**trigger** extraction" (line 637), and the row schema collapses
  `{trigger, requester, detector, route, mail/nudge target, prompt fragment,
  script branch, runtime-state path, named-session}` into one field joined by
  "or" (line 178: one value per row). Concrete existential proof from the source
  the plan itself splits: `examples/gastown/packs/maintenance/formulas/
  mol-shutdown-dance.toml` is one asset that lines 451-453 split ("generic
  stuck-session due process in Core, but Gastown detector/requester examples
  move to Gastown"). That single file carries 3 distinct nudge sites (60s/120s/
  240s with different messages), ≥4 distinct mail targets (`PARDON`, `DOG_DONE`,
  `EXECUTE_FAILED`, kill-escalation), 4 detector→target edges (Deacon health-scan
  →Witness, →Refinery, →Dog; Witness check-polecat-health→Polecat), a
  `gc session kill` side effect, and a warrant-requester relationship. A per-asset
  row collapses all of these into one row + one packcompat assertion and cannot
  represent a split that runs *along trigger lines*. Under per-asset rows, lane
  question 1 is unsatisfiable.

- **[Major] The old witness may be downgraded to a static "source assertion,"
  so "the same trigger fires" is asserted rather than proven — even though
  execution-level old witnesses already exist.** The witness-kind enum (lines
  180-183) permits the *old* witness to be a "source assertion" (static), while
  only the *new* witness must be "executable" (line 194); the hinge between them
  is a human "semantic-equivalence assertion" (lines 189-190). But the repo
  already holds 196 execution-level old tests that `exec.Command` the real
  scripts, inject fake `gc`/`bd`/`jq`, and assert exact rendered output
  (`"orphan-sweep: reset 1 orphaned beads"`, `"all required binaries available
  (jq)"`), command logs (`"bd update ga-orphan --status=open --assignee="`), and
  negative conditions (secret not leaked; live-race not reset). Allowing a row to
  cite a mere static source assertion when such an executable witness exists
  silently drops coverage and reduces equivalence to a claim. The plan also never
  requires the old witness to be *executed* at `--old-baseline` to confirm it
  passed there — it is cited, not replayed — so the old half of the chain is
  unproven.

- **[Major] Two parallel inventories are never required to reconcile.** The
  behavior-preservation manifest (my section, Slice 1b) selects old/new
  witnesses; AC13 `coverage-transfer.yaml` (Slice 5b) separately maps the 196
  old tests/assertions to replacements. The plan never requires bidirectional
  cross-checking between them, so a trigger can carry a weak source-assertion
  witness in the manifest while being "accounted for" in coverage-transfer (or
  vice versa), and neither gate catches a trigger that falls between the two.

**Missing evidence:**
- Extraction semantics for prose-encoded relationships. The detector→target
  table ("Who files warrants") and the nudge/mail sites live as Markdown inside
  TOML `description` strings, not structured metadata. A freshness-checked
  generator cannot reliably enumerate prose relationships without either lifting
  them into declarable metadata or hand-curating — and hand-curation is the
  named red flag. The plan does not say which path it takes.
- No measurable completeness floor. AC7 names an "authoritative supported-Gastown
  before-state denominator," but the Behavior Evidence Contract section never
  binds the manifest's old-witness column to a frozen digest of the 196-test /
  assertion baseline, so "every behavior has a row" has nothing to count against.
- Undefined packcompat semantics for non-executable rows. Line 212 asserts "one
  assertion per manifest row," but `retired-approved` and `docs-only` rows (lines
  183-184, 192) have no fireable behavior to assert.

**Required changes:**
- Reconcile granularity: declare the manifest row key as **per-trigger**, define
  the extraction contract mapping each source construct to a row (one row per
  nudge site, per mail/notification target, per detector→target edge, per script
  branch, per prompt fragment, per route), make the CI failure condition
  per-trigger, and require a split asset to emit one row per moved trigger.
- Strengthen the old witness: when an executable old witness (test/golden/
  transcript) exists for a trigger, require the manifest to cite *that* witness
  (not a static source assertion) and require the new witness to demonstrate
  equivalence to its assertion (same rendered nudge text / same script stdout /
  same mail subject), not merely "formula composes." Restrict "source assertion"
  and "approved removal record" old-witness kinds to triggers with no
  pre-existing executable witness, recorded as an explicit gap. Require old
  witnesses to be executed/validated at `--old-baseline`.
- Add a reconciliation gate that fails if any frozen AC13 old-test assertion is
  not bound to a behavior-evidence manifest row, and if any `moved`/`split`/
  `generalized` manifest row lacks a corresponding coverage-transfer mapping.
- Define packcompat behavior for `retired-approved`/`docs-only` rows (assert
  *absence*/non-resolution rather than firing), so "one assertion per row" is
  well-defined for every evidence class.

**Questions:**
- Is the manifest row keyed per-asset or per-trigger? The plan states both
  (lines 203-204 vs. 637).
- When an executable old test already witnesses a trigger, is the manifest
  required to cite it, or may it cite a weaker static source assertion?
- Is there a single gate binding the AC7 behavior-evidence manifest to the AC13
  coverage-transfer denominator bidirectionally?
- Does the actual source-deletion slice (Slice 7) require the live public-network
  AC14 proof that the cited public commit is real and landed, or may deletion
  merge on a digest-verified *cached* fixture with live validation deferred to a
  later release gate? The two-repo release-order rule mitigates this, but the
  merge-vs-release timing for the deletion slice is not pinned down.
