# Felix Moreau - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design treats operator wording as executable, not editorial. The generated wording matrix covers allowed and forbidden contexts, exact phrases, replacement phrases, consumers, freshness tests, golden ids, and release-gate ids.
- The docs scope is broad enough in the stronger sections: Markdown, MDX, JSON, TXT, TOML, Go strings, shell, TypeScript, generated schemas/references, CLI help, doctor JSON/text, examples, scripts, pack comments, public-pack docs, troubleshooting, and tutorial transcripts are all in scope.
- The canonical wording at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3227` gives operators a consistent model: Core is the required host system pack, `bd` and `dolt` are provider-dependent host system packs, Maintenance is retired, Gastown is an explicit public pack import, and stale generated paths are ignored legacy state rather than deletion targets.

**Critical risks:**
- [Major] The final recommended docs inventory command at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3220` is much narrower than the generated wording-linter contract. It searches only `docs examples cmd internal` and only `*.md`, `*.toml`, `*.go`, and `*.sh`, while the stronger contract includes MDX, JSON, TXT, TypeScript, generated schema/reference output, CLI/help, doctor JSON/text, public-pack docs, and docs navigation. If implementers use the final command as the actual source of truth, stale operator wording can survive in generated or non-Markdown surfaces.
- [Major] The design requires a public Gastown companion reference, but the final proposed section does not name its exact path or how Gas City verifies it. `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2633` says public Gastown owns a companion reference, while `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3260` makes companion docs a release gate. Without a concrete path and validator, the companion docs can drift from Gas City's wording matrix.
- [Minor] The final wording mixes "Core maintenance-agent behavior" at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3234` with the stronger "Core maintenance worker" terminology at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2642`. That is understandable internally, but operator-facing docs should choose one canonical phrase and reserve `dog` for compatibility/config examples.

**Missing evidence:**
- No generated `system-pack-wording.generated.yaml` exists yet, so this review could not verify actual forbidden contexts, accepted variants, docs-navigation checks, or golden ids.
- No public Gastown companion docs path is named in the design, and there is no transcript showing tutorial 01, troubleshooting, `gc doctor`, and `gc init --template gastown` with the new terminology.

**Required changes:**
- Replace the narrow final inventory command with the generated wording-linter command or explicitly label the `rg` command as a quick manual sanity check that cannot satisfy the release gate.
- Name the exact public Gastown companion reference path and require the docs/wording linter to verify that path from the pinned public Gastown checkout.
- Normalize operator terminology to "Core maintenance worker" or another single phrase across docs, doctor output, CLI help, and generated references. Treat "dog" as the default compatibility name, not the concept name.
- Add an explicit docs-navigation assertion for `docs/reference/system-packs.md` and the public Gastown companion link before the first operator-facing behavior change.

**Questions:**
- Is the public Gastown companion reference intended to live under `gastown/docs/`, and will Gas City CI read it from the exact pinned public checkout or from a local worktree?
- Should `maintenance-agent` remain an accepted variant for operator text, or should the wording matrix reject it in favor of `maintenance worker` to avoid confusing the retired Maintenance pack with an agent role?
