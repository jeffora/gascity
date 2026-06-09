# 07-pack-versioning-ecosystem - Codex Review

Persona: Saoirse Raman
Mandate: pack revision boundary; external author workflow; pinning and reproducibility; lint surface
Verdict: PASS

## Blocking Findings

None.

## Review

The current design answers the pack-versioning concerns I would block on.

It clearly removes formula-level artifact semver as the reproducibility boundary. The "Formula Artifact Versioning" section states that formula files should not carry their own semver field and that consumers pin the containing pack by semver, ref, or commit SHA, with the formula's authored revision coming from the pack revision that provided the winning layer file. Later terminology rules keep formula `version` as legacy metadata only, not a compiler selector or artifact version.

The design also preserves a concrete provenance chain from resolved formula to pack revision. Accepted artifact identity includes the formula name, source identity, content hash, ordered search path, host capability, and resolver-owned provenance. The provenance tuple covers pack/import binding identity, lockfile entry, requested ref or constraint, locked revision, dirty state, and content hash. The binding hash includes winning and losing layer identities, so shadowed formulas and layer-order changes remain explainable instead of becoming anonymous staged files.

External pack compatibility is handled conservatively enough for SHA-pinned consumers. Legacy `contract` spellings accepted by the released compiler remain deprecated aliases during the compatibility window. Alias removal is blocked until release artifacts show no supported first-party, external, SHA-pinned, transitive, or shadowed consumer still depends on the legacy spelling. Unknown external packs are treated as supported legacy consumers until the external-support artifact marks them expired or not needed with saved inventory evidence.

The pack-author migration surface is sufficiently explicit. `gc formula validate` has pack-path, pack-source, all-packs, provenance, requirement-diff, legacy-contract-report, requires-only conversion, alias-drain, and alias-removal gate modes with typed JSON exit contracts. The required `migration_hints` schema names source path, source key, current and recommended snippets, binding path, requested ref, locked revision, lockfile key, required `requires_gc` floor, and whether an automatic edit would be safe. That gives external tooling stable data instead of forcing pack authors to parse diagnostic prose.

The docs and stale-guidance gates close the main ecosystem drift path. The design requires `docs/reference/formula.md`, PackV2 author docs, CLI help, generated schemas, examples, tutorials, and release-report schemas to update in the same branch that exposes the relevant behavior. The proposed `make formula-docs-check` covers stale `contract`, stale `version`, pack `requires_gc`, pack import refs, pinned SHA behavior, and generated validation flags.

## Non-Blocking Watchpoints

- The first implementation PR must actually add `gc formula validate` before any user-visible requirement diagnostic lands. The design depends heavily on that command as the pack-author lint surface.
- Packman schema 2, or its equivalent typed provenance API, is load-bearing. If it slips, Phase 7b, Phase 8, and Phase 9 must remain blocked rather than substituting advisory reports.
- The docs rewrite should keep the external-pack fixture runnable, not just illustrative, so pack authors can copy the formula and `pack.toml` pair without resolving hidden placeholders.

## Required Changes

No changes required from this persona.
