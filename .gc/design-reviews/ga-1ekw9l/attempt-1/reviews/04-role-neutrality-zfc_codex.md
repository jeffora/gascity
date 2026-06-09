# Anand Krishnaswamy

**Verdict:** approve-with-risks

**Top strengths:**
- The plan treats role neutrality as an enforcement surface, not prose: it covers Go, TOML, shell, prompts, generated command text, API/dashboard references, tmux helpers, defaults, metadata, tests, docs, and public Gastown companion files.
- The active-root rule is the right architectural boundary. Forcing loaders, expanders, prompt scanners, script resolvers, hook readers, and doctor checks through `internal/packsource.ActiveRootsFor(kind)` directly attacks the current class of glob/string-prefix leaks.
- The binding model is directionally ZFC-safe: city and system-pack data provide bindings, formula/step metadata references symbolic bindings, missing optional bindings produce diagnostics, and Go may not substitute `mayor`, `deacon`, `dog`, or any concrete fallback.

**Critical risks:**
- [Major] The binding resolver is specified as behavior, but not as auditable evidence. The plan says `target_binding` and `gc.run_target_binding` resolve through a merged binding table, and tests cover representative bindings, but it does not require a machine-readable resolution record for every route/mail/nudge/warmup/formula consumer. Given the current tree has many literal `gc.routed_to`, `pool:dog`, `mol-dog-*`, and Gastown role strings in examples, docs, tests, and formulas, token scanning alone will not prove that active behavior came through the generic resolver rather than through string interpolation or a helper-specific fallback.
- [Major] The `dog` exception is still too broad for a zero-hardcoded-roles gate. The design allows `dog` as Core's default configurable maintenance worker in pack configuration and tests, but does not pin that exception to exact config keys, exact rendered-output expectations, and a negative test matrix proving that the same literal fails if it appears in a route target, prompt fallback, formula default, mail recipient, warmup default, generated help, or Go branch.
- [Major] Retired `maintenance` identity is enforced mainly through pack-source/docs scanners, not the role-neutrality contract itself. The persona risk is not lowercase maintenance prose; it is `maintenance` as a pack id, import source, route/default namespace, system-pack path, or generated behavior owner. The role-surface denied-token model should include retired standalone pack identifiers with context-sensitive matching, otherwise active Core can be role-neutral while still depending on a retired Maintenance identity.
- [Minor] The review artifact/docs allowlist language needs a hard anti-wildcard rule. Requirements allow generated review artifacts and migration docs as exceptions; without per-file, per-token, per-context rows tied to inactive roots, that can become the broad escape hatch that lets compatibility fixtures leak into generated or materialized behavior.

**Missing evidence:**
- A `BindingResolution` or equivalent trace schema that records consumer id, consumer kind, source pack, binding key, required/optional mode, source of the selected binding, resolved target, absence behavior, diagnostic id, and whether the consumer is behavior-driving.
- A role-surface manifest example showing exact allowed rows for the Core `dog` default and proving the exception is limited to data keys rather than any Core asset or generated output.
- A dual-mode scan contract: Core-only city/rendered outputs must be free of Gastown role identifiers, while a city that explicitly imports public Gastown may contain Gastown-owned role configuration without teaching the SDK those roles.
- Evidence that `maintenance` as a retired pack/source/route identifier is banned from active Core paths, separate from allowed operator prose about maintenance work.

**Required changes:**
- Add a generic binding-resolution evidence contract and require every behavior-driving target field to flow through it: `gc.routed_to`, formula target/run target, mail recipient, nudge target, warmup defaults, prompt fallbacks, notification targets, and any route metadata. Scanner tests should reject direct writes or template interpolation of those fields unless the allowlist row proves the surface is inactive historical/test data.
- Narrow the `dog` exception to exact configuration paths and exact tests. Each exception row needs owner, reason, expiry or explicit permanent-design rationale, token kind, rendered-output expectation, and negative fixtures proving the token cannot move into active routing, notification, prompt, formula, generated help, API/dashboard, or Go code.
- Extend `role-neutrality-scan.yaml` with context-sensitive retired-identity tokens for `maintenance` and legacy Maintenance/Gastown source paths. Match them as pack ids, import keys, route namespaces, generated owner names, and path components while preserving legitimate lowercase maintenance terminology through the docs terminology matrix.
- Make the Core-only versus public-Gastown scan modes explicit in the plan and gates. Public Gastown role strings should pass only when owned by the pinned public Gastown pack; the same strings in Core, provider packs, generated SDK defaults, stale materialized system packs, or Go routes should fail.

**Questions:**
- Is the Core `dog` default intended as a permanent shipped default, or a compatibility bridge that should expire once public Gastown activation is complete?
- Will `internal/systempacks` or the config result expose binding-resolution provenance, or will that live only in test manifests?
- Are public Gastown companion scans run against the exact pinned external commit/cache entry, or against a local checkout that could drift from the consumed pin?
