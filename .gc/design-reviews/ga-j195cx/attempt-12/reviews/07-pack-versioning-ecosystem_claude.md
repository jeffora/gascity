# Saoirse Raman - Claude

**Verdict:** approve-with-risks

The design names the boundary correctly: `[requires]` declares what the
formula needs, the host decides whether it can satisfy that need, and
pack semver/ref/SHA is the durable identity for "which authored workflow
am I running." That kills the "version becomes de facto formula semver"
failure mode at the conceptual layer, and the auditable-provenance table
(lines 585–595) is the right shape on paper. What is weak is everything
the *ecosystem* depends on: legacy `version` is preserved silently on
every runtime path, the only external-author alias-removal gate is a
release-checklist note, the lint surface diagnoses but does not migrate,
the `--legacy-contract-report` and `--provenance` commands have very
different specification rigor (one has a JSON schema, the other has only
prose), and reproducibility metadata is recorded with no operational
consumer. None of these are existential blockers; all of them must close
before first-party requires-only conversion (Rollout Phase 6), or "version
becomes de facto formula semver" and "external pack authors only get a
changelog note" both come true in practice.

## Top strengths

- The Decision section and "Why `requires`" together remove the
  selector/identity confusion that `contract` and `version` were creating.
  Combined with Goals 3–4, the Formula Artifact Versioning section, and
  the Glossary entry for "Pack revision" as the artifact identity and
  reproducibility boundary, the doc is unambiguous that formula files do
  not carry their own semver and pack pinning is the consistency
  mechanism. The Non-Goals (lines 56–60) close the door on per-formula
  semver and remote-pinning changes, which is the right scope discipline.
- The auditable-provenance table for every new workflow root (lines
  585–595) names the fields a consumer needs to tie a winning layer file
  back to its pack revision: `gc.formula_pack_name`,
  `gc.formula_pack_source`, `gc.formula_pack_ref`,
  `gc.formula_pack_revision`, plus a typed `gc.formula_reproducibility`
  state and a `gc.formula_compile_artifact` escape hatch when metadata
  size is bounded. At the data-model layer, this directly addresses
  "layer winners cannot be tied back to pack revision."
- The alias-removal gate is criteria-based, not time-based, with three
  measurable first-party signals (two-minor-release floor for
  `[requires]`, exit-0 from `--legacy-contract-report`, CI stale-guidance
  check). `--legacy-contract-report` itself has a stable JSON schema and
  exit-code contract (lines 478–494), which is the right shape for a CI
  gate.

## Critical risks

- **[Major] Silent preservation of `version` on every runtime path
  invites de-facto formula semver.** Lines 663–667 emit
  `formula.version_deprecated` only on validate/show/preview surfaces and
  explicitly suppress it on launch, order dispatch, retry, convergence,
  and controller paths "so operational logs are not polluted." Combined
  with no removal timeline, no bound on accepted values (any integer?
  negatives? `2.0`?), and no `--strict` mode that fails on `version`
  presence in first-party packs, the design teaches authors that
  `version = N` is a free-form metadata field they can keep bumping. That
  is the named red flag. The current public reference
  (`docs/reference/formula.md` line 47) still describes `version` as an
  "optional formula version marker," and the design lists that file under
  required updates without stating the canonical post-migration author
  guidance for `version` (remove? keep? what value?). One-time
  deprecation warnings do not stop snippets from re-spreading.

- **[Major] The single external-author alias-removal gate is a paperwork
  gate, not a measurable one.** The four alias-removal gates (lines
  449–457) include three measurable first-party signals and one
  external-author gate that reads "The release checklist records that
  externally pinned legacy formulas remain supported by either the alias
  or a documented compatibility branch." The compatibility branch is
  named nowhere else in the document — owner, supported-version matrix,
  opt-in mechanism (separate binary, build tag, env toggle), duration,
  and communication channel are all undefined. There is no published
  deprecation timeline or notice period, no telemetry on legacy-`contract`
  usage from registered cities, and no
  `gc formula validate --pack <git-url>@<ref>` for an external author to
  verify their own pack against a current binary. As written, the design
  realizes the named red flag — *external pack authors only get a
  changelog note* — almost exactly.

- **[Major] No migration tooling for pack authors.** The lint surface
  diagnoses but does not act. There is no `gc formula validate --fix`,
  no `gc pack lint`, no rewrite path that turns
  `contract = "graph.v2"` into `[requires] formula_compiler = ">=2"` (or
  adds `[requires]` while preserving `contract` for the alias window).
  External pack authors with N formulas migrate by hand, and hand-edited
  dual declarations are easy to drift — which then trips
  `formula.compiler_requirement_conflict` (line 444) and sends the author
  back to fix it. A migration helper is the single highest-leverage
  thing that would shorten the alias window measurably.

- **[Major] `--legacy-contract-report` and `--provenance` have very
  different specification rigor.** `--legacy-contract-report` has a
  stable JSON schema, exit codes 0/1/2 with documented meanings, and is
  named explicitly as a release-gate command (lines 474–494). Good.
  `gc formula validate --provenance` is described in prose only (lines
  610–624) — no JSON schema, no exit-code contract, no statement of
  whether it is intended for human reading or scriptable, no example. As
  written, an external pack-author CI cannot pin against `--provenance`
  output, and the persona's "lint surface" mandate fails for the case
  most relevant to the persona — pack-author CI verifying reproducibility
  before publishing a tag.

- **[Minor] `gc.formula_reproducibility = "local-not-reproducibly-pinned"`
  has no operational consumer.** Line 594 lists it as a possible value
  and `--provenance` reports it. Nothing in the design says it blocks
  production launches, gates controller dispatch, surfaces in the
  dashboard as a "this city is not reproducible" indicator, or appears
  in any release-gate command. As a metadata-only marker it is
  decorative; the pinning/reproducibility mandate from this lane needs
  at least one enforcement path that consumes this field.

- **[Minor] The `--legacy-contract-report` scope does not distinguish
  first-party from external packs.** The alias-removal gate reads "zero
  first-party legacy-only formulas" (line 451), but the report's `items`
  schema records `whether the formula is first-party or external` (line
  490) — yet the exit-code contract collapses both into the same exit-2
  signal ("first-party legacy-only formulas remain"). Line 451's gate is
  consistent with the narrow interpretation, but then there is no
  equivalent gate or signal for imported third-party packs that a city
  consumes, and a release captain has no scriptable way to ask "what
  fraction of cities I know about still depend on legacy `contract`?"
  This is the observability surface that gate #3 of alias removal
  silently relies on.

- **[Minor] Pack composition (pack-imports-pack) deprecation routing is
  unaddressed.** Provenance records the *winning* file's pack, but if
  pack A imports pack B and B ships `contract = "graph.v2"`, the city
  operator running A sees a deprecation diagnostic sourced to B that
  only B's author can fix. The remediation message is correctly
  addressed to whoever can fix it, but the design does not say how the
  operator silences/scopes that warning while waiting on B, how
  `--legacy-contract-report` classifies "I imported it, I didn't author
  it," or whether per-step pack attribution is recorded for compiled
  formulas spanning multiple packs through `extends`, `compose.expand`,
  or `compose.aspects`.

- **[Minor] Pack-level `requires_gc` vs. formula-level `[requires]`
  interaction is glossary-only.** Line 645 marks them as distinct, but
  the body never tells a pack author when bumping a formula's
  `formula_compiler` from `>=1` to `>=2` should also bump pack
  `requires_gc` (or pack semver). Without that guidance, two reasonable
  authors will diverge on what "we now need formula compiler v2" means
  at the pack manifest layer, and SHA-pinned consumers may break in
  surprising ways during binary upgrades. SHA-pinned consumers should
  be able to fail fast at *pack import* rather than at sling time.

## Missing evidence

- The `Provenance` type referenced in `CompileResult` (line 327) has no
  struct definition. The reader cannot tell what fields are populated,
  whether the type travels through callers as data or only as a string
  projection, or whether the auditable-provenance metadata table at
  lines 585–595 is its persisted shape. The diagnostics struct is fully
  typed; provenance is not.
- No worked example of an external pack at SHA `abc123` that contains
  `contract = "graph.v2"` traversing the alias window (works, with
  warning), then alias removal (fails — with what diagnostic, against
  what binary version, surfaced to whom). The compatibility matrix
  covers source-shape variants but not the SHA-pinned-external scenario
  end to end.
- No statement on whether `gc formula validate --all-packs` scans
  imported third-party packs in a city's resolved layer set, or only
  first-party packs vendored into the gc tree. The alias-removal gate
  language ("zero first-party legacy-only formulas") is consistent with
  the narrow interpretation; the `--legacy-contract-report` items schema
  hints at the broader one. The doc should commit.
- No definition of "documented compatibility branch": branch of what
  artifact, owned by whom, supported how long, opted into how, with
  what removal criteria.
- No bound on accepted `version` values, no rows in the compatibility
  matrix exercising `version` interactions with `[requires]` and
  `contract`, and no alias-window-style criteria for ever removing the
  field.
- No statement of whether `--provenance` output is a stable JSON
  contract that external pack-author CI can pin against, or a
  human-readable surface only.
- No coverage of how `ResolveFormulas` (`cmd/gc/formula_resolve.go`)
  carries pack provenance through to the staged `.beads/formulas/`
  symlinks. The provenance keys are required at compile time; if
  resolution drops pack identity, `gc.formula_pack_revision` cannot be
  reliably stamped at root creation.
- No source-type normalization rules for `gc.formula_pack_source`. A
  `git+ssh://` import and an `https://` import of the same repo at the
  same SHA should normalize to the same provenance, but the design lists
  the value taxonomy without normalization rules.

## Required changes

1. State the canonical author guidance for legacy `version` and add a
   measurable removal plan. Minimum: (a) update
   `docs/reference/formula.md` examples to remove `version` and mark it
   deprecated rather than "optional formula version marker"; (b) add a
   `--strict` mode to `gc formula validate` that fails on `version`
   presence in first-party packs; (c) add `version` rows to the
   compatibility matrix; (d) record an alias-window criterion for
   `formula.version_deprecated` becoming a hard error in the rollout
   plan, mirroring the discipline used for `contract`.

2. Replace the external-author alias-removal gate with at least one
   measurable signal. Acceptable forms: a published minimum notice
   period (e.g. "≥90 days from changelog header X to alias removal"),
   a telemetry feed of legacy-`contract` usage from cities that opt
   into reporting, or extending `--legacy-contract-report` so it
   classifies findings as `first_party` vs. `external` with separate
   exit codes that a release captain can compose. Either the release
   checklist consumes one of these, or the gate is not measurable.

3. Replace "documented compatibility branch" with a concrete plan, or
   drop it. At minimum: who maintains it, the supported `gc` version
   matrix that retains `contract` parsing, the consumer opt-in mechanism
   (build tag, separate binary, env toggle), the removal criteria, and
   the explicit communication channel for external pack authors. A
   release-notes line is not sufficient — that is the red flag.

4. Add a migration helper that rewrites `contract = "graph.v2"` to
   `[requires] formula_compiler = ">=2"` with a flag to preserve
   `contract` during the alias window for dual-declared output. Specify
   TOML comment/whitespace preservation behavior. Even a minimal
   `gc formula validate --fix` covering the single most common stale
   snippet would shorten external-author migration friction by an
   order of magnitude.

5. Specify operational consequences for
   `gc.formula_reproducibility = "local-not-reproducibly-pinned"`.
   Required: at least one of (a) a `gc formula validate
   --require-reproducible` mode usable in CI, (b) a dashboard surface
   that flags non-reproducible roots, or (c) launch-time refusal under
   a `[daemon] require_reproducible_formulas = true` toggle. Recording
   the value with no consumer turns reproducibility into decoration.

6. Specify the `--provenance` output contract (JSON schema + sample +
   exit codes), parallel to the `--legacy-contract-report` contract.
   Document source-type normalization rules
   (`gc.formula_pack_source`) so two consumers importing the same pack
   via different transports stamp identical provenance.

7. Define the `Provenance` type alongside `NormalizedRequirements` and
   `Diagnostic`. State whether it is persisted to the workflow root
   (the auditable-provenance table at lines 585–595 implies yes; the
   type is invisible). Specify per-step pack attribution for compiled
   formulas spanning multiple packs through `extends`/expansion/aspect
   chains — multi-pack contribution is a real reproducibility case
   under the existing layer model.

8. Document pack-composition deprecation routing. When pack A imports
   pack B and B uses `contract`, define: (a) which pack the warning
   attaches to in `Diagnostic.Formula`/`SourcePath`, (b) how a city
   operator suppresses or scopes the warning while waiting on B, and
   (c) how `--legacy-contract-report` classifies imported-only legacy
   usage for first-party-vs-external accounting.

9. Add a section on pack-level `requires_gc` interaction with
   formula-level `[requires]`. Minimum: "if any formula in a pack
   declares `[requires] formula_compiler = ">=2"`, the pack manifest
   must declare pack-level `requires_gc` ≥ the minimum binary that
   supports compiler v2." This is the only way SHA-pinned consumers
   fail fast at pack import rather than at sling time.

## Questions

- For a brand-new formula written today, what is the canonical author
  guidance for `version`? Remove it entirely, set it to 1, or leave
  whatever value and ignore the warning? The design says "preserved
  only as legacy metadata" but does not tell new authors what to write.
- Is the `bd` shell-out path being taught `[requires]`, or is it being
  removed before requires-only conversion (per
  `engdocs/proposals/formula-migration.md` Phase 4)? The compatibility
  matrix only works if exactly one of these is true; the design should
  commit.
- After alias removal, what is the published expectation for an
  external pack at a frozen SHA whose formula still contains
  `contract = "graph.v2"`? Does the consumer get a clean error with the
  source path and a remediation pointing to the alias-removal release
  notes, or does compilation fail more diffusely at dispatch time?
- Does `gc formula validate --all-packs` scan imported third-party
  packs in a city's resolved layer set? Does the result count toward
  the alias-window gate, or only first-party formulas? The current
  wording is consistent with the narrow interpretation, but then there
  is no equivalent gate or signal for imported third-party packs.
- For `extends`, `compose.expand`, and `compose.aspects` chains spanning
  multiple packs at different pinned revisions, does `--provenance`
  (and any persisted `gc.formula_pack*` metadata) report per-step pack
  attribution, or only the root formula's pack?
- Does the design intend any in-product warning when layer resolution
  results in an external pack overriding a first-party formula, or vice
  versa? Reproducibility is undermined when a floating layer can change
  the resolved formula even though every input the consumer can see is
  pinned.
- What is the discovery surface for external pack-author-affecting
  changes? Is there a CHANGELOG section, release-notes header, or
  mailing-list channel that external authors can subscribe to, distinct
  from internal-change noise?
- When a pack bumps a formula from `>=1` to `>=2`, is that pack-semver-
  major, pack-semver-minor, or up to the author? The design treats pack
  semver as opaque, but consumer pinning behavior depends on the answer.
