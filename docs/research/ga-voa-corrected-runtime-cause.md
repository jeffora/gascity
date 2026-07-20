# ga-pwr CORRECTED: why the wired city→rig control sweep doesn't sweep at runtime

**Repo pin verified:** `git -C rigs/gascity cat-file -e 87c592692 && git -C rigs/gascity cat-file -e 06da50623` printed `REPO_OK`. All citations below are against commit **87c592692** in `/Users/jeff.burn/Code/gasland/rigs/gascity`. `/Users/jeff.burn/Code/outskirts_bak/gascity` was not touched.

## Verdict

**Confirmed: Hypothesis 1 — per-store env scoping.** `controlDispatcherSweepStorePaths` (`cmd/gc/dispatch_runtime.go:491`) correctly enumerates every rig store directory, and `drainWorkflowServeWork` correctly iterates each one — but it queries every directory in `sweepDirs` using **one `workEnv` value built once for the bare city dispatcher**, and that env pins the query to the **city's own beads store** via `BEADS_DIR`, independent of which directory is actually being iterated. The rig store directory is *walked* but never actually *queried* — the query silently re-reads the city store every time and finds nothing, because a rig-store finalize bead was never created there.

This is not hypothesis 2 (no error is thrown — the query succeeds, it's just answering the wrong store) and not hypothesis 3 (the guard and candidate enumeration are both correct; `sweepDirs` genuinely contains the rig store path).

## The exact break, traced end to end

1. **`workEnv` is computed exactly once per `--serve` invocation**, for the bare city-level control-dispatcher `agentCfg` (`Dir == ""`):
   `cmd/gc/dispatch_runtime.go:355`
   ```go
   workEnv, err := controllerWorkQueryEnv(cityPath, cfg, &agentCfg)
   ```
   That single `workEnv` value is then threaded unchanged into `drainWorkflowServeWork` (`:369`, `:372`) and, in `--follow` mode, `runWorkflowServeFollow` (`:598` → `:623`), which calls `drainWorkflowServeWork` again every poll cycle with the **same unchanged `workEnv`**.

2. **`controllerWorkQueryEnv` resolves rig-vs-city scope from `agentCfg`, not from the directory being queried** (`cmd/gc/work_query_probe.go:52-82`). Since the *bare* control dispatcher's `agentCfg.Dir == ""`, `configuredRigName(cityPath, agentCfg, cfg.Rigs)` short-circuits to `""` (`cmd/gc/cmd_start.go:1329-1334`: `if a == nil || a.Dir == "" { return "" }`), so `controllerQueryRuntimeEnv` (`work_query_probe.go:15-50`) takes the **city** branch and calls `bdRuntimeEnvWithError(cityPath)` (`work_query_probe.go:31,37`) — never `bdRuntimeEnvForRigWithError`.

3. **`bdRuntimeEnvWithErrorRecovery` hardcodes the city's `.beads` directory into the env it returns** (`cmd/gc/bd_env.go:1331-1333`):
   ```go
   env := cityRuntimeEnvMapForCity(cityPath)
   env["BEADS_DIR"] = filepath.Join(cityPath, ".beads")
   env["GC_RIG"] = ""
   env["GC_RIG_ROOT"] = ""
   ```
   The comment on the rig variant of this function makes the mechanism explicit (`bd_env.go:1262-1264`, on `bdRuntimeEnvForRigWithErrorRecovery`):
   > "Pin the rig store explicitly. The gc-beads-bd provider derives its Dolt data root from `GC_CITY_PATH` unless `BEADS_DIR` is set, so cwd-based discovery is not sufficient for rig-scoped operations."

   In other words: **`BEADS_DIR` (an env var), not the subprocess's working directory, is what the beads provider uses to pick its store.** `bdRuntimeEnvWithErrorRecovery` bakes in the *city's* `BEADS_DIR` — and that is the `workEnv` the sweep loop reuses for every store in `sweepDirs`.

4. **The sweep loop passes this city-pinned env to every directory, including the rig store dirs it enumerated:**
   `cmd/gc/dispatch_runtime.go:509-524`
   ```go
   sweepDirs := controlDispatcherSweepStorePaths(agentCfg, cityPath, storePath, cfg)
   for {
       ...
       for _, dir := range sweepDirs {
           dirQueue, err := workflowServeList(serveQuery, dir, workEnv)   // <-- same workEnv for every dir
   ```
   `workflowServeList` is a var-aliased call to `nextWorkflowServeBeads` (`dispatch_runtime.go:85`, defined at `:972-993`), which runs the query via `shellWorkQueryWithEnv(workQuery, dir, mergeRuntimeEnv(os.Environ(), env))` (`:976`). `shellWorkQueryWithEnv` (`cmd/gc/cmd_hook.go:555-586`) does set `cmd.Dir = dir` (the subprocess's cwd is correctly set to the rig store path), but that's cosmetic for this provider: the beads/`bd` CLI resolves its Dolt store from `BEADS_DIR` (an explicit env var) in preference to cwd-based discovery, and `BEADS_DIR` here is still `<cityPath>/.beads` from step 3 — the subprocess's cwd is ignored for store resolution.

**Net effect:** for every dir beyond the primary `storePath` — i.e. every rig store `controlDispatcherSweepStorePaths` adds — the query silently runs against the **city** store again (with the subprocess merely `cd`'d somewhere cosmetic). It returns whatever the city store has for that query (typically nothing, since the rig-store finalize bead was never created in the city's own beads store). No error occurs — `nextWorkflowServeBeads` returns cleanly with a (correct, for the store it *actually* asked) empty/mismatched result — so the `err != nil` / `workflowTracef("serve query-error ...")` branch (`:518-533`) never fires, and there is nothing to silently swallow. The rig-store finalize bead sits open forever with no trace of a failed query, because from the sweep's point of view it never queried the rig store's database at all.

## Why the empirical facts line up

- **On-demand `gc convoy control <finalize> --rig gascity` works**: that path resolves `agentCfg` with `Dir` set to the rig (or otherwise routes through `bdRuntimeEnvForRigWithError`, which explicitly sets `BEADS_DIR` to the *rig's* `.beads`), so it queries the correct store.
- **The `--serve --follow` bare city dispatcher never advances it**: it always runs with the *city* `workEnv`, computed once at loop start from the bare `agentCfg` (`Dir == ""`), and reuses it for the entire life of the loop across every sweep dir.
- **Config is correctly ruled out** (per the Mayor's prior verification): this is 100% compiled Go behavior — a single env map built once from the wrong scope and never re-derived per sweep directory.

## The fix

`controlDispatcherSweepStorePaths` and `drainWorkflowServeWork` treat `sweepDirs` as if querying a different directory is sufficient to query a different store. It is not, for this beads provider — the store is selected by `BEADS_DIR`/related env vars, not cwd. The fix must derive **a per-directory env**, not reuse one city-wide `workEnv` for every dir in the sweep.

Concretely, in `drainWorkflowServeWork` (`cmd/gc/dispatch_runtime.go:509-556`), the loop `for _, dir := range sweepDirs { ... workflowServeList(serveQuery, dir, workEnv) ... }` needs to resolve a store-scoped env per `dir`, the same way `bdRuntimeEnvForRigWithError` is already used elsewhere for rig-scoped operations. Concretely:

1. Change `drainWorkflowServeWork`'s signature (or add a helper) to accept `cfg *config.City` (already a parameter) and, for each `dir` in `sweepDirs` that is *not* the primary `storePath`, resolve the rig owning that dir (`rigConfigForScopeRoot(cityPath, dir, cfg.Rigs)` / the same lookup `convoyStoreCandidates` used to produce it) and call `bdRuntimeEnvForRigWithError(cityPath, cfg, dir)` (`cmd/gc/bd_env.go:1248`) to get a store-scoped env, instead of passing the single city-scoped `workEnv` through unchanged.
2. Keep `workEnv` (the city env) only for `dir == storePath` (the primary/bare city store), since that's the scope `controllerWorkQueryEnv` correctly built it for.
3. Thread that per-dir env into `workflowServeList(serveQuery, dir, perDirEnv)` at `dispatch_runtime.go:517` instead of the shared `workEnv`.

This mirrors what `controlDispatcherSweepStorePaths`'s own doc comment (`:475-489`) already says should happen ("each rig's control beads live in that rig's own store... without sweeping those stores too...") — the sweep-dir enumeration was fixed (ga-c7m Root A), but the per-dir *store resolution* for the query itself was not, and that's Root B / the actual runtime cause of ga-voa persisting after ga-c7m.

## Deliverable notes

- Research only — no code changes made.
- No real store was mutated; no finalize was drained. All investigation was static source reading against pinned commit `87c592692`, plus read-only `ls`/`cat` of `.beads/dolt-server.port` files to confirm the city and rig scopes run distinct managed-Dolt-backed stores (both currently resolve to the same host `dolt-server.port` value in this environment, `53577`, which is consistent with a single Dolt SQL server hosting multiple named databases — reinforcing that `BEADS_DIR`/database selection, not host:port, is the actual per-scope differentiator being clobbered here).
