# Felix Moreau - Claude

**Verdict:** approve-with-risks

Lane: documentation consistency, operator terminology, tutorial integrity,
Maintenance-word disambiguation. All findings are verified against the current
design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md`, line
refs below) and the design's own `rig_root` docs tree
(`/data/projects/gascity-fresh-main-20260604-VLKm8c/docs/`). I confirmed file
existence and exact strings before reporting.

**Top strengths:**
- **The enforcing docs lint is extension-agnostic, whole-tree, and
  allowlist-based** (lines 1459-1460, 2021, 2350, 2638, 1918 all name `.mdx`,
  `.json`, `.txt` explicitly; 3257 makes any retained old wording require an
  explicit allowlist row). This is the right structural answer to lane Q1/Q2:
  unnamed stale files are still forced to conform, so the worklist need not
  enumerate every page. The gate would catch stale content rather than ship it.
- **One generated wording matrix is the shared source for doctor strings, CLI
  help, docs, and tutorials**, with goldens, freshness tests, and docs-move-in-
  the-same-slice-as-behavior + non-release marking (lines 1168-1175, 3258-3264).
  A single contract for doctor and doc wording is the correct fix for my top red
  flag (doctor wording diverging from doc wording).
- **Maintenance-word disambiguation is handled head-on** (lane Q3): the
  `[maintenance.dolt]` store-maintenance carve-out (3234-3235, 3253-3254), the
  out-of-scope rule for supervisor/store maintenance (requirements §Out Of
  Scope), and the case-aware matrix (capital "Maintenance" = retired pack vs
  lower-case English) cleanly separate retired-pack from store/Dolt maintenance.

**Critical risks:**
- **[Major] The recommended docs-inventory command contradicts the design's own
  extension-agnostic lint contract.** Line 3222's command globs only
  `-g '*.md' -g '*.toml' -g '*.go' -g '*.sh'`, yet line 3220 frames it as how to
  "Build the docs/operator inventory from the current tree before editing docs."
  It is provably blind to the highest-risk operator file:
  `docs/troubleshooting/gc-start-walkthrough.mdx` carries
  `includes = ["packs/gastown"]` (line 263) and routes a missing-agent error to
  `.gc/system/packs/gastown/agents/mayor/agent.toml` (lines 134-135) — a retired
  local-Gastown path *and* a Gastown role name in operator troubleshooting
  output. It is also blind to `docs/docs.json` (the nav index the design itself
  requires system-packs.md be registered in, line 3245) and to generated
  `docs/schema/*.json` (a release gate, lines 3260-3261). Because the enforcing
  lint *is* extension-agnostic, this is not a silent operator-facing ship — but
  the planning inventory and the enforcing gate disagree, so `.mdx`/`.json` work
  surfaces as an unplanned late CI failure instead of being scoped. Red flag:
  inventory step blind to half-migrated local-Gastown/Maintenance paths; lane Q2.
- **[Major] The canonical operator wording teaches "maintenance-agent" while the
  config key operators must type is `maintenance_worker`.** Line 3234 ("Core
  maintenance-agent behavior") and the tutorial-01 instruction at line 3251
  ("describe the `dog` pool as Core's configurable maintenance agent") use
  "agent," but the binding is `[gc.bindings.maintenance_worker]`, override key
  `maintenance_worker`, diagnostic `core.maintenance_worker` (lines 1799-1817),
  and the rest of the design standardizes on "maintenance worker" (942, 2360,
  2150, 2461). An operator who follows the canonical wording or tutorial looks
  for a `maintenance_agent` key that does not exist. This noun lands in the
  matrix and the tutorial golden, so it bakes in at generation time. Red flag:
  doctor/doc wording divergence; lane Q1.
- **[Major] No single canonical public-Gastown source string is pinned; the docs
  teach a non-install `/tree/main/` URL.** The design's canonical wording uses
  bare `github.com/gastownhall/gascity-packs/gastown` (line 3233), its install
  source uses `https://github.com/gastownhall/gascity-packs.git//gastown` (line
  1253), but `docs/guides/shareable-packs.md` uses
  `https://github.com/gastownhall/gascity-packs/tree/main/gastown` in every
  `[imports.gastown]` example (rig_root lines 113, 154, 172, 192) — a GitHub
  *browser* URL, not the pinned install form — while its `gc pack registry add`
  uses `…gascity-packs.git` (line 128). The design names shareable-packs.md only
  to "remove 'core and maintenance stay implicit' guidance" (3246-3247) and
  never reconciles the source-string form. An operator copying the doc gets a
  non-canonical (likely non-resolving) source. Lane Q1.
- **[Minor] The seeded doctor string violates the design's own case-aware rule.**
  Line 3086 changes `import_state_doctor_check.go` messaging to "**m**aintenance
  is retired; Core supplies generic maintenance and Gastown supplies
  Gastown-specific behavior" — lower-casing the *pack* name, while the canonical
  matrix entry (line 3231) is "**M**aintenance is retired as a standalone pack."
  Because this seed becomes the doctor golden, the matrix would reject its own
  seeded string. Lock the capitalized form and assert it in the doctor golden.
- **[Minor] Residual "creates" vs "baseline" wording for the canonical
  reference.** `docs/reference/system-packs.md` **exists** in the design's
  rig_root (verified, 2114 bytes), and the design now mostly treats it as a
  baseline to update and nav-link (lines 3242-3245, 1761 "canonical baseline").
  But the attempt-7 contract at line 933 still says the slice "creates and
  nav-registers" it. Reconcile to "designate and update the existing page;
  docs-nav test enforces it," so the docs-nav test author is not misled. (Much
  improved over earlier attempts — this is now only a residual wording nit.)

**Missing evidence:**
- Whether the wording lint distinguishes a forbidden *token* from a wrong
  *operator instruction*. The `packs/gastown` token in
  `gc-start-walkthrough.mdx` will trip a token-keyed lint, but the
  troubleshooting *logic* — telling operators to add `includes =
  ["packs/gastown"]` to resolve a missing agent — is now wrong end-to-end and
  needs a prose rewrite to the public-import model, not a token swap. The design
  does not say the lint enforces instruction correctness, only wording.
- The exact path of the public-Gastown operator-facing **companion reference**.
  The behavior *manifest* path is named (`gastown/docs/behavior-manifest.generated.yaml`,
  line 328), and line 2635 says "Public Gastown owns a companion reference that
  uses the [matrix]," but no concrete companion-doc path or cross-repo validator
  is named, so it can drift from Gas City's matrix.
- Whether `docs.json` nav-registration of system-packs.md is actually verified by
  a test, given the inventory command (3222) cannot read `docs.json`.

**Required changes:**
- Make the line-3222 inventory command extension-agnostic (add at least
  `-g '*.mdx' -g '*.json' -g '*.txt'`, and TypeScript globs for generated
  dashboard text), **or** replace it with the generated wording-linter
  invocation and label the `rg` form as a non-authoritative sanity check. The
  planning inventory must see every format the release gate enforces.
- Choose one operator noun for the Core maintenance worker — "maintenance
  worker," matching the `maintenance_worker` key — and apply it in the canonical
  wording list (3234), the tutorial-01 instruction (3251), and the matrix;
  reserve `dog` for the compatibility default-name examples only. If the matrix
  intends to accept "maintenance agent" as a variant, say so explicitly and
  still make the tutorial use the key noun.
- Pin exactly one canonical public-Gastown source string in the wording matrix
  and reconcile `.git//gastown`, bare-path, and `/tree/main/gastown` across
  design prose (1253, 3233), shareable-packs.md imports, and `gc pack registry
  add`. Add shareable-packs.md's source-string fix to the named worklist, not
  only the "implicit" removal.
- Lock the capitalized doctor phrase ("Maintenance is retired; …") at line 3086
  and assert it in the doctor golden.
- Reconcile line 933 ("creates") with the "baseline"/update framing, since
  system-packs.md already exists in the rig_root.

**Questions:**
- Does the manifest-derived lint enforce *instruction* correctness, or only
  *wording*? If only wording, what catches that `gc-start-walkthrough.mdx` tells
  operators to add a now-retired `includes = ["packs/gastown"]`?
- What is the single canonical public-Gastown source string the matrix enforces,
  and does it accept or reject the `/tree/main/gastown` browser form currently in
  shareable-packs.md?
- Should the canonical operator noun be "maintenance worker" (matching the config
  key), and is "maintenance agent" an accepted matrix variant or a forbidden one?
- What concrete path is the public-Gastown companion reference, and does Gas City
  CI validate it from the exact pinned public checkout?
