# Claire Dubois - Claude

**Verdict:** approve-with-risks

I reviewed only my lane: operator upgrade DX, the terminology matrix, docs and
schema generated artifacts, doctor messages, and tutorial integrity. I also
schema-checked the artifact. The wording machinery is genuinely strong — a
token-level terminology matrix with allowed/denied contexts, false-positive
rules, and golden fixtures, plus same-slice "regenerate then lint" release gates.
That directly answers my "executable enough to distinguish retired wording" and
"behavior ships with stale docs" concerns. The risks are coverage and
discoverability gaps that the lint machinery does not by itself close: the
plan's hardcoded docs/tutorial update list is provably narrower than the live
tree, the canonical reference is created but not required to be navigable, and
there is no single operator-readable upgrade narrative — which is the literal
center of this lane (Q3). All are bounded additions on top of sound machinery.

**Top strengths:**
- **The wording matrix is executable and discriminating, which is the anti-
  wholesale-suppression defense.** Each row carries "token, class,
  allowed-contexts, denied-contexts, owner, examples, false-positive rule, and
  golden fixture" and explicitly distinguishes retired standalone
  Maintenance-pack wording from "valid lowercase maintenance, Dolt/store-
  maintenance terminology, Core-maintenance-worker bindings, public Gastown
  docs, historical examples, and operator recovery text" (L512–516). That is
  exactly the Q2 discrimination my red flag ("noisy false positives get
  suppressed wholesale") requires.
- **Docs are release-gated to the behavior they describe, not deferred.** "Update
  [system-packs, shareable-packs, troubleshooting, tutorials], CLI help, doctor
  strings, generated references, examples, and script comments in the same slice
  as the behavior change they describe" (L521–525), and OpenAPI/dashboard/docs-
  schema/CLI-help/tutorial/doctor goldens "regenerate before wording lint runs"
  (L517–519). The Rollout slices treat docs as release gates (AC12 → first
  operator-facing slice). This closes my "non-release docs debt" red flag at the
  process level.
- **The plan grounds the doctor-string change in a real stale string and a real
  missing page.** It correctly targets the "Core and Maintenance are supplied
  implicitly" doctor wording (L52–55) — which is live today at
  `cmd/gc/import_state_doctor_check.go:194` (`"maintenance/core is supplied
  implicitly"`) — and correctly anticipates that `docs/reference/system-packs.md`
  does not yet exist, mandating its creation (L508–510, confirmed absent).

**Critical risks:**

- **[Major] The hardcoded docs/tutorial update set is provably incomplete against
  the live tree, and the omissions are the migration docs an upgrader reads
  first.** L521–525 names `system-packs.md`, `shareable-packs.md`,
  `troubleshooting.md`, and tutorials `01/05/07`. But live operator-facing docs
  that reference retired pack paths today and are *not* named include
  `docs/getting-started/coming-from-gastown.md`,
  `docs/guides/migrating-to-pack-vnext.md`, `docs/reference/cli.md`, and
  `docs/tutorials/index.md`. The migration/onboarding pages are exactly where a
  Gastown operator upgrading would land, so omitting them is my "doctor/
  tutorials/docs still point operators to Maintenance or in-tree Gastown" red
  flag. AC12's `docs-authority-audit.yaml` is the correct complete-by-construction
  inventory; the plan must state that the *audit drives the update set* and that
  the prose list is illustrative, not the binding scope — otherwise the plan
  reads as a narrower set than reality.

- **[Major] There is no named single operator upgrade narrative — the literal
  lane-Q3 deliverable.** The plan has the pieces (AC10 upgrade-matrix tests, AC11
  diagnostics schema, AC15 pin ledger, the release compatibility matrix at
  L810–818), but those are test/spec artifacts inside the plan, not an
  operator-readable runbook that threads "missing orders / stale packs →
  `gc doctor` diagnosis → public-pin verification → recovery" into one followable
  story. `docs/guides/migrating-to-pack-vnext.md` exists and is the natural home,
  yet it is unnamed. Decomposition cannot produce "one narrative" from a list of
  disjoint diagnostic artifacts.

- **[Major] The canonical reference is created but not required to be
  discoverable.** The plan mandates creating `docs/reference/system-packs.md`
  (L508–510) but says nothing about nav/index registration or a docs-nav test
  (the design lineage required nav-registration). No `mkdocs.yml` or docs nav
  manifest surfaced in the tree, so "navigable" is itself under-defined here. A
  canonical page an operator cannot find does not satisfy the "one narrative"
  mandate.

- **[Minor] Tutorial integrity is lexical, not behavioral, as specified.** The
  plan wording-lints and regenerates "tutorial transcripts" (L518, L674) but
  never says tutorials are *executed* — commands run and output captured against
  post-migration behavior. A tutorial can pass the wording matrix while showing
  commands or output that no longer match (e.g., `gc init --template gastown`
  surface). Specify execution-captured transcripts so tutorial integrity is
  behavioral.

- **[Minor] The wording matrix needs paired positive *and* negative controls per
  discriminated token to keep the lint from being neutered.** L513 lists
  "examples, false-positive rule, and golden fixture" (singular). To hold the
  "maintenance" lowercase vs. "Maintenance pack" vs. "store/Dolt maintenance"
  line under pressure, each ambiguous token needs a passing positive control and
  a failing negative control, plus owner+expiry on any wording allowlist row —
  otherwise a broadened context rule silently launders a real stale reference.

**Missing evidence:**

- No statement that `docs-authority-audit.yaml` (AC12) is the authoritative
  source of the docs-update set, nor that the doctor *message templates* and the
  AC11 condition-code registry render through the same terminology matrix (the
  plan lints doctor-output goldens but does not bind condition-code messages to
  the matrix).
- No nav/index target or docs-nav test for the new `system-packs.md`.
- No named operator upgrade-narrative doc, and no end-to-end upgrade transcript
  golden.
- The inline `<!-- REVIEW: added per … -->` markers (e.g. L496) are review
  provenance the schema says belongs in the workflow artifact directory; front
  matter, section set, and order otherwise conform.

**Required changes:**

1. Make the docs-update scope derive from `docs-authority-audit.yaml`, not a
   hardcoded prose list, and explicitly include the migration/onboarding docs
   (`coming-from-gastown.md`, `migrating-to-pack-vnext.md`, `reference/cli.md`,
   `tutorials/index.md`).
2. Name the canonical operator upgrade-narrative doc (e.g. `migrating-to-pack-
   vnext.md`) that threads missing-Core/stale-pack → doctor → pin verification →
   recovery, and add an executed end-to-end transcript golden for it.
3. Require `docs/reference/system-packs.md` to be nav/index-registered with a
   docs-nav test, not merely created.
4. Specify tutorial transcripts are execution-captured against post-migration
   behavior, and bind doctor message templates + condition-code rendering to the
   terminology matrix.
5. Require paired positive/negative golden controls per discriminated wording
   token, with owner+expiry on wording allowlist rows.

**Questions:**

- Is the docs-update file set the AC12 audit (complete) or the L521–525 prose
  list (incomplete)? They disagree on at least four live files.
- Which doc is the canonical operator upgrade narrative, and is it gated by an
  executed transcript?
- How is `system-packs.md` made discoverable in this repo's docs structure (no
  nav manifest is evident), and what test enforces it?
