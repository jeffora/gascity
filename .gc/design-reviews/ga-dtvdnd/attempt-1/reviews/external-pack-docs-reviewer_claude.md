# Simone Kaye - Claude

**Verdict:** approve-with-risks

Lane: external Gastown pack authority; registry behavior; source-tree
cleanliness; documentation consistency. Reviewed
`plans/core-gastown-pack-migration/requirements.md` (`updated_at:
2026-06-09T17:23:58Z`) against my mandate and, per AC1, the non-empty
`gc.mayor.requirements.v1` schema. None of my three red flags fire. This
revision also closed a prior gap: AC15 (line 121) now lists `docs/registry/repair
output` inside the must-agree coherence set, so operator-facing docs are no
longer outside the pin-coherence contract. The remaining items are Minor.
Grounding check (requirements phase, no implementation yet): retired roots
`examples/gastown/packs/{gastown,maintenance}` and legacy
`internal/bootstrap/packs/core` are present; the designated end-state
`internal/packs/core` does not exist yet — correctly framed as a migration
target, not a precondition.

**Top strengths:**
- **Public Gastown is made the proven, non-maskable authority — including
  workflow checks (lane Q1).** AC14 (line 120) requires the public checkout or
  pinned cache to prove "roles, prompts, commands, formulas, orders, overlays,
  scripts, checks, recovery flows, docs, generated references, registry/catalog
  metadata, and operator-facing text," forbids a local in-tree copy from masking
  a broken external pack, and records a two-repository release order so Gas City
  never ships a pin lacking the validated behavior manifest. AC7 (line 113) adds
  executable checks "with in-tree fallback disabled," and AC16 (line 122) makes
  offline resolution fail closed and never select in-tree/embedded/retired
  synthetic content. This forecloses "works locally but fails from public
  resolution."
- **Retired roots are stripped of authority as a closure, not a vibe (lane
  Q2).** AC4 (line 110) deletes or isolates the former in-tree Gastown/Maintenance
  roots as non-resolvable fixtures "excluded from runtime resolution, docs
  authority, init templates, and public-pack proof"; AC12 (line 118) forbids
  presenting `packs/maintenance`, `packs/gastown`, `examples/gastown/packs/*`, or
  `.gc/system/packs/*` as authoritative outside path-classified history recorded
  in `docs-authority-audit.yaml`; AC5 (line 111) `source-consumer-closure.yaml`
  classifies every in-tree-Gastown, mock-registry, docs, and test consumer with a
  retirement decision and proof command. No authoritative-looking dead Gastown
  source survives unclassified.
- **Cross-repo pin/docs/terminology coherence is pinned end to end (lane Q3).**
  AC15 (line 121) now requires the pin ledger, version-skew matrix, compatibility
  proof, lock/cache provenance, `docs/registry/repair output`, and fresh-init
  output to agree on source, subpath, immutable commit, pack digest, and
  behavior-manifest digest; AC12 enforces consistent
  Core-required/Gastown-external/Maintenance-retired language across docs,
  examples, CLI help, doctor, import-state, and registry/catalog/discovery
  surfaces, with a terminology matrix; deterministic CI uses pinned caches while a
  named live-network gate (AC14) validates fetchability.

**Critical risks:**
- **[Minor] Tutorials are not explicitly named in the docs-authority scope.**
  This project treats tutorials as the authoritative DX surface ("Tutorials win
  over architecture docs" is a settled design decision), so a tutorial or
  quickstart that still runs `gc init` against `examples/gastown/packs/gastown`,
  or that describes Maintenance as implicit, is the highest-impact form of my red
  flag "documentation points operators at retired Maintenance or examples paths."
  AC12 names "Documentation, examples, CLI help, ... generated references" and a
  `docs-authority-audit.yaml`, which very likely subsumes tutorials — but given
  their authoritative status, tutorials/quickstarts (Gas City and the public
  pack's) should be enumerated explicitly so they cannot drift unaudited.
- **[Minor] The canonical public host/org is asserted, not yet evidenced
  (carryover).** Every operator-facing surface — init templates, lock/cache
  provenance, docs, CLI messages, AC4/AC15 — inherits
  `https://github.com/gastownhall/gascity-packs.git//gastown` (lines 79, 87, 110,
  121), but nothing establishes that `gastownhall/gascity-packs` is the confirmed
  canonical source rather than a placeholder. AC15's pin ledger and AC14's
  live-network gate would catch drift at release, but the URL is asserted as fact
  across multiple ACs and propagates into docs/templates before that gate runs.
- **[Minor] Whether a checked-in worked Gastown example survives is left
  implicit (carryover).** AC4 permits former roots to be "deleted or isolated as
  non-resolvable fixtures," and the happy path (line 87) uses `gc init --template
  gastown` as the positive public-pin demonstration. That satisfies lane Q2, but
  the document does not state whether a browsable `examples/gastown` city remains
  as an operator reference (importing the public pin) or whether template-init is
  the sole worked example — relevant to source-tree cleanliness and to operators
  who learn from a checked-in example.

**Missing evidence:**
- The external `gascity-packs/gastown` pack's behavior manifest, public-pack
  docs/registry metadata, and validated pin do not exist yet (cross-repo; the
  planning worktree present is `gc-plan-pack`, not the public Gastown pack).
  Expected and gated — AC7/AC14/AC15 are blocked before implementation approval —
  but the external authority's readiness is unverifiable at this phase.
- Confirmation that `gastownhall/gascity-packs` (host, org, repo, `//gastown`
  subpath) is the real canonical public authority rather than a stand-in.
- Whether registry/catalog/discovery surfaces, for a non-Gastown city, present
  Gastown as "available, external/optional, not imported," and whether that exact
  wording is in AC12's terminology matrix.

**Required changes:**
- Enumerate tutorials/quickstarts (both Gas City and the public Gastown pack)
  explicitly in AC12's consistency scope and `docs-authority-audit.yaml`
  coverage, so the project's authoritative DX surface cannot keep presenting
  retired `examples/gastown/packs/*` or implicit Maintenance.
- State the canonical public Gastown source identity (host/org/repo/subpath) as a
  confirmed product fact anchored to its AC14/AC15 acceptance evidence, so a
  placeholder cannot leak into every inherited doc and template.
- State explicitly whether a checked-in `examples/gastown` city remains as a
  positive public-import example, or that `gc init --template gastown` is the sole
  canonical worked example, so the "non-resolvable fixture" escape hatch cannot
  silently re-establish examples as an authoritative source.

**Questions:**
- Is `https://github.com/gastownhall/gascity-packs.git//gastown` the finalized
  canonical public authority (host/org/subpath), and where is its authority
  anchored within acceptance evidence?
- For a non-Gastown city, what do registry/catalog/discovery surfaces show for
  Gastown, and is that "external/optional, not imported" wording captured in the
  AC12 terminology matrix?

**Schema conformance (bead-requested):** Conforms to `gc.mayor.requirements.v1`.
Front matter carries the required keys with `phase: requirements`, valid `status:
draft`, no downstream `*_file` keys, and no bead IDs/formula targets; all six
sections appear once in the required order; W6H is complete; Example Mapping has
happy/negative/edge rows with an Evidence column; ACs name verification; Open
Questions is `None`. The docs/terminology ACs (AC12/AC14/AC15) stay in
product-language and reference support artifacts as acceptance-evidence contracts
rather than implementation assignments — appropriate for this schema.
