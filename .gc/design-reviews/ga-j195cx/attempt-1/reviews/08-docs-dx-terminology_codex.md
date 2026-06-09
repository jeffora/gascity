# 08 Docs DX Terminology - Codex Review

Persona: Avery Brooks
Mandate: documentation consistency; TOML user experience; terminology clarity; open question closure
Verdict: block

## Findings

[Blocker] Pack compatibility snippets are required to be copy-paste-safe/runnable but are not valid pack examples.

The design requires every TOML snippet in `docs/reference/formula.md` to parse or be explicitly invalid, then immediately requires pack snippets that omit the required `schema` key:

- `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4722` introduces "exact TOML snippets".
- `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4731` shows `[pack] name = "acme-workflows"` plus `requires_gc`, but no `schema`.
- `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4770` says external-pack examples must be "one runnable fixture".
- `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4781` repeats the incomplete `pack.toml`.

That conflicts with the actual pack contract: `internal/config/pack.go:2174` validates pack metadata and fails when `[pack].schema` is missing. It also conflicts with the current PackV2 docs, whose minimal pack example includes `schema = 2` at `docs/guides/shareable-packs.md:58`.

This is a user-visible docs/DX blocker because the proposed reference page is supposed to teach the distinction between formula `[requires]`, pack `requires_gc`, and pack revision. The required snippet currently teaches an invalid pack shape at exactly the point where users are being asked to edit pack metadata. It also makes the doctest design ambiguous: the fixture says valid snippets go through the formula loader, but these are pack snippets, not formula snippets.

Required fix: make the pack snippets complete (`schema = 2` or the current generated floor/schema value) and extend the doctest fixture schema so each TOML block declares its snippet kind/path, with formula snippets validated by the formula loader, pack snippets by the pack/config loader, and city snippets by the city config loader. If a placeholder template is shown, require a paired resolved fixture that is actually runnable.

[Major] The stale-guidance matcher for `formula_compiler` does not scan the TOML example paths it claims to protect.

The design says the stale-guidance matcher catches "`formula_compiler` examples that omit the top-level `[requires]` context" (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5007`), and the global path globs include `examples/**/*.toml` plus formula testdata (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4880`). But the actual `requires-positive-content` matcher only scopes itself to `docs/reference/formula.md`, `docs/tutorials/**/*.md`, and `examples/**/*.md` (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:4935`).

That leaves authored TOML examples and fixtures relying on the separate inventory/parser checks rather than the explicit stale-guidance rule, so the rule can pass while the report fails to enforce its stated coverage. For a docs/DX gate, that mismatch will make failures harder to understand and creates a blind spot for inline comments or non-formula TOML snippets.

Required fix: include TOML-bearing example/testdata globs in the matcher or split it into prose and parsed-TOML checks with separate names and reports. The output should make clear whether a failure came from stale prose, a misplaced key in TOML, or an inventory/parser validation failure.

## What Looks Solid

The current design closes the original high-risk terminology questions well: it defines a glossary, separates formula `[requires]` from host `[daemon] formula_v2`, pack `[[pack.requires]]`, pack `requires_gc`, formula `version`, schema, and pack revision, and makes same-branch docs/generated artifacts a predecessor to user-visible diagnostics. The phase ordering also correctly prevents "docs follow-up" from being accepted after diagnostics are visible.

## Open Questions

Should the docs reference page own pack/city config snippet validation itself, or should `make formula-docs-check` delegate those snippets to the existing config docs/schema test loop? Either is fine, but the current design needs to choose one so the reference page can safely include pack and city TOML.
