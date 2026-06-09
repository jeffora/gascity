# Elias Sato

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek, Gemini

**Consensus findings:**
- [Blocker] Core inclusion is still not proven by the proposed provenance assertion. The design relies on `assertRequiredSystemPackProvenance` without defining a safe identity signal. A raw path/name check can miss the validated system Core, accept a user/imported pack named `core`, or fail when Core contributes no effective keys. The design needs a typed system-pack participation/content-identity contract, not incidental `Sources`/agent provenance.
- [Blocker] Loud failure for missing, corrupt, stale, or partial Core must be defined as two separate fatal checks. Manifest integrity (`unusableRequiredBuiltinPackNames` / `packContainsEmbeddedManifest`) catches corrupt and partial content; provenance presence catches "materialized but dropped from resolved config." Provenance alone does not satisfy the lane. Both checks must run on every normal production config load.
- [Blocker] The controller reload path is explicitly unsafe in the current design. `tryReloadConfig` uses `cityConfigIncludesWithBuiltinPacks` plus raw `config.LoadWithIncludes`, while `cityConfigIncludesWithBuiltinPacks` returns only include paths, cannot assert post-load provenance, and does not clearly share the same manifest-integrity path as `builtinPackIncludesForConfigLoad`.
- [Blocker] `LoadWithIncludes` deduplication by pack name can silently drop the required system Core when the root import closure reaches another pack named `core`. The design must make a required system-pack name collision with any user/imported `core` a hard composition error, or introduce a non-colliding identity mechanism with actionable diagnostics.
- [Major] The scanner/allowlist plan is not deliverable without a seeded call-site inventory. Reviewers found many direct production `config.Load` / `config.LoadWithIncludes` paths in `cmd/gc`, including `effectiveCityName`, `RawConfig`, drift/registration paths, convoy commands, init/readiness paths, and no-refresh completion/stop flows. Each must be classified as assertion-required, migrated, or a narrowly tested partial-read exception.
- [Major] Include-builders and config-returning wrappers are conflated. `cityConfigIncludesWithBuiltinPacks` and `builtinPackIncludesForConfigLoad` can enforce materialization/integrity, but they cannot perform post-load assertions. The assertion belongs in config-returning wrappers such as `loadCityConfig`, `loadCityConfigFS`, `loadCityConfigWithBuiltinPacks`, `loadConfigCommandCityConfig`, and `tryReloadConfig`.
- [Major] The warning-vs-fail policy for required Core is unresolved. Existing refresh paths may warn and proceed when prior content appears usable, while the design says production loads must fail loudly for corrupt/partial Core. The design must state the single rule for required Core and apply it uniformly.
- [Minor] Several edge contracts remain unspecified: path normalization for provenance comparison, whether extra unexpected files under `.gc/system/packs/core` are fatal or proven inert, whether zero-effective-contribution Core still appears in provenance, whether `usesOSFS` is only a test escape hatch, and whether doctor fix wording should say it repairs generated system packs rather than adding a user import.

**Disagreements:**
- Codex returned `approve-with-risks`, while Claude, DeepSeek, and Gemini returned `block`. My assessment is `block` because Codex's own required changes include the same unresolved identity model, production-loader inventory, wrapper placement, and required-Core fail policy that this persona is responsible for enforcing.
- Reviewers differed on how much existing `Provenance.Sources` can support the assertion. DeepSeek treated it as a real substrate; Claude, Codex, and Gemini emphasized that it is not enough by itself for content identity, name-collision safety, or orders/formulas-only packs. My assessment: `Sources` may be an implementation detail, but only after the design adds a typed required-system-pack participation/identity contract and focused tests.
- Reviewers used different severities for specific bypasses such as `effectiveCityName`, no-refresh completion/stop, `config.Load` raw reads, and `usesOSFS`. My assessment: they do not all need the same fix, but the design must inventory and classify them before slice 4 can be approved.
- Extra unexpected files under Core are not settled. Some reviews asked for full file-set rejection; others allowed proof that unexpected files cannot be consumed by loaders. My assessment: either policy can work, but the design must pick one and test it.

**Missing evidence:**
- A canonical required-system-pack provenance/identity data model, including pack name, materialized system path, content identity or manifest result, required/system source, and path-normalization rules.
- A complete non-test `cmd/gc` inventory of direct `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, aliases, and wrapper helpers, with per-function classification and focused tests.
- Tests for user/imported `core` collision, validated Core omitted from includes, non-system `core`, Core contributing zero effective keys, missing/corrupt/partial/stale Core, no-refresh callers, and production non-OSFS bypass prevention.
- A stated decision on required-Core refresh failures: hard-fail, or warn only for non-required packs after manifest revalidation.
- A stated decision on unexpected extra files under `.gc/system/packs/core`.
- Updated requirements/doctor wording that clarifies the fix repairs generated system pack materialization and the normal include path, not a `[imports.core]` user config entry.

**Required changes:**
- Define the required Core load contract as manifest-integrity fatality plus provenance-presence fatality. Treat provenance as defense-in-depth for dropped includes, not as the corruption/partial detector.
- Define required system Core identity and collision handling. A user/imported pack named `core` must not silently replace the validated system Core; make the collision fail closed with an actionable error or use a disambiguated identity that cannot collide.
- Move post-load Core assertions into config-returning wrappers and name `tryReloadConfig` as in-scope. Keep include-builders as helper APIs, and have the scanner reject include-builder plus raw `config.LoadWithIncludes` without assertion.
- Require a slice-4 call-site inventory and scanner allowlist before implementation. Include `config.Load` and `config.LoadCity`, not only `config.LoadWithIncludes`; allowlist entries must name file, function, loader symbol, reason, allowed fields if partial, and the focused test.
- Reconcile `materializeBuiltinPacksForConfigLoad`, `cityConfigIncludesWithBuiltinPacks`, and `builtinPackIncludesForConfigLoad` so required Core manifest-integrity failures are handled identically and loudly on production paths.
- Classify and test no-refresh paths (`loadCityConfigWithoutBuiltinPackRefresh*`, completion, stop), `effectiveCityName`, `RawConfig`, drift/registration reads, convoy loads, init/readiness loads, and the `usesOSFS` escape-hatch behavior.
- Specify path normalization, zero-effective-contribution provenance behavior, and unexpected-file policy for Core.
