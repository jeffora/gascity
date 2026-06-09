# Simone Kaye

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, Gemini (DeepSeek V4 Flash perspective)

**Consensus findings:**
- [Info] The requirements now establish the public `gascity-packs/gastown` pack, addressed by immutable `sha:` pin, as the authoritative Gastown source and explicitly deny in-tree examples, `.gc/system/packs/gastown`, bundled synthetic aliases, and implicit Maintenance fallback as runtime authority.
- [Info] The pack-resolution, offline/cache, version-skew, fresh-init, and public-validation criteria are strong enough to catch the central lane failure mode: a migration that appears to work only because stale embedded or example content still exists locally.
- [Info] Source-tree cleanliness is handled as a closure rather than a cleanup preference. AC4 and AC5 require former in-tree Gastown and Maintenance roots to be deleted or isolated from runtime resolution, docs authority, init templates, and public-pack proof, with `source-consumer-closure.yaml` classifying in-tree Gastown, mock registry, docs, and test consumers.
- [Info] Operator-facing terminology is substantially covered. AC12 requires consistent Core-required / Gastown-external / Maintenance-retired language across docs, examples, CLI help, doctor messages, import-state output, and related surfaces.
- [Major] The documentation authority audit is still too Gas City-local for an external-pack migration. The public pinned `gascity-packs/gastown` pack's own operator-facing surfaces must be included: `pack.toml` comments, command help, README/docs, formula descriptions, prompts, prompt fragments, doctor checks, generated docs, and release notes. Otherwise stale public-pack text can remain authoritative even after Gas City docs are clean.
- [Major] Registry and discovery behavior need stronger proof for the observed repository shape. The requirements should prove that fresh init, repair guidance, docs, catalog/registry lookup, and operator discovery resolve the intended immutable pack even if the current default branch or a local checkout does not contain `gastown/`.
- [Major] Public-pack stale terminology needs row-level disposition. Pinned-pack references to an "implicit maintenance/core utility layer", "maintenance pack", or "Dispatch registered maintenance formulas" must be fixed, explicitly historical/source-attribution-only, or blocked before release-pin acceptance.
- [Major] Rolling migration documentation can drift from the active config. If compatibility or coexistence mode remains possible during rollout, CLI help, doctor output, and repair guidance must reflect the active resolution mode rather than presenting only the final decoupled state.
- [Major] Retired local cache/source directories must be proven non-resolvable under manual edits, third-party imports, and mirror/offline paths. If the classifier cannot provide that proof, the plan needs a quarantine or disable step for stale directories instead of merely leaving them physically present.
- [Minor] The canonical public host/org/repo/subpath is asserted across operator surfaces but still needs acceptance evidence proving it is the final product source rather than a placeholder.
- [Minor] Docs-to-pin coherence is not explicitly included in AC15's agreement set. Hand-written docs that quote a URL, subpath, pin, or digest should match the canonical pin ledger just like lock/cache provenance and fresh-init output.
- [Minor] The post-migration fate of a checked-in `examples/gastown` worked example is unclear. The requirements should state whether it remains as a public-pin example or whether `gc init --template gastown` is the sole canonical worked example.
- [Minor] The named support artifacts are not present yet. Draft status makes this non-blocking, but approval must fail closed if the docs-authority audit, public validation, pin ledger, version-skew matrix, offline/cache proof, and acceptance-proof matrix are still missing.

**Disagreements:**
- Claude treats AC12 as strong for Gas City-local documentation consistency, while Codex and Gemini say the boundary must include the public pack's own docs, prompts, metadata, and formula text. Assessment: the public pack is part of the operator-facing product, so AC12/AC14 must audit both repositories at the pinned commit.
- Claude frames the canonical source identity as asserted-but-not-evidenced; Codex accepts the explicit URL but flags discovery proof when default-branch shape differs; Gemini accepts the authority model but focuses on rollout and cache hazards. Assessment: the URL can remain the intended authority only if AC14/AC15/AC17 provide concrete public source, registry, and discovery evidence.
- Claude sees AC4/AC5/AC16 as closing the stale source fallback surface; Gemini wants active quarantine of stale directories. Assessment: quarantine is one implementation option. The required property is stronger: retired local directories and stale cache aliases must be non-resolvable under normal resolution, manual path edits, third-party imports, mirrors, and offline repair flows, with actionable diagnostics when blocked.
- Gemini requires dynamic CLI and doctor wording for compatibility versus fully decoupled mode; Claude and Codex do not call this out. Assessment: static docs can describe the target architecture, but live CLI, doctor, and repair output must describe the user's actual active resolution mode during rollout.
- Gemini flags cache namespace collision under mirror configurations; Claude and Codex do not. Assessment: because AC16 allows mirrors or URL rewriting, cache identity should be keyed by the effective resolved remote and verified digest to prevent alias or collision bugs.
- Claude asks whether a checked-in `examples/gastown` city survives, while Gemini says old `examples/gastown/packs/*` directories are deleted. Assessment: these can both be true, but the requirements must distinguish an allowed public-pin example city from forbidden maintained in-tree pack source.

**Missing evidence:**
- `plans/core-gastown-pack-migration/support/docs-authority-audit.yaml`, including public-pack docs, metadata, prompts, prompt fragments, formula descriptions, doctor checks, generated docs, and release notes at the exact pinned commit.
- Confirmation that `https://github.com/gastownhall/gascity-packs.git//gastown` is the final canonical public source and not a placeholder.
- Public registry/catalog/discovery proof for the immutable pinned pack, including behavior when the default branch or local checkout shape differs from the pinned commit.
- Row-level stale-text disposition for public-pack references to implicit Maintenance, in-tree examples, `.gc/system/packs/*`, and retired maintenance formulas.
- Public-pack validation, pin ledger, version-skew matrix, offline/cache proof, and acceptance-proof matrix artifacts.
- Evidence that operator-facing docs references to URL, subpath, immutable commit, pack digest, and behavior-manifest digest match the canonical pin ledger.
- The decision on whether a checked-in `examples/gastown` city remains as a public-pin example.
- Proof that CLI, doctor, and repair output describe compatibility/coexistence mode when that is the active resolution state.
- Proof that stale `.gc/system/packs/*` and retired example roots cannot be reactivated by manual config, third-party imports, mirror mappings, or offline cache aliases.
- Cache namespace proof for mirrors and URL redirects.

**Required changes:**
- Amend AC12 and AC14 so docs-authority and public-pack validation include every operator-facing surface inside the pinned public `gascity-packs/gastown` pack, not only Gas City repository docs and CLI output.
- Require the cross-repo docs scan to cover `pack.toml` comments, command help, README/docs, prompts, prompt fragments, formula descriptions, doctor checks, generated docs, and release notes for retired Maintenance, in-tree example, `.gc/system/packs/*`, and implicit-core/implicit-maintenance authority language.
- Add an explicit registry/discovery proof row to AC15 or AC17 tying together source URL, subpath, immutable commit, pack digest, behavior-manifest digest, registry/catalog selector, fresh-init output, repair output, and operator-facing docs.
- State how default-branch absence of `gastown/` is treated if the release intentionally relies on an immutable historical commit.
- Anchor the canonical public source identity as a confirmed product fact with acceptance evidence.
- Extend AC15, or AC12, so operator-facing documentation references to URL, subpath, pin, and digest must match the canonical public pin ledger.
- Require row-level public-pack stale-text disposition: fixed in `gascity-packs/gastown`, allowed only as historical/source-attribution text, or blocked as live operator guidance.
- State whether a checked-in `examples/gastown` city remains and imports the public pin as a positive example, or whether `gc init --template gastown` is the sole canonical worked example.
- Require CLI, doctor, and repair output to reflect the active resolution mode when compatibility/coexistence and fully decoupled modes differ during rollout.
- Prove retired local source and cache directories are non-resolvable under manual path edits, third-party imports, mirrors, and offline cache paths, or add an explicit quarantine/disable step.
- Require cache namespace identity for mirrors and URL redirects to be derived from the effective resolved remote plus verified digest, preventing cache collision or hijacking.
