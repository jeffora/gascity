# Natasha Volkov - Behavior Parity Guardian (Codex)

**Persona verdict:** approve-with-risks

This revision is acceptable for decomposition only if the first executable
work remains limited to non-mutating Slice 0. It should not approve API target
classification, wake, close, reconciler, event, or runtime behavior movement
yet. The design now makes the right product-safety move: it treats
`REQUIREMENTS.md` as the behavior source, blocks later slices on a closed Slice
0 evidence bead, and requires exact current-behavior proof before a caller is
delegated.

## What changed enough to pass this persona

- The design no longer lets later implementation slices treat the requirements
  ledger as fresh by assertion. `internal/session/DESIGN.md:173-179` explicitly
  requires stale or missing evidence, including `SESSION-RECON-002`,
  `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`, to be
  repaired or owner-retired before a dependent slice cites it. Those rows still
  cite absent proof paths in `internal/session/REQUIREMENTS.md:129-134`, so this
  gate is necessary and correctly scoped.
- Slice 0 is now non-mutating by contract. `internal/session/DESIGN.md:139-141`
  says it may not move callers, change target resolution, materialize sessions,
  repair metadata, add event payloads, or change reconciler policy. That makes
  it safe to decompose Slice 0 even while behavior proof is incomplete.
- The Slice 0 artifact set is broad enough to prevent proof drift if enforced:
  `SCENARIO_PARITY.yaml`, route inventory, worker-boundary exceptions,
  diagnostics, boundary matrix, vocabulary checkpoints, and mutation symbols are
  all required at `internal/session/DESIGN.md:150-165`, with machine-readable
  row metadata at `internal/session/DESIGN.md:167-170`.
- The target-classification section now names the actual API resolver
  precedence instead of inventing a new generic taxonomy. The current chain in
  `internal/api/session_resolution.go:433-470` matches the design's order in
  `internal/session/DESIGN.md:223-240`: reject `template:`, exact ID,
  configured named session, ordinary live session name/alias, path alias, then
  allow-closed lookup with configured named targets rejected first.
- The `RepairEmptyType` side effect is finally treated as a parity-sensitive
  behavior change. Current helpers still mutate as a read side effect
  (`internal/session/resolve.go:55-57`, `internal/session/resolve.go:142`,
  `internal/session/resolve.go:222-228`); the design requires the read-only
  classifier to return `repair-needed` and prove either separate repair parity
  or owner-approved behavior change at `internal/session/DESIGN.md:246-251`.
- Characterization is framed at the user-visible surface, not just helper
  internals. `internal/session/DESIGN.md:279-282` requires positive/negative
  precedence fixtures plus legacy mux/Huma wire parity, and
  `internal/session/DESIGN.md:640-652` requires caller behavior tests before
  duplicated logic is deleted.

## Required Slice 0 close gates

These are not blockers to approving the design shape, but they must be hard
close conditions for the Slice 0 bead.

1. **The proof command must not pass with zero matching validators.** I ran the
   published minimum proof command from `internal/session/DESIGN.md:189-193` in
   this checkout. It returned success while `cmd/gc`, `internal/session`, and
   `internal/events` reported `[no tests to run]`; `rg` found only
   `TestEveryKnownEventTypeHasRegisteredPayload` under the named validators.
   Slice 0 must add a bootstrap validator or equivalent `go test -list` check
   that fails if any expected validator symbol is absent, skipped, build-tagged
   out, or matched by zero tests.
2. **`SCENARIO_PARITY.yaml` must separate executable behavior proof from static
   source inventory.** The design allows "exact tests or static selectors" for
   parity rows at `internal/session/DESIGN.md:158`, but active behavior rows
   cannot be green because a source path exists. The stronger rule at
   `internal/session/DESIGN.md:177-179` should win: every active `SESSION-*`
   parity row needs an executable selector, or an owner-approved retirement or
   amendment record. Static selectors are supporting evidence only.
3. **Owner approval needs validator-checkable identity.** The product rule at
   `internal/session/DESIGN.md:48-50` is directionally correct, but Slice 0
   should make the approval artifact schema include changed row IDs, approval
   authority, artifact ID, amendment state, expiry if temporary, and a row
   content hash or equivalent changed-text fingerprint. Otherwise a later edit
   can silently drift away from the approved behavior decision.
4. **The first-adopter route set must be exact, not "related query endpoints."**
   `internal/session/DESIGN.md:203-208` is safe only if Slice 1 names the exact
   legacy handlers, Huma operation IDs, generated-client methods, and tests it
   will move. Session stream resolution, for example, is query-side resolution
   but immediately calls worker `History` and `State`
   (`internal/api/handler_session_stream.go:74-108`). That can still be a valid
   first adopter, but only with route-specific wire and no-mutation proof.
5. **`RepairEmptyType` parity must be explicitly decided for each first-adopter
   endpoint.** Returning `repair-needed` instead of repairing during lookup is a
   behavior change unless an adapter performs an audited repair before rendering
   the same user-visible result. Slice 1 must prove the chosen result for exact
   ID, session-name, alias, path alias, transcript, pending, and stream query
   paths.

## Residual risks

- The design still carries future-facing contracts for commands, events,
  diagnostics, runtime intents, and migration coexistence. It mitigates the
  risk by marking them non-authoritative until a slice uses them, but reviewers
  should reject any downstream bead that cites those sections without adding
  operation-specific preflight rows.
- The requirements ledger still contains stale reconciler proof references.
  That is acceptable for Slice 0 only because Slice 0 is tasked with repairing
  or retiring them. It is not acceptable evidence for any behavior-moving
  reconciler, pool, health, or progress slice.
- API query classification is the right first behavior slice, but it is close
  to mutating surfaces that share resolver helpers. The route inventory and
  static guard need to prove mutating API commands, CLI fallback paths, mail,
  extmsg, assignee normalization, nudge, attach, and pool resume remain out of
  scope until their own compatibility matrices pass.

## Required changes before behavior-moving slices

- Close Slice 0 with machine-readable artifacts and self-validating tests that
  fail on missing validators, zero-match selectors, stale paths, static-only
  parity, missing negative fixtures, and missing owner approval metadata.
- Restore, replace, or owner-retire missing proof for `SESSION-RECON-002`,
  `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`.
- Make every active `SESSION-*` row trace to one current oracle, one proof
  command, exact selectors, touched surfaces, and amendment state before any
  slice may cite it.
- For Slice 1, enumerate exact API query routes and operation IDs, including
  legacy and Huma variants, and require unchanged wire behavior for status,
  problem bodies, generated-client shapes, closed lookup, config-orphan
  rejection, path aliases, and `RepairEmptyType` handling.

## Bottom line

Approve the design for Slice 0 decomposition. Do not approve any
behavior-moving slice until Slice 0 proves the ledger and test selectors are
fresh, executable, and validator-enforced.
