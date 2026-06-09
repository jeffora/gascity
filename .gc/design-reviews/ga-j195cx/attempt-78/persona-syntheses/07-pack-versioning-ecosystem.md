# Saoirse Raman

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The design keeps pack revision, not formula `version`, as the artifact identity. Both reviews approve the separation among pack refs, lockfile entries, content hashes, accepted artifact provenance, formula compiler requirements, and legacy formula metadata. This is the correct ecosystem boundary and avoids turning formula files into their own semver surface.
- [Major] The external validation surface is strong, but its credibility depends on packman provenance being explicitly sequenced. Claude flags packman schema 2 or an equivalent packman-owned contract as a load-bearing prerequisite for content hash, requested ref, locked revision, parent binding, transitive depth, dirty state, registry mirror, and `requires_gc` evidence. Codex accepts the design but also names those generated release artifacts and provenance inventories as implementation-time evidence that must exist before rollout gates pass.
- [Major] External pack maintainer and consumer migration need more operational proof. The design specifies useful commands and reports, including `gc formula validate --pack-path`, `--pack-source`, `--ref`, `--requirement-diff`, structured `migration_hints`, support artifacts, and alias-removal gates. It still needs a concrete discovery and outreach process for known external packs plus a worked consumer path for SHA-pinned or ref-pinned packs so users do not discover the migration only through import or dispatch failures.
- [Minor] The JSON contract for external migration hints has one important gap. Claude specifically calls out `formula.migration.pin_pack_revision` as load-bearing for external pinned-pack migrations, but the design only gives a full example for `formula.migration.add_requires`.
- [Minor] Schema-1 packman fallback needs an explicit user-visible diagnostic. Failing closed is correct, but callers should receive a typed cause such as `pack.schema1_insufficient` when schema 1 cannot prove release-gate evidence rather than inferring that from absent fields.
- [Minor] The current context docs still teach pre-migration formula shape, including formula `version` as a common top-level key and no canonical `[requires]` section. Codex considers the planned same-branch docs, doctest, stale-guidance, generated help, and schema gates sufficient if implementation preserves them.

**Disagreements:**
- Claude returns `approve-with-risks`; Codex returns `approve`. Assessment: choose `approve-with-risks`. The design direction is sound, but the ecosystem lane should not treat packman provenance, external outreach, and external-consumer examples as optional cleanup.
- Claude treats packman schema 2 tracking and external-maintainer outreach as required design additions. Codex does not block on them and frames remaining evidence as implementation-time release artifacts. Assessment: retain Claude's required changes in reduced form because they determine whether external pack support can be evaluated before alias removal.
- Claude asks for a typed schema-1 insufficiency diagnostic and a worked `pin_pack_revision` hint. Codex does not mention these. Assessment: keep both as required contract clarifications because they affect machine-readable tooling consumers.
- Codex flags stale existing docs as the only risk; Claude focuses more on provenance and external ecosystem process. Assessment: these are complementary. The docs gates should remain phase-blocking, while the provenance and outreach gaps should be made explicit in the design.

**Missing evidence:**
- No Kimi 2.6 review artifact was present.
- No concrete packman schema 2 issue, owner, PR home, or phase gate is linked to the provenance fields required by alias-removal, external pinned-pack support expiration, imported-pack floor enforcement, and requirement-diff reports.
- No external-pack discovery and maintainer outreach process explains how the release captain seeds the supported-external list, contacts known maintainers, or decides that an `active` pack has a named migration path before flipping it to `expired`.
- No worked external-consumer recipe ties together `formula.contract_deprecated`, `gc formula validate --pack-source <url> --ref <ref> --json`, `migration_hints`, upstream fix or pinning, `--requirement-diff old.lock new.lock`, and lockfile update behavior.
- No full JSON example defines the `formula.migration.pin_pack_revision` hint shape, including authoritative field names for recommended ref, recommended revision, and evidence source if those fields are intended.
- No typed `pack.schema1_insufficient` or equivalent diagnostic is specified for schema-1 packman evidence gaps.
- The generated release artifacts, external-support rows, compatibility corpus output, provenance inventory, doctest output, and stale-guidance reports are not present yet; they remain rollout-phase evidence.

**Required changes:**
- Add a concrete packman schema 2 or equivalent provenance-contract tracking reference. Name the required lockfile fields, schema migration behavior, schema-1 compatibility/fail-closed behavior, and which rollout phases are blocked until that evidence exists.
- Define the external-pack discovery and outreach process. Include how the release captain seeds the supported-external list, what evidence proves a named migration path, what notification is sent to known maintainers, and what allows `status: active` to become `expired`.
- Add a worked JSON example for `formula.migration.pin_pack_revision`, parallel to the existing `formula.migration.add_requires` example.
- Require `gc formula validate --pack-source` to emit a typed schema-1 insufficiency diagnostic when old packman metadata cannot prove a release-gate field.
- Add a consumer-facing external-pack migration recipe to the reference docs or a sibling guide, covering deprecated contract diagnostics, validation, migration hints, upstream coordination or temporary pinning, requirement diff, and lockfile update.
- Preserve the same-branch docs/doctest/stale-guidance/generated-help gates so stale formula `version` guidance is removed before the corresponding diagnostics become user-visible.
