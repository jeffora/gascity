# 06 Pack Boundary Containment - Codex Review

Persona: Owen Gallagher

Verdict: approve with one required rollout fix. The design has the right
boundary model, but the rollout plan must explicitly place the retired-source
classifier before any slice that can consume a public pin, move behavior, remove
Maintenance, or enumerate active pack behavior.

## Findings

### Required: assign `internal/packsource` to an early slice and gate every active enumerator through it

The proposed boundary contract is strong in the design text:
`internal/packsource` is the sole authority for retired Maintenance and in-tree
Gastown source classification, and load/install/cache/lockfile/materialization/
discovery/doctor/docs-lint/generated-reference paths must use it instead of
duplicating string checks (`implementation-plan.md:320`). The plan also adds an
active behavior enumeration rule that loaders, installers, cache readers,
docs/reference generators, prompt scanners, formula/order expanders, script
resolvers, hook overlay readers, and doctor checks obtain roots through
`internal/packsource.ActiveRootsFor(kind)` or an equivalent typed API, with
scanner tests rejecting raw walks/globs over pack roots (`implementation-plan.md:418`).

That API is not concretely placed in the rollout slices. Slices 2 through 7
cover public pin adoption, Core extraction, `internal/systempacks`, doctorfix,
activation, Maintenance fold, registry/cache cleanup, and source deletion, but
none names the first slice that lands `internal/packsource`, ports active
enumerators to it, and turns on the raw-enumeration scanner
(`implementation-plan.md:739`, `implementation-plan.md:747`,
`implementation-plan.md:752`, `implementation-plan.md:772`,
`implementation-plan.md:780`, `implementation-plan.md:787`). The AC17 table
also points AC5/source-consumer closure at Slice 5b and AC8 role scanning at
Slice 3, but it does not make the active-root classifier itself a prerequisite
for Slice 2 public-pin adoption, Slice 3 Core extraction, or Slice 4a runtime
loading (`implementation-plan.md:710`, `implementation-plan.md:713`).

This leaves a containment gap: implementers can update pins, move Core assets,
or add runtime loaders while existing discovery, prompt/formula/order/script,
hook, cache, docs, or doctor paths still glob retired directories directly. The
design already says stale `.gc/system/packs/maintenance`,
`.gc/system/packs/gastown`, and `.gc/runtime/packs/maintenance` must remain on
disk but inactive (`implementation-plan.md:353`, `implementation-plan.md:598`);
that is only enforceable if the classifier is active before any behavior source
changes.

Required change: add an explicit slice, or amend Slice 2/3, that lands
`internal/packsource`, migrates all active pack-root enumeration through its
typed API, adds scanner allowlists for historical/non-behavior audits, and runs
negative tests with stale Maintenance/Gastown directories present. Make that
slice a prerequisite for public-pin adoption, Core extraction, runtime loader
cutover, Maintenance fold, registry/cache cleanup, and source deletion.

## What Is Solid

- The design prevents duplicate active behavior in principle. The
  zero-duplicate-active and zero-merge gates compare bundled, public, stale
  generated, synthetic cache, ordinary remote cache, compatibility-pin,
  activation-pin, and old/new binary views, and fail runtime loading if the same
  behavior id is active from more than one source (`implementation-plan.md:347`).
- The stale-state policy is correct. Startup and doctor must not delete stale
  generated Maintenance/Gastown directories automatically; they are ignored by
  active discovery and reported as legacy state (`implementation-plan.md:353`,
  `implementation-plan.md:598`). That preserves operator edits while preventing
  hidden fallback.
- The ownership split is specific enough for task planning. Core gets
  SDK-generic behavior; public Gastown gets role-specific formulas, prompts,
  detector/requester examples, branch pruning/Polecat behavior where applicable,
  and Gastown-specific notification policy; Maintenance is not recreated as a
  system pack (`implementation-plan.md:89`, `implementation-plan.md:96`,
  `implementation-plan.md:448`).

## Residual Risk

The active-root scanner needs positive and negative controls, not just a token
search. It should fail on `fs.WalkDir`, `filepath.Walk`, `Glob`, hand-built
cache/system-pack paths, and helper wrappers that enumerate active prompts,
formulas, orders, scripts, hook overlays, generated docs, or doctor inputs
without going through `internal/packsource`. Historical fixture and migration
audit paths can be allowlisted, but the allowlist must prove they cannot drive
runtime behavior.
