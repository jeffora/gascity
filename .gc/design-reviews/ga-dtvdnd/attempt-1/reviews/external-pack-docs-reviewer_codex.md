# External Pack Docs Review - Codex

Persona: Simone Kaye  
Mandate: external Gastown pack authority; registry behavior; source-tree cleanliness; documentation consistency  
Verdict: approve-with-required-changes

## Summary

The current requirements do establish the right product direction: Gastown behavior moves to the public `gascity-packs/gastown` authority, Core remains the only required Gas City-owned runtime pack, Maintenance is retired, and docs/diagnostics/registry surfaces must stop treating in-tree examples or `.gc/system/packs/*` as authoritative. AC4, AC12, AC14, AC15, and AC16 are especially strong and close most of the external-pack/docs failure modes.

The remaining issues are not about the desired model. They are about proof contracts that still leave room for stale examples or public-pack documentation to pass review without a stable, auditable evidence artifact.

## Findings

### [Major] Public Gastown validation is required but has no stable artifact contract

AC14 requires public Gastown validation, and AC17 says the acceptance-proof matrix must map it to concrete evidence. The problem statement also says public Gastown validation is one of the named support artifacts required before implementation approval. However, unlike the pack-resolution matrix, docs-authority audit, pin ledger, version-skew matrix, and diagnostics schema, the public-pack validation proof has no explicit path, row schema, required fields, or drift rules.

This creates a review gap: the release could satisfy AC14 with scattered command output or a one-off matrix row while leaving no durable artifact that proves the pinned public pack's docs, registry/catalog metadata, prompts, commands, formulas, generated references, and operator-facing text were checked at the exact immutable commit.

Required change: either add a named artifact such as `plans/core-gastown-pack-migration/support/public-gastown-validation.yaml`, or explicitly make `acceptance-proof-matrix.yaml` the schema-checked home for AC14 row-level evidence. The contract should require source URL, subpath, immutable commit, pack digest, behavior-manifest digest, checked operator surfaces, command outputs, live-network release-gate result, and deterministic pinned-cache/local-fixture result.

### [Major] The checked-in `examples/gastown` outcome remains ambiguous

The requirements retire `examples/gastown/packs/gastown` and `examples/gastown/packs/maintenance`, and AC4 says former in-tree roots must be absent or non-resolvable fixtures. They do not clearly state whether a checked-in `examples/gastown` worked example may remain after the split, and if so what authority it has.

That ambiguity matters because examples are operator-facing documentation. A remaining example city could accidentally become the tutorial authority even if its `packs/*` subtree is gone, especially if docs or tests continue to link to it. Conversely, deleting the example without replacement could remove an important public-pin witness for Gastown init.

Required change: add an explicit requirement for the top-level `examples/gastown` fate. Either remove it as a maintained example, or retain it only as a public-pin example that imports `https://github.com/gastownhall/gascity-packs.git//gastown` at the canonical immutable pin and is covered by docs-authority, pin-coherence, and no-retired-load checks.

### [Minor] Pin coherence omits examples and tutorials as first-class checked surfaces

AC15 requires source, subpath, immutable commit, pack digest, and behavior-manifest digest to agree across the pin ledger, compatibility proof, lock/cache provenance, docs/registry/repair output, and fresh-init output. AC12 separately audits docs and examples for authority language, but the pin-coherence gate does not explicitly include examples, tutorials, generated quickstarts, or checked-in sample `pack.toml` / rig-import snippets.

The result could be clean authority wording with stale pins in examples. That is still an operator-facing failure, especially during a two-repository release where the public pack commit is the product contract.

Required change: extend AC15 or AC12 so examples, tutorials, generated references, sample configs, checked-in example `pack.toml` files, and quickstart snippets are validated against `public-gastown-pin-ledger.yaml` whenever they quote the public URL, subpath, immutable commit, pack digest, or behavior-manifest digest.

## Positive Coverage

- AC4 correctly rejects in-tree Gastown, `.gc/system/packs/gastown`, bundled synthetic aliases, and implicit Maintenance as fresh-init authorities.
- AC12 correctly extends documentation consistency beyond Gas City-local docs to pinned public `gascity-packs/gastown` docs and comments.
- AC14 correctly prevents a local in-tree copy from masking a broken external pack and requires public docs/registry/catalog validation.
- AC15 and AC16 correctly tie public pack pins, cache behavior, version skew, mirrors, and offline behavior to fail-closed provenance checks.
- AC17 correctly requires an acceptance proof matrix before decomposition; it just needs to make the AC14 public-pack proof durable enough to audit.

## Required Changes

1. Add a named, schema-checkable home for public Gastown validation evidence, or make the acceptance-proof matrix explicitly own that row-level schema.
2. State the post-migration fate and constraints for any checked-in `examples/gastown` worked example.
3. Include examples, tutorials, generated quickstarts, sample configs, and checked-in example snippets in the public-pin coherence gate.
