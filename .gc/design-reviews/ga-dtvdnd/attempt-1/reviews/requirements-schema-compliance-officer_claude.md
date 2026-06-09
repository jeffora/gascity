# Mara Voss - Claude

**Verdict:** approve-with-risks

> Lane: requirements schema compliance, W6H completeness, example-mapping
> readiness, AC behavior-focus. Reviewed the *current*
> `plans/core-gastown-pack-migration/requirements.md` (front-matter `updated_at:
> 2026-06-09T17:23:58Z`, `status: draft`) against the non-empty
> `gc.mayor.requirements.v1` schema. Every structural claim and line citation is
> re-derived from the current document; an earlier pass at this path reviewed a
> prior revision (`updated_at: 15:35:47Z`) whose example counts and line numbers
> no longer match. Current Example Mapping is 5 happy / 4 negative / 7 edge.

**Top strengths:**
- Exact schema conformance, front matter and body. Front matter (L1–10) is
  precisely the eight required keys in schema order with `phase: requirements`, a
  valid `status: draft`, both path keys absolute, ISO timestamps, and none of the
  forbidden `requirements_file` / `implementation_plan_file` / `design_file` keys.
  All six body sections appear once each in the mandated order — Problem Statement
  (L14) → W6H (L70) → Example Mapping (L82) → Acceptance Criteria (L103) → Out Of
  Scope (L125) → Open Questions (L146) — under a correct `# Requirements: <title>`
  heading (L12). Schema red flags #1 (missing sections / unsupported front matter)
  and #2 (approved-without-W6H/example-mapping; here `status: draft`) both clear.
- W6H and Example Mapping are concrete enough for design without downstream
  inference (lane Q2). W6H (L72–80) answers all seven dimensions with
  non-placeholder content — named stakeholder classes, the three legacy roots, the
  explicit public source URL with `sha:` pin, the repair surface, and a real
  how-much/scale enumeration. Example Mapping (L85–101) exceeds the schema floor
  with 5 happy / 4 negative / 7 edge rows, each carrying both an Expected-behavior
  and an Evidence cell, including the hard cases (offline cache-hit/miss/digest-
  mismatch, in-flight retired-content sessions, pin-window skew, diamond conflicts,
  the `port_resolve.sh`→`dolt-target.sh` helper-closure case).
- The artifact polices its own scope boundary and stays a requirements artifact.
  No bead IDs, formula targets, production-source file assignments, or appended
  review-attempt summaries appear anywhere. The Problem Statement pre-labels the
  support artifacts as "acceptance evidence, not inline implementation design"
  (L49–54); Out Of Scope excludes file moves, code changes, and bead creation
  during the requirements phase (L136–137) and gates approval on the
  AC6/AC7/AC14–AC17 proofs passing (L143–144); Open Questions is a clean `None`
  (L148).

**Critical risks:**
- [Major] Compound acceptance criteria undercut per-criterion verifiability and
  the traceability the document itself demands. AC3 (L109), AC6 (L112), AC7
  (L113), AC10 (L116), and AC11 (L117) each bundle 8–15 distinct sub-requirements
  behind a single ID with one Verification cell listing a dozen test types. The
  schema requires each criterion to be "verifiable by a unit test, integration
  test, command verification, or explicit manual check" — a 15-part cluster is
  verifiable by a *suite* but not as a single criterion, so no one test maps to one
  criterion. This is the sharpest readiness risk in my lane because the document
  itself requires fine-grained traceability the clusters cannot expose: AC6 (L112)
  demands "a bidirectional link between source behavior IDs and AC7 witnesses at
  behavior-row or call-site granularity," which a monolithic criterion cannot
  surface. It does not breach the schema's hard rules (no implementation files,
  helper names, formula targets, or bead IDs leak), so it is not a blocker — but it
  should be split before the document is used as a decomposition basis.
- [Minor] Support-artifact paths and formats are hardcoded in the Acceptance
  Criteria. The ACs name ~11 exact artifact paths/filenames/formats — e.g.
  `support/pack-resolution-matrix.yaml` (AC3, L109),
  `support/behavior-preservation-manifest.yaml` (AC7, L113),
  `support/migration-diagnostics.schema.json` (AC11, L117),
  `support/version-skew-matrix.yaml` (AC15, L121). The schema's iteration rules
  prohibit "choosing implementation files, helper names." The document mitigates
  this by labeling the artifacts "acceptance evidence, not inline implementation
  design" (L49–54) and by AC17 (L123) designating them binding evidence contracts,
  which legitimizes their *existence* as verification anchors — but it still
  pre-commits exact filenames and `.yaml`/`.json` formats that are a design-phase
  choice. This is the document's principal brush with the "don't choose files"
  rule, and is defensible rather than a violation.
- [Minor] Mechanism phrasing in a few criteria and in the Problem Statement. AC16
  (L122) prescribes "randomized or process-unique staging paths" — a technique;
  the behavior-focused outcome is "concurrent cache writes never publish partial,
  corrupt, or unproven pack state." Separately, the Problem Statement's
  pack-resolution paragraph (L56–68) enumerates a resolved-config field schema
  ("source, version or digest, materialized path, lock/cache provenance, and
  collision state"); that representation detail is already carried — correctly — by
  AC2 (L108) and AC3 (L109), so stating it in the problem framing duplicates
  mechanism into product language.

**Missing evidence:**
- Whether the precedence calls, `bd`/`dolt` provider-pack cardinality/mutual-
  exclusion policy (AC3, L109), and the version-skew compatibility *window width*
  (AC15, L121) deferred to the support matrices are *resolved* product decisions or
  unresolved ones parked in not-yet-existing artifacts. The document claims
  `Open Questions: None` (L148), yet AC3/AC6/AC7/AC15/AC17 describe those decisions
  as future deliverables that "must exist and pass before implementation approval."
  By the document's own framing (L49–54, L143–144) this is an acknowledged
  acceptance gate rather than a hidden product unknown, so `None` is defensible —
  but the line between "deferred to a proof artifact" and "unmade product decision"
  is the one place the `None` claim is not fully self-evident.
- Whether the named `support/*.yaml|.json` artifact paths are normative (binding
  paths/formats) or illustrative acceptable evidence. The document does not say.

**Required changes:**
- Split the compound criteria (AC3, AC6, AC7, AC10, AC11 at minimum) into atomic,
  independently verifiable criteria or numbered sub-IDs, each with a single
  Verification anchor, so each criterion maps to one test, command, or manual check
  and the AC6↔AC7 bidirectional traceability the document requires becomes
  expressible. (Sharpest item; resolve before decomposition.)
- (Recommended) Restate the AC/W6H references to support artifacts as required
  *outcomes and evidence* (e.g. "a validated asset-migration ledger exists proving
  every active legacy pack asset maps to a Core/Gastown/retired successor with no
  dropped, duplicated, or basename-collision assets") and move exact filenames and
  `.yaml`/`.json` formats to the design phase — or state explicitly (e.g. in Out Of
  Scope) that the paths are normative acceptance deliverables, which would close
  the path-leakage finding outright.
- (Recommended) Return the Problem Statement's pack-resolution paragraph (L56–68)
  to product language, letting AC2/AC3 carry the resolved-config field specifics so
  they are stated once, in acceptance scope; and reword AC16's "randomized or
  process-unique staging paths" to the observable no-partial/corrupt-cache outcome.

**Questions:**
- Are the `support/*.yaml|.json` artifact names normative requirements (binding
  paths/formats that must ship exactly as written) or illustrative examples of
  acceptable evidence? The artifact self-labels them "acceptance evidence" (L49–54)
  and AC17 calls them binding contracts; stating that intent explicitly in Out Of
  Scope or the Problem Statement would let a schema reviewer treat the field-level
  detail in AC3/AC6/AC7 as legitimate acceptance scoping and close the [Minor]
  path-leakage finding.
- Does deferring pack-resolution precedence, `bd`/`dolt` provider-pack cardinality,
  and the version-skew window width to the matrix artifacts constitute "no open
  product questions," or are some of those calls still unmade product decisions
  that belong in Open Questions with `status: questions`?
