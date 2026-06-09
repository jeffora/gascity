# Elias Sato - Claude

**Verdict:** block

I reviewed only my lane: required Core inclusion, config provenance, production
loader bypass containment, and loud failure on corrupt Core. The design's
*intent* here is excellent and unusually complete — the two fatal gates, the
resolver-produced participation record, and the read-only-doctor split close all
three of my red flags at the contract level. The block is narrow and concrete:
the one mechanism that actually catches "Core materialized but absent from
resolved config" is asserted but never designed against the resolver that
exists, the bypass-containment scanner is cited against a model that cannot
deliver it, and the design's only worked loader inventory misses a *live in-tree
behavior-driving path that already resolves config with no Core and swallows the
error*. All are fixable with bounded design edits, not a rethink.

**Top strengths:**
- **Fail-closed is table-driven and pre-behavior, not doctor-dependent.** The
  Attempt 14 failure matrix (`Resolver-Produced Required-Pack Provenance`,
  design.md ~L2083–2095) plus the two fatal gates (pre-resolution strict
  file-set integrity, then post-resolution typed participation) make
  missing/corrupt/partial/stale/shadowed/extra-file Core fail "before any
  behavior-bearing config, formula, order, script, prompt, or overlay can be
  read." Plain `gc doctor` is explicitly read-only (Attempt 14/17), so absence
  is caught by the loader itself — not "discovered later by doctor after
  behavior already degraded." This directly closes my worst red flag.
- **The participation proof bans every weak form up front.** Path-only,
  include-count, provenance-string, successful-materialization, and helper-name
  proof (`assertRequiredSystemPackProvenance`, ~L1555) are all explicitly
  rejected; a pack that materializes but produces no resolver participation
  record *for the same digest* is a load failure (~L2080–2081, ~L984–987). This
  is the correct shape for "materialized-but-unresolved = fatal."
- **The dev/test escape hatch is structural, not a runtime switch.** Production
  `gc` always includes Core; no-Core is reachable only by `_test.go` calling
  lower-level loaders, and `internal/config` tests are explicitly allowed to use
  `config.LoadWithIncludes` directly (~L3041–3048). `GC_BOOTSTRAP=skip` is
  retired as a production behavior switch (Attempt 9/10/17). This closes the
  "test-only no-Core escape hatch leaks into production" red flag at the design
  level.

**Critical risks:**

- **[Blocker] A live behavior-driving path already resolves config with no Core
  and swallows the error, and it is absent from every concrete inventory in the
  design.** `internal/dispatch/control.go:832 loadAttemptRouteConfig` calls
  `config.LoadWithIncludes(fsys.OSFS{}, filepath.Join(cityPath, "city.toml"))`
  with **no extra/builtin includes**, then `if err != nil { return nil }` —
  silently. It drives attempt-route binding from four sites (`control.go:424`,
  `fanout.go:299`, `ralph.go:358`, `ralph.go:626`). This is precisely the class
  my invariant exists to kill: a production path that (a) omits required Core and
  (b) degrades silently rather than failing loud. The design lists
  `internal/dispatch` in prose (~L990, ~L3031) but the *only* worked disposition
  table — the Attempt 4 "Required Core Loader Bypass Inventory" (~L468–501) — is
  `cmd/gc`-only and never assigns this site a loader class. A design whose
  enforcement inventory misses a known, in-tree instance of the exact violation
  is not yet safe to implement against.

- **[Major] The linchpin — resolver-produced, digest-keyed participation — is
  asserted but never designed against the resolver that exists.** Attempt 14
  requires participation "produced by the resolver, not inferred by a caller,"
  with deterministic layer/import-edge ids "derived from normalized source
  identity, content digest, layer ordinal, and import parent id" (~L2058–2081).
  The real resolver (`internal/config.LoadWithIncludes` →
  `LoadWithIncludesOptions`, compose.go:108–113) returns `(*City, *Provenance,
  error)`, and `Provenance` (compose.go:72) keys imports by **binding name →
  source string** with an `"(implicit)"` sentinel — i.e. path/string labels, not
  content digests or stable edge ids. So the proof the design demands does not
  exist today, and the design forbids the only proof the resolver currently
  offers (provenance strings). The "System Pack Loading" section frames
  everything as `internal/systempacks` wrappers and never states how
  `internal/config` is extended to emit digest-backed import-edge participation,
  nor where the `RequiredSystemPackParticipation` type lives so the resolver can
  produce it without `internal/config` importing `internal/systempacks` (a
  layering inversion). Constructively: `Provenance` already carries `Sources` in
  load order and `sourceContents map[string][]byte` (raw bytes), so the digest
  binding can be built *on top of* the existing provenance rather than as a
  parallel mechanism — but the design must say so. Left to implementation, this
  silently degrades back to provenance-string matching, which the design itself
  rejects.

- **[Major] The named enforcement model cannot carry the stated requirement.**
  The bypass scanner is "modeled on `TestGCNonTestFilesStayOnWorkerBoundary`"
  (~L3030). That test (verified, `cmd/gc/worker_boundary_import_test.go`) is a
  `strings.Contains` scan over a **single, non-recursive directory** with no
  alias resolution. Attempt 17 (~L2411–2419) requires a generator that scans all
  production `internal/`, "follows aliases and wrappers, and emits a row for
  every call that can reach `config.Load`." These are incompatible. Direct-call
  detection is feasible — the whole-production surface is small (~37 non-test
  `config.Load*` sites in `cmd/gc`, ~8 in `internal/`) — but a substring test
  cannot follow the real wrappers `loadCityConfig`,
  `loadCityConfigWithoutBuiltinPackRefresh(FS)` (cmd/gc/cmd_agent.go:32–84,
  used by completion.go), or `loadCityConfigWithBuiltinPacks`
  (cmd/gc/cmd_config.go:21), nor reach through a helper package that wraps
  `config.Load`. The "every call that can reach config.Load" guarantee needs an
  AST/call-graph analyzer (e.g. `golang.org/x/tools/go/callgraph`) with a stated
  false-negative story for interface dispatch — or the contract must narrow to
  direct-reference detection plus a curated wrapper allowlist. As written, an
  aliased/wrapped bypass defeats the guard — my "loader bypass" red flag.

- **[Minor] The scanner contract names a loader symbol that does not exist.**
  `config.LoadCity` is referenced in the bypass/scanner contract repeatedly
  (~L499, ~L692, ~L1233, ~L3027, ~L2415) but there is no `func LoadCity` and no
  `config.LoadCity` call site anywhere in the tree (verified). The real
  entrypoints are `config.Load`, `config.LoadWithIncludes`, and
  `config.LoadWithIncludesOptions`. Harmless to forbid a phantom, but it is
  evidence the inventory was authored from memory rather than generated from
  source — which is in direct tension with the "generated, default-deny,
  source-derived" guarantee that the whole containment story rests on.

- **[Minor] A second production `GC_BOOTSTRAP` consumer is unaddressed in the
  cleanup prose.** Bootstrap Cleanup (~L3151–3195) treats only
  `internal/bootstrap/bootstrap.go:72`. There is a second production consumer,
  `internal/doctor/implicit_import_cache_check.go:236–245`, which
  unsets/restores `GC_BOOTSTRAP` around its check. The Attempt 17 symbol guard
  would scan the token, but the concrete retirement prose should name this site
  so retirement does not silently alter that doctor check.

- **[Minor] Partial-read allowlist entries are reusable foot-guns.** Each
  allowlisted partial loader legitimately omits Core; its "focused test proving
  it cannot drive runtime behavior" is point-in-time. Nothing structural stops a
  *future* behavior-driving caller from invoking an already-allowlisted partial
  helper (Core silently absent again), and a substring scanner cannot catch
  reuse of an allowlisted symbol. The design should state how reuse of partial
  helpers on behavior paths is prevented (call-graph, package placement, or a
  review gate), not only that each row has a one-time test.

**Missing evidence:**
- The `internal/config` resolver API change that produces
  `RequiredSystemPackParticipation`: return shape, owning package, how
  layer/edge ids and the file-set digest are derived, and a worked example
  showing a *shadowed/overridden-to-empty* required pack is reported as
  non-participating rather than passing. Today's `*Provenance` cannot express
  this, and the design never reconciles the two.
- A concrete loader disposition (runtime / no-refresh / partial-edit /
  partial-doctor) for the behavior-driving `internal/` sites that exist now —
  `internal/dispatch/control.go:loadAttemptRouteConfig`, plus the
  `internal/doctor` and `internal/configedit` callers — rather than deferring
  entirely to the not-yet-existing `loader-inventory.generated.yaml`.
- Proof that the deny-by-default scanner's reachability/alias claim is
  achievable with the cited model, or an explicit switch to an AST/call-graph
  scanner with the substring worker-boundary test demoted to a backstop.

**Required changes:**
- Add a "Resolver Participation API" subsection to **System Pack Loading**
  specifying how `internal/config` emits digest-keyed import-edge participation
  out of resolution (extending or wrapping the existing `*Provenance`, which
  already holds `Sources` and `sourceContents`), which package owns the
  `RequiredSystemPackParticipation` type so there is no `config →
  systempacks` import inversion, the blast radius on existing `LoadWithIncludes`
  / `*Provenance` consumers (e.g. `gc config explain`, import-state doctor), and
  a test asserting a materialized-but-shadowed required pack fails
  post-resolution.
- Add the behavior-driving `internal/` call sites to a concrete disposition
  inventory — at minimum `internal/dispatch/control.go:loadAttemptRouteConfig`
  — with their target loader class, and explicitly note that this function
  currently resolves attempt-route config with no required-pack includes *and
  swallows the load error*; both must change under the invariant (omitting Core
  and silent failure are each violations).
- Split loader-bypass enforcement into two named mechanisms: (a) deny direct
  `config.Load*` outside `internal/systempacks` + generated allowlist
  (substring/AST-feasible), and (b) prevent allowlisted *partial* helpers and
  wrapper functions (`loadCityConfig*`) from being called on behavior paths
  (call-graph or review gate). Stop citing the single-directory substring
  worker-boundary test as the model for the "whole-production, alias-following,
  reachability" requirement.
- Correct the generated loader-inventory symbol set: drop the non-existent
  `config.LoadCity`; enumerate the real entrypoints (`config.Load`,
  `config.LoadWithIncludes`, `config.LoadWithIncludesOptions`); and add a
  freshness test asserting the inventory references only loader symbols that
  actually exist in the tree.
- Name the `internal/doctor/implicit_import_cache_check.go` `GC_BOOTSTRAP`
  consumer in the Bootstrap Cleanup section so its retirement is explicit.

**Questions:**
- Can the existing resolver distinguish "required pack included and survived the
  merge" from "required pack included but its layer was overridden/emptied by a
  later layer"? If not, post-resolution participation needs resolver-internal
  work beyond a `systempacks` wrapper — is `internal/config` resolver change in
  scope for the Core loading/doctor slice (slice 4), and is its cost reflected in
  slicing?
- For `LoadRuntimeCityNoRefresh`, what is the authoritative comparison that
  proves the *already-materialized* on-disk Core matches the embedded manifest
  digest (vs. only checking presence)? The no-refresh class is the path most
  likely to accept stale-but-present Core, and the controller may serve a
  long-lived in-memory config across an on-disk Core corruption between reloads.
- Does any production consumer of `*Provenance` depend on the current
  binding-name/string-label shape in a way that constrains how digest-keyed
  participation is threaded out of the resolver?
