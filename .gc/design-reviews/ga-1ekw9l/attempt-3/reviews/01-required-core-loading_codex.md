# Camille Sato - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The plan names a single required-host-pack boundary, `internal/systempacks`, and gives it the right ownership surface: materialization, file-set validation, runtime includes, and runtime city loading (`design-before.md:171`-`design-before.md:179`).
- The loader contract is fail-closed in the two places this persona cares about: pre-resolution file-set validation and post-resolution typed `RequiredSystemPackParticipation` validation (`design-before.md:187`-`design-before.md:192`).
- The no-refresh path explicitly avoids repair, preserves last-known-good config only for read-only status/reporting, and refuses behavior-changing dispatch, formula, order, hook, prompt, and agent-start operations until a refreshed load succeeds (`design-before.md:194`-`design-before.md:200`).

**Critical risks:**
- [Major] The fail-closed behavior is specified, but the loader API shape is still underdefined. `LoadRuntimeCity` and `LoadRuntimeCityNoRefresh` are named, yet the plan does not state what typed result they return when Core is invalid and last-known-good config is still available. If they return an ordinary `config.City`, behavior-changing callers can accidentally use stale config despite the prose requirement to pause/refuse work. The plan should require a typed runtime-load result that carries freshness, diagnostic, participation records, and allowed-use mode, or an equivalent guard that behavior-changing operations must consume.
- [Minor] Provider-conditioned required packs are described, but `RequiredPackNames(cityPath string)` does not show how `bd`/`dolt` participation is selected without duplicating provider detection outside the boundary. The surrounding APIs take `provider`, so this may be only a signature sketch, but the final plan should keep provider selection inside `internal/systempacks` and avoid a second caller-side include decision.
- [Minor] The bypass scanner is strong, including aliases, wrappers, function values, and manual include assembly, but the allowlist process is still abstract. The plan should require an initial generated inventory of existing `config.Load*` call sites with disposition before decomposition, so the scanner cannot be added after large migrations have already created informal exceptions.

**Missing evidence:**
- A typed return/status contract for `LoadRuntimeCity` and `LoadRuntimeCityNoRefresh`, including how stale last-known-good config is represented and how behavior-changing callers are blocked.
- An initial direct-loader inventory naming current `cmd/gc`, API, and behavior-driving `internal/` call sites, with which ones migrate, which ones are partial-read exceptions, and which tests prove exceptions cannot drive runtime behavior.
- A concrete participation proof example showing how Core, `bd`, and `dolt` records point back to resolved config edges rather than path/digest coincidence.

**Required changes:**
- Define the runtime loader result type or equivalent guard. It should make invalid/stale config distinguishable from fresh validated config at compile time or through a mandatory mode check.
- State that dispatch, formula, order, hook, prompt, agent-start, API mutation, and controller paths must reject stale/invalid loader results before reading behavior.
- Keep provider-conditioned pack selection inside `internal/systempacks`; adjust the `RequiredPackNames` signature or document that it derives provider requirements internally from city config.
- Add an implementation-planning artifact or test fixture that inventories current `config.Load*` bypasses and seeds the generated allowlist with owner, reason, call kind, fields consumed, and proof test.

**Questions:**
- What exact Go type crosses the `internal/systempacks` boundary: a plain `config.City`, or a wrapper that carries participation records, diagnostics, and behavior-readiness?
- Does `LoadRuntimeCityNoRefresh` ever return last-known-good config to callers that might start agents or dispatch formulas, or is there a separate read-only accessor for status/reporting?
- Will the scanner run before the migration lands to catch existing wrappers and function-value aliases, or only after the new loader has been adopted?
